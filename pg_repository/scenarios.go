package pg_repository

import (
	"context"
	"errors"
	"fmt"
	"marble/marble-backend/app"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type dbScenario struct {
	ID                string      `db:"id"`
	OrgID             string      `db:"org_id"`
	Name              string      `db:"name"`
	Description       string      `db:"description"`
	TriggerObjectType string      `db:"trigger_object_type"`
	CreatedAt         time.Time   `db:"created_at"`
	LiveVersionID     pgtype.Text `db:"live_scenario_iteration_id"`
}

func (s *dbScenario) dto() app.Scenario {
	return app.Scenario{
		ID:                s.ID,
		Name:              s.Name,
		Description:       s.Description,
		TriggerObjectType: s.TriggerObjectType,
		CreatedAt:         s.CreatedAt,
	}
}

func (r *PGRepository) GetScenarios(ctx context.Context, orgID string) ([]app.Scenario, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("scenarios").
		Where(squirrel.Eq{
			"org_id": orgID,
		}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to build scenario query: %w", err)
	}

	rows, _ := r.db.Query(ctx, sql, args...)
	scenarios, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbScenario])

	scenarioDTOs := make([]app.Scenario, len(scenarios))
	for i, scenario := range scenarios {
		scenarioDTOs[i] = scenario.dto()
	}
	return scenarioDTOs, err
}

func (r *PGRepository) GetScenario(ctx context.Context, orgID string, scenarioID string) (app.Scenario, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("scenarios").
		Where(squirrel.Eq{
			"org_id": orgID,
			"id":     scenarioID,
		}).ToSql()

	if err != nil {
		return app.Scenario{}, fmt.Errorf("unable to build scenario query: %w", err)
	}

	rows, _ := r.db.Query(ctx, sql, args...)
	scenario, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbScenario])
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Scenario{}, app.ErrNotFoundInRepository
	} else if err != nil {
		return app.Scenario{}, fmt.Errorf("unable to get scenario: %w", err)
	}

	scenarioDTO := scenario.dto()

	if scenario.LiveVersionID.Valid {
		liveScenarioIteration, err := r.GetScenarioIteration(ctx, orgID, scenario.LiveVersionID.String)
		if err != nil {
			return app.Scenario{}, fmt.Errorf("unable to get live scenario iteration: %w", err)
		}
		scenarioDTO.LiveVersion = &liveScenarioIteration
	}

	return scenarioDTO, err
}

func (r *PGRepository) PostScenario(ctx context.Context, orgID string, scenario app.Scenario) (app.Scenario, error) {
	sql, args, err := r.queryBuilder.
		Insert("scenarios").
		Columns(
			"org_id",
			"name",
			"description",
			"trigger_object_type",
		).
		Values(
			orgID,
			scenario.Name,
			scenario.Description,
			scenario.TriggerObjectType,
		).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return app.Scenario{}, fmt.Errorf("unable to build scenario query: %w", err)
	}

	rows, _ := r.db.Query(ctx, sql, args...)
	createdScenario, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbScenario])
	if err != nil {
		return app.Scenario{}, fmt.Errorf("unable to create scenario: %w", err)
	}

	return createdScenario.dto(), err
}

// TODO: work on this, the method can be used to link any iteration with a scenario (no check on relationship)
// Idea: only pass a scenarioIterationID and get the scenarioID from it ?
func (r *PGRepository) PublishScenarioIteration(ctx context.Context, orgID string, scenarioID string, scenarioIterationID string) error {
	sql, args, err := r.queryBuilder.
		Update("scenarios").
		Set("live_scenario_iteration_id", scenarioIterationID).
		Where("id = ?", scenarioID).
		Where("org_id = ?", orgID).
		ToSql()

	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to run query: %w", err)
	}

	return nil
}

// TODO: do we want to unpublish a scenario form scenarioID ? or from scenarioIterationID ?
// Comment is here to think of both publication signature together
func (r *PGRepository) UnpublishScenarioIteration(ctx context.Context, orgID string, scenarioID string) error {
	sql, args, err := r.queryBuilder.
		Update("scenarios").
		Set("live_scenario_iteration_id", nil).
		Where("id = ?", scenarioID).
		Where("org_id = ?", orgID).
		ToSql()

	if err != nil {
		return fmt.Errorf("unable to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to run query: %w", err)
	}

	return nil
}
