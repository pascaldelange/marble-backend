package usecases

import (
	"errors"
	"testing"

	"github.com/checkmarble/marble-backend/mocks"
	"github.com/checkmarble/marble-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ScenarioUsecaseTestSuite struct {
	suite.Suite
	transaction             *mocks.Transaction
	transactionFactory      *mocks.TransactionFactory
	enforceSecurity         *mocks.EnforceSecurity
	scenarioReadRepository  *mocks.ScenarioReadRepository
	scenarioWriteRepository *mocks.ScenarioWriteRepository

	organizationId string
	scenarioId     string
	scenario       models.Scenario
	securityError  error
}

func (suite *ScenarioUsecaseTestSuite) SetupTest() {
	suite.transaction = new(mocks.Transaction)
	suite.enforceSecurity = new(mocks.EnforceSecurity)
	suite.transactionFactory = &mocks.TransactionFactory{TxMock: suite.transaction}
	suite.scenarioReadRepository = new(mocks.ScenarioReadRepository)
	suite.scenarioWriteRepository = new(mocks.ScenarioWriteRepository)
	suite.securityError = errors.New("some security error")

	suite.organizationId = "25ab6323-1657-4a52-923a-ef6983fe4532"
	suite.scenarioId = "c5968ff7-6142-4623-a6b3-1539f345e5fa"
	suite.scenario = models.Scenario{
		Id:             suite.scenarioId,
		OrganizationId: suite.organizationId,
	}

}

func (suite *ScenarioUsecaseTestSuite) makeUsecase() *ScenarioUsecase {
	return &ScenarioUsecase{
		transactionFactory: suite.transactionFactory,
		organizationIdOfContext: func() (string, error) {
			return suite.organizationId, nil
		},
		enforceSecurity:         suite.enforceSecurity,
		scenarioReadRepository:  suite.scenarioReadRepository,
		scenarioWriteRepository: suite.scenarioWriteRepository,
	}
}

func (suite *ScenarioUsecaseTestSuite) AssertExpectations() {
	t := suite.T()
	suite.transaction.AssertExpectations(t)
	suite.enforceSecurity.AssertExpectations(t)
	suite.transactionFactory.AssertExpectations(t)
	suite.scenarioReadRepository.AssertExpectations(t)
	suite.scenarioWriteRepository.AssertExpectations(t)
}

func (suite *ScenarioUsecaseTestSuite) TestListScenarios() {

	var expected = []models.Scenario{suite.scenario}
	suite.scenarioReadRepository.On("ListScenariosOfOrganization", nil, suite.organizationId).Return(expected, nil)
	suite.enforceSecurity.On("ReadScenario", suite.scenario).Return(nil)

	result, err := suite.makeUsecase().ListScenarios()

	t := suite.T()
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestListScenarios_security() {

	suite.scenarioReadRepository.On("ListScenariosOfOrganization", nil, suite.organizationId).Return([]models.Scenario{suite.scenario}, nil)
	suite.enforceSecurity.On("ReadScenario", suite.scenario).Return(suite.securityError)

	_, err := suite.makeUsecase().ListScenarios()

	assert.ErrorIs(suite.T(), err, suite.securityError)
	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestGetScenario() {

	suite.scenarioReadRepository.On("GetScenarioById", nil, suite.scenarioId).Return(suite.scenario, nil)
	suite.enforceSecurity.On("ReadScenario", suite.scenario).Return(nil)

	result, err := suite.makeUsecase().GetScenario(suite.scenarioId)

	t := suite.T()
	assert.NoError(t, err)
	assert.Equal(t, suite.scenario, result)

	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestGetScenario_security() {

	suite.scenarioReadRepository.On("GetScenarioById", nil, suite.scenarioId).Return(suite.scenario, nil)
	suite.enforceSecurity.On("ReadScenario", suite.scenario).Return(suite.securityError)

	_, err := suite.makeUsecase().GetScenario(suite.scenarioId)

	assert.ErrorIs(suite.T(), err, suite.securityError)
	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestUpdateScenario() {

	scenarioInput := models.UpdateScenarioInput{
		Id: suite.scenarioId,
	}

	updatedScenario := models.Scenario{
		Id:   suite.scenarioId,
		Name: "updated scenario",
	}

	suite.transactionFactory.On("Transaction", models.DATABASE_MARBLE_SCHEMA, mock.Anything).Return(nil)
	suite.scenarioReadRepository.On("GetScenarioById", suite.transaction, suite.scenarioId).Return(suite.scenario, nil).Once()
	suite.enforceSecurity.On("UpdateScenario", suite.scenario).Return(nil)

	suite.scenarioWriteRepository.On("UpdateScenario", suite.transaction, scenarioInput).Return(nil)
	suite.scenarioReadRepository.On("GetScenarioById", suite.transaction, suite.scenarioId).Return(updatedScenario, nil).Once()

	result, err := suite.makeUsecase().UpdateScenario(scenarioInput)

	t := suite.T()
	assert.NoError(t, err)
	assert.Equal(t, updatedScenario, result)

	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestUpdateScenario_security() {

	scenarioInput := models.UpdateScenarioInput{
		Id: suite.scenarioId,
	}

	suite.transactionFactory.On("Transaction", models.DATABASE_MARBLE_SCHEMA, mock.Anything).Return(nil)
	suite.scenarioReadRepository.On("GetScenarioById", suite.transaction, suite.scenarioId).Return(suite.scenario, nil).Once()
	suite.enforceSecurity.On("UpdateScenario", suite.scenario).Return(suite.securityError)

	_, err := suite.makeUsecase().UpdateScenario(scenarioInput)

	assert.ErrorIs(suite.T(), err, suite.securityError)
	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestCreateScenario() {

	createScenarioInput := models.CreateScenarioInput{
		Name: "new scenario",
	}

	suite.enforceSecurity.On("CreateScenario", suite.organizationId).Return(nil)

	suite.transactionFactory.On("Transaction", models.DATABASE_MARBLE_SCHEMA, mock.Anything).Return(nil)
	suite.scenarioWriteRepository.On("CreateScenario", suite.transaction, suite.organizationId, createScenarioInput, mock.Anything).Return(nil)
	suite.scenarioReadRepository.On("GetScenarioById", suite.transaction, mock.Anything).Return(suite.scenario, nil).Once()

	result, err := suite.makeUsecase().CreateScenario(createScenarioInput)

	t := suite.T()
	assert.NoError(t, err)
	assert.Equal(t, suite.scenario, result)

	suite.AssertExpectations()
}

func (suite *ScenarioUsecaseTestSuite) TestCreateScenario_security() {

	suite.enforceSecurity.On("CreateScenario", suite.organizationId).Return(suite.securityError)

	_, err := suite.makeUsecase().CreateScenario(models.CreateScenarioInput{})
	assert.ErrorIs(suite.T(), err, suite.securityError)

	suite.AssertExpectations()
}

func TestScenarioUsecase(t *testing.T) {
	suite.Run(t, new(ScenarioUsecaseTestSuite))
}
