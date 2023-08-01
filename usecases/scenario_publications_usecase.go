package usecases

import (
	"context"
	"marble/marble-backend/models"
	"marble/marble-backend/repositories"
	"marble/marble-backend/usecases/scenarios"
	"marble/marble-backend/usecases/security"
)

type ScenarioPublicationUsecase struct {
	transactionFactory              repositories.TransactionFactory
	scenarioPublicationsRepository  repositories.ScenarioPublicationRepository
	scenarioReadRepository          repositories.ScenarioReadRepository
	scenarioIterationReadRepository repositories.ScenarioIterationReadRepository
	OrganizationIdOfContext         func() (string, error)
	enforceSecurity                 security.EnforceSecurityScenario
	scenarioPublisher               scenarios.ScenarioPublisher
}

func (usecase *ScenarioPublicationUsecase) GetScenarioPublication(scenarioPublicationID string) (models.ScenarioPublication, error) {
	scenarioPublication, err := usecase.scenarioPublicationsRepository.GetScenarioPublicationById(nil, scenarioPublicationID)
	if err != nil {
		return models.ScenarioPublication{}, err
	}

	// Enforce permissions
	if err := usecase.enforceSecurity.ReadScenarioPublication(scenarioPublication); err != nil {
		return models.ScenarioPublication{}, err
	}
	return scenarioPublication, nil
}

func (usecase *ScenarioPublicationUsecase) ListScenarioPublications(filters models.ListScenarioPublicationsFilters) ([]models.ScenarioPublication, error) {
	organizationId, err := usecase.OrganizationIdOfContext()
	if err != nil {
		return nil, err
	}

	// Enforce permissions
	if err := usecase.enforceSecurity.ListScenarios(organizationId); err != nil {
		return nil, err
	}

	return usecase.scenarioPublicationsRepository.ListScenarioPublicationsOfOrganization(nil, organizationId, filters)
}

func (usecase *ScenarioPublicationUsecase) ExecuteScenarioPublicationAction(ctx context.Context, input models.PublishScenarioIterationInput) ([]models.ScenarioPublication, error) {
	return repositories.TransactionReturnValue(usecase.transactionFactory, models.DATABASE_MARBLE_SCHEMA, func(tx repositories.Transaction) ([]models.ScenarioPublication, error) {
		scenarioIteration, err := usecase.scenarioIterationReadRepository.GetScenarioIteration(tx, input.ScenarioIterationId)
		if err != nil {
			return []models.ScenarioPublication{}, err
		}

		scenario, err := usecase.scenarioReadRepository.GetScenarioById(tx, scenarioIteration.ScenarioId)
		if err != nil {
			return []models.ScenarioPublication{}, err
		}

		// Enforce permissions
		if err := usecase.enforceSecurity.PublishScenario(scenario); err != nil {
			return []models.ScenarioPublication{}, err
		}

		return usecase.scenarioPublisher.PublishOrUnpublishIteration(tx, scenario.OrganizationId, input)
	})

}
