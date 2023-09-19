package evaluate_scenario

import (
	"context"

	"github.com/checkmarble/marble-backend/models"
	"github.com/checkmarble/marble-backend/repositories"
	"github.com/checkmarble/marble-backend/usecases/transaction"
)

type DataAccessor struct {
	DataModel                  models.DataModel
	Payload                    models.PayloadReader
	orgTransactionFactory      transaction.Factory
	organizationId             string
	ingestedDataReadRepository repositories.IngestedDataReadRepository
}

func (d *DataAccessor) GetDbField(ctx context.Context, triggerTableName string, path []string, fieldName string) (interface{}, error) {

	return transaction.InOrganizationSchema(
		d.orgTransactionFactory,
		d.organizationId,
		func(tx repositories.Transaction) (interface{}, error) {
			return d.ingestedDataReadRepository.GetDbField(tx, models.DbFieldReadParams{
				TriggerTableName: models.TableName(triggerTableName),
				Path:             models.ToLinkNames(path),
				FieldName:        models.FieldName(fieldName),
				DataModel:        d.DataModel,
				Payload:          d.Payload,
			})
		})
}
