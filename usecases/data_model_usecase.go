package usecases

import (
	"context"
	"fmt"

	"github.com/checkmarble/marble-backend/models"
	"github.com/checkmarble/marble-backend/repositories"
	"github.com/checkmarble/marble-backend/usecases/executor_factory"
	"github.com/checkmarble/marble-backend/usecases/security"
	"github.com/google/uuid"
)

type DataModelUseCase struct {
	dataModelRepository          repositories.DataModelRepository
	enforceSecurity              security.EnforceSecurityOrganization
	executorFactory              executor_factory.ExecutorFactory
	organizationSchemaRepository repositories.OrganizationSchemaRepository
	transactionFactory           executor_factory.TransactionFactory
}

func (usecase *DataModelUseCase) GetDataModel(ctx context.Context, organizationID string) (models.DataModel, error) {
	if err := usecase.enforceSecurity.ReadDataModel(); err != nil {
		return models.DataModel{}, err
	}

	return usecase.dataModelRepository.GetDataModel(
		ctx,
		usecase.executorFactory.NewExecutor(),
		organizationID,
		true,
	)
}

func (usecase *DataModelUseCase) CreateDataModelTable(ctx context.Context, organizationId, name, description string) (string, error) {
	if err := usecase.enforceSecurity.WriteDataModel(organizationId); err != nil {
		return "", err
	}

	defaultFields := []models.DataModelField{
		{
			Name:        "object_id",
			Description: fmt.Sprintf("required id on all objects in the %s table", name),
			Type:        models.String.String(),
			Nullable:    false,
		},
		{
			Name:        "updated_at",
			Description: fmt.Sprintf("required timestamp on all objects in the %s table", name),
			Type:        models.Timestamp.String(),
			Nullable:    false,
		},
	}

	tableID := uuid.New().String()
	err := usecase.transactionFactory.Transaction(ctx, func(tx repositories.Executor) error {
		err := usecase.dataModelRepository.CreateDataModelTable(ctx, tx, organizationId, tableID, name, description)
		if err != nil {
			return err
		}

		for _, field := range defaultFields {
			err := usecase.dataModelRepository.CreateDataModelField(ctx, tx, tableID, uuid.New().String(), field)
			if err != nil {
				return err
			}
		}

		// if it returns an error, rolls back the other transaction
		return usecase.transactionFactory.TransactionInOrgSchema(ctx, organizationId, func(orgTx repositories.Executor) error {
			if err := usecase.organizationSchemaRepository.CreateSchemaIfNotExists(ctx, orgTx); err != nil {
				return err
			}
			return usecase.organizationSchemaRepository.CreateTable(ctx, orgTx, name)
		})
	})
	return tableID, err
}

func (usecase *DataModelUseCase) UpdateDataModelTable(ctx context.Context, tableID, description string) error {
	if table, err := usecase.dataModelRepository.GetDataModelTable(
		ctx,
		usecase.executorFactory.NewExecutor(),
		tableID,
	); err != nil {
		return err
	} else if err := usecase.enforceSecurity.WriteDataModel(table.OrganizationID); err != nil {
		return err
	}

	return usecase.dataModelRepository.UpdateDataModelTable(
		ctx,
		usecase.executorFactory.NewExecutor(),
		tableID,
		description,
	)
}

func (usecase *DataModelUseCase) CreateDataModelField(ctx context.Context, tableID string, field models.DataModelField) (string, error) {
	fieldID := uuid.New().String()
	err := usecase.transactionFactory.Transaction(ctx, func(tx repositories.Executor) error {
		table, err := usecase.dataModelRepository.GetDataModelTable(ctx, tx, tableID)
		if err != nil {
			return err
		}
		if err := usecase.enforceSecurity.WriteDataModel(table.OrganizationID); err != nil {
			return err
		}

		if err := usecase.dataModelRepository.CreateDataModelField(ctx, tx, tableID, fieldID, field); err != nil {
			return err
		}

		// if it returns an error, automatically rolls back the other transaction
		return usecase.transactionFactory.TransactionInOrgSchema(
			ctx,
			table.OrganizationID,
			func(orgTx repositories.Executor) error {
				return usecase.organizationSchemaRepository.CreateField(ctx, orgTx, table.Name, field)
			},
		)
	})
	return fieldID, err
}

func (usecase *DataModelUseCase) UpdateDataModelField(ctx context.Context, fieldID string, input models.UpdateDataModelFieldInput) error {
	exec := usecase.executorFactory.NewExecutor()
	field, err := usecase.dataModelRepository.GetDataModelField(ctx, exec, fieldID)
	if err != nil {
		return err
	}

	if table, err := usecase.dataModelRepository.GetDataModelTable(ctx, exec, field.TableId); err != nil {
		return err
	} else if err := usecase.enforceSecurity.WriteDataModel(table.OrganizationID); err != nil {
		return err
	}

	if input.IsEnum != nil && *input.IsEnum &&
		(field.DataType != models.String &&
			field.DataType != models.Int &&
			field.DataType != models.Float) {
		return fmt.Errorf("enum fields can only be of type string or numeric: %w", models.BadParameterError)
	}

	return usecase.dataModelRepository.UpdateDataModelField(ctx, exec, fieldID, input)
}

func (usecase *DataModelUseCase) CreateDataModelLink(ctx context.Context, link models.DataModelLink) error {
	if err := usecase.enforceSecurity.WriteDataModel(link.OrganizationID); err != nil {
		return err
	}
	return usecase.dataModelRepository.CreateDataModelLink(ctx,
		usecase.executorFactory.NewExecutor(), link)
}

func (usecase *DataModelUseCase) DeleteDataModel(ctx context.Context, organizationID string) error {
	if err := usecase.enforceSecurity.WriteDataModel(organizationID); err != nil {
		return err
	}

	return usecase.transactionFactory.Transaction(ctx, func(tx repositories.Executor) error {
		if err := usecase.dataModelRepository.DeleteDataModel(ctx, tx, organizationID); err != nil {
			return err
		}

		// if it returns an error, automatically rolls back the other transaction
		return usecase.transactionFactory.TransactionInOrgSchema(ctx, organizationID, func(orgTx repositories.Executor) error {
			return usecase.organizationSchemaRepository.DeleteSchema(ctx, orgTx)
		})
	})
}
