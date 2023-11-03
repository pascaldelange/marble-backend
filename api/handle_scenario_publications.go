package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/checkmarble/marble-backend/dto"
	"github.com/checkmarble/marble-backend/models"
	"github.com/checkmarble/marble-backend/utils"
)

type APIScenarioPublication struct {
	Id                  string    `json:"id"`
	Rank                int32     `json:"rank"`
	ScenarioId          string    `json:"scenarioID"`
	ScenarioIterationId string    `json:"scenarioIterationID"`
	PublicationAction   string    `json:"publicationAction"`
	CreatedAt           time.Time `json:"createdAt"`
}

func NewAPIScenarioPublication(sp models.ScenarioPublication) APIScenarioPublication {
	return APIScenarioPublication{
		Id:                  sp.Id,
		Rank:                sp.Rank,
		ScenarioId:          sp.ScenarioId,
		ScenarioIterationId: sp.ScenarioIterationId,
		PublicationAction:   sp.PublicationAction.String(),
		CreatedAt:           sp.CreatedAt,
	}
}

func (api *API) ListScenarioPublications(c *gin.Context) {
	scenarioID := c.Query("scenarioID")
	scenarioIterationID := c.Query("scenarioIterationID")

	usecase := api.UsecasesWithCreds(c.Request).NewScenarioPublicationUsecase()
	scenarioPublications, err := usecase.ListScenarioPublications(models.ListScenarioPublicationsFilters{
		ScenarioId:          utils.PtrTo(scenarioID, &utils.PtrToOptions{OmitZero: true}),
		ScenarioIterationId: utils.PtrTo(scenarioIterationID, &utils.PtrToOptions{OmitZero: true}),
	})
	if presentError(c.Writer, c.Request, err) {
		return
	}
	c.JSON(http.StatusOK, utils.Map(scenarioPublications, NewAPIScenarioPublication))
}

func (api *API) CreateScenarioPublication(c *gin.Context) {
	var data dto.CreateScenarioPublicationBody
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	usecase := api.UsecasesWithCreds(c.Request).NewScenarioPublicationUsecase()
	scenarioPublications, err := usecase.ExecuteScenarioPublicationAction(models.PublishScenarioIterationInput{
		ScenarioIterationId: data.ScenarioIterationId,
		PublicationAction:   models.PublicationActionFrom(data.PublicationAction),
	})
	if presentError(c.Writer, c.Request, err) {
		return
	}
	c.JSON(http.StatusOK, utils.Map(scenarioPublications, NewAPIScenarioPublication))
}

func (api *API) GetScenarioPublication(c *gin.Context) {
	scenarioPublicationID := c.Param("publication_id")

	usecase := api.UsecasesWithCreds(c.Request).NewScenarioPublicationUsecase()
	scenarioPublication, err := usecase.GetScenarioPublication(scenarioPublicationID)
	if presentError(c.Writer, c.Request, err) {
		return
	}
	c.JSON(http.StatusOK, NewAPIScenarioPublication(scenarioPublication))
}
