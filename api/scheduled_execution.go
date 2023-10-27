package api

import (
	"archive/zip"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ggicci/httpin"

	"github.com/checkmarble/marble-backend/dto"
	"github.com/checkmarble/marble-backend/models"
	"github.com/checkmarble/marble-backend/utils"
)

func (api *API) handleGetScheduledExecution() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheduledExecutionID, err := requiredUuidUrlParam(r, "scheduledExecutionID")
		if presentError(w, r, err) {
			return
		}

		usecase := api.UsecasesWithCreds(r).NewScheduledExecutionUsecase()
		execution, err := usecase.GetScheduledExecution(scheduledExecutionID)

		if presentError(w, r, err) {
			return
		}

		PresentModelWithName(w, "scheduled_execution", dto.AdaptScheduledExecutionDto(execution))
	}
}

func (api *API) handleGetScheduledExecutionDecisions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		scheduledExecutionID, err := requiredUuidUrlParam(r, "scheduledExecutionID")
		if presentError(w, r, err) {
			return
		}

		usecase := api.UsecasesWithCreds(r).NewScheduledExecutionUsecase()

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		fileWriter, err := zipWriter.Create(fmt.Sprintf("decisions_of_execution_%s.ndjson", scheduledExecutionID))
		if err != nil {
			presentError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"decisions.ndjson.zip\"")
		numberOfExportedDecisions, err := usecase.ExportScheduledExecutionDecisions(scheduledExecutionID, fileWriter)
		if err != nil {
			// note: un case of security error, the header has not been sent, so we can still send a 401
			presentError(w, r, err)
			return
		}

		// nice trailer
		w.Header().Set("X-NUMBER-OF-DECISIONS", strconv.Itoa(numberOfExportedDecisions))
	}
}

func (api *API) handleListScheduledExecution() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		input := r.Context().Value(httpin.Input).(*dto.ListScheduledExecutionInput)
		scenarioId := input.ScenarioId

		if scenarioId != "" {
			if err := utils.ValidateUuid(scenarioId); err != nil {
				presentError(w, r, fmt.Errorf("search param 'scenarioId' is not a valid uuid: %w, %w", err, models.BadParameterError))
			}
		}

		usecase := api.UsecasesWithCreds(r).NewScheduledExecutionUsecase()
		executions, err := usecase.ListScheduledExecutions(scenarioId)

		if presentError(w, r, err) {
			return
		}

		PresentModelWithName(w, "scheduled_executions", utils.Map(executions, dto.AdaptScheduledExecutionDto))
	}
}

func (api *API) handleCreateScheduledExecution() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		organizationId, err := utils.OrgIDFromCtx(ctx, r)
		if presentError(w, r, err) {
			return
		}
		scenarioIterationId, err := requiredIterationParam(r)
		if presentError(w, r, err) {
			return
		}

		usecase := api.UsecasesWithCreds(r).NewScheduledExecutionUsecase()
		err = usecase.CreateScheduledExecution(models.CreateScheduledExecutionInput{
			OrganizationId:      organizationId,
			ScenarioIterationId: scenarioIterationId,
		})

		if presentError(w, r, err) {
			return
		}
		PresentNothingStatusCode(w, http.StatusCreated)
	}
}
