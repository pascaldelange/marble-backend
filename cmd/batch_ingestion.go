package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/checkmarble/marble-backend/infra"
	"github.com/checkmarble/marble-backend/jobs"
	"github.com/checkmarble/marble-backend/repositories"
	"github.com/checkmarble/marble-backend/usecases"
	"github.com/checkmarble/marble-backend/utils"
	"github.com/getsentry/sentry-go"
)

func RunBatchIngestion() error {
	// This is where we read the environment variables and set up the configuration for the application.
	gcpConfig := infra.GcpConfig{
		EnableTracing:      utils.GetEnv("ENABLE_GCP_TRACING", false),
		ProjectId:          utils.GetEnv("GOOGLE_CLOUD_PROJECT", ""),
		GcsIngestionBucket: utils.GetRequiredEnv[string]("GCS_INGESTION_BUCKET"),
		FakeGcsRepository:  utils.GetEnv("FAKE_GCS", false),
	}
	pgConfig := infra.PgConfig{
		Database:            "marble",
		DbConnectWithSocket: utils.GetEnv("PG_CONNECT_WITH_SOCKET", false),
		Hostname:            utils.GetRequiredEnv[string]("PG_HOSTNAME"),
		Password:            utils.GetRequiredEnv[string]("PG_PASSWORD"),
		Port:                utils.GetEnv("PG_PORT", "5432"),
		User:                utils.GetRequiredEnv[string]("PG_USER"),
	}
	jobConfig := struct {
		env           string
		appName       string
		loggingFormat string
		sentryDsn     string
	}{
		env:           utils.GetEnv("ENV", "development"),
		appName:       "marble-backend",
		loggingFormat: utils.GetEnv("LOGGING_FORMAT", "text"),
		sentryDsn:     utils.GetEnv("SENTRY_DSN", ""),
	}

	logger := utils.NewLogger(jobConfig.loggingFormat)
	ctx := utils.StoreLoggerInContext(context.Background(), logger)

	infra.SetupSentry(jobConfig.sentryDsn, jobConfig.env)
	defer sentry.Flush(3 * time.Second)

	tracingConfig := infra.TelemetryConfiguration{
		ApplicationName: jobConfig.appName,
		Enabled:         gcpConfig.EnableTracing,
		ProjectID:       gcpConfig.ProjectId,
	}
	telemetryRessources, err := infra.InitTelemetry(tracingConfig)
	if err != nil {
		utils.LogAndReportSentryError(ctx, err)
		return err
	}
	ctx = utils.StoreOpenTelemetryTracerInContext(ctx, telemetryRessources.Tracer)

	pool, err := infra.NewPostgresConnectionPool(ctx, pgConfig.GetConnectionString())
	if err != nil {
		utils.LogAndReportSentryError(ctx, err)
		return err
	}

	repositories := repositories.NewRepositories(nil, pool, nil, "")
	uc := usecases.NewUsecases(repositories,
		usecases.WithGcsIngestionBucket(gcpConfig.GcsIngestionBucket),
		usecases.WithFakeGcsRepository(gcpConfig.FakeGcsRepository),
	)

	err = jobs.IngestDataFromCsv(ctx, uc)
	if err != nil {
		logger.ErrorContext(ctx, "failed to ingest data from csvs", slog.String("error", err.Error()))
	}

	return err
}
