package api

import (
	"context"
	"encoding/json"
	"errors"
	"marble/marble-backend/app"
	"net/http"

	"github.com/ggicci/httpin"
	"golang.org/x/exp/slog"
)

type OrganizationAppInterface interface {
	GetOrganizations(ctx context.Context) ([]app.Organization, error)
	CreateOrganization(ctx context.Context, organization app.CreateOrganizationInput) (app.Organization, error)

	GetOrganization(ctx context.Context, organizationID string) (app.Organization, error)
	UpdateOrganization(ctx context.Context, organization app.UpdateOrganizationInput) (app.Organization, error)
	SoftDeleteOrganization(ctx context.Context, organizationID string) error
}

type APIOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAPIOrganization(org app.Organization) APIOrganization {
	return APIOrganization{
		ID:   org.ID,
		Name: org.Name,
	}
}

func (api *API) handleGetOrganizations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		organizations, err := api.app.GetOrganizations(ctx)
		if err != nil {
			api.logger.ErrorCtx(ctx, "Error getting organizations: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		apiOrganizations := make([]APIOrganization, len(organizations))
		for i, org := range organizations {
			apiOrganizations[i] = NewAPIOrganization(org)
		}

		err = json.NewEncoder(w).Encode(&apiOrganizations)
		if err != nil {
			api.logger.ErrorCtx(ctx, "Could not encode response JSON: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

type CreateOrganizationBody struct {
	Name         string `json:"name"`
	DatabaseName string `json:"databaseName"`
}

type CreateOrganizationInput struct {
	Body *CreateOrganizationBody `in:"body=json"`
}

func (api *API) handlePostOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		input := ctx.Value(httpin.Input).(*CreateOrganizationInput)
		requestData := input.Body

		org, err := api.app.CreateOrganization(ctx, app.CreateOrganizationInput{
			Name:         requestData.Name,
			DatabaseName: requestData.DatabaseName,
		})
		if err != nil {
			api.logger.ErrorCtx(ctx, "Error creating organizations: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(NewAPIOrganization(org))
		if err != nil {
			api.logger.ErrorCtx(ctx, "Could not encode response JSON: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

type GetOrganizationInput struct {
	orgID string `in:"path=orgID"`
}

func (api *API) handleGetOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		orgID := ctx.Value(httpin.Input).(*GetOrganizationInput).orgID
		logger := api.logger.With(slog.String("orgID", orgID))

		org, err := api.app.GetOrganization(ctx, orgID)
		if errors.Is(err, app.ErrNotFoundInRepository) {
			http.Error(w, "", http.StatusNotFound)
			return
		} else if err != nil {
			logger.ErrorCtx(ctx, "Error getting organization: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(NewAPIOrganization(org))
		if err != nil {
			logger.ErrorCtx(ctx, "Could not encode response JSON: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

type UpdateOrganizationBody struct {
	Name         *string `json:"name,omitempty"`
	DatabaseName *string `json:"databaseName,omitempty"`
}

type UpdateOrganizationInput struct {
	OrgID string                  `in:"path=orgID"`
	Body  *UpdateOrganizationBody `in:"body=json"`
}

func (api *API) handlePutOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		input := ctx.Value(httpin.Input).(*UpdateOrganizationInput)
		requestData := input.Body
		orgID := input.OrgID
		logger := api.logger.With(slog.String("orgID", orgID))

		org, err := api.app.UpdateOrganization(ctx, app.UpdateOrganizationInput{
			ID:           orgID,
			Name:         requestData.Name,
			DatabaseName: requestData.DatabaseName,
		})
		if errors.Is(err, app.ErrNotFoundInRepository) {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		} else if err != nil {
			logger.ErrorCtx(ctx, "Error updating organizations: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(NewAPIOrganization(org))
		if err != nil {
			logger.ErrorCtx(ctx, "Could not encode response JSON: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

type DeleteOrganizationInput struct {
	orgID string `in:"path=orgID"`
}

func (api *API) handleDeleteOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		orgID := ctx.Value(httpin.Input).(*DeleteOrganizationInput).orgID
		logger := api.logger.With(slog.String("orgID", orgID))

		err := api.app.SoftDeleteOrganization(ctx, orgID)
		if errors.Is(err, app.ErrNotFoundInRepository) {
			http.Error(w, "", http.StatusNotFound)
			return
		} else if err != nil {
			// Could not execute request
			logger.ErrorCtx(ctx, "Error deleting organization: \n"+err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
