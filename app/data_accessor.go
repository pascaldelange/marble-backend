package app

import (
	"context"
	"errors"
	"fmt"
	"marble/marble-backend/app/operators"
)

type DataAccessorImpl struct {
	DataModel  DataModel
	Payload    DynamicStructWithReader
	repository RepositoryInterface
}

type DbFieldReadParams struct {
	TriggerTableName TableName
	Path             []LinkName
	FieldName        FieldName
	DataModel        DataModel
	Payload          DynamicStructWithReader
}

var ErrNoRowsReadInDB = errors.New("No rows read while reading DB field")

func (d *DataAccessorImpl) GetPayloadField(fieldName string) interface{} {
	return d.GetPayloadField(fieldName)
}
func (d *DataAccessorImpl) GetDbField(path []string, fieldName string) (interface{}, error) {
	return d.repository.GetDbField(context.TODO(), DbFieldReadParams{
		Path:      toLinkNames(path),
		FieldName: FieldName(fieldName),
		DataModel: d.DataModel,
		Payload:   d.Payload,
	})
}

func (d *DataAccessorImpl) ValidateDbFieldReadConsistency(path []string, fieldName string) error {
	if len(path) == 0 {
		return fmt.Errorf("Path is empty: %w", operators.ErrDbReadInconsistentWithDataModel)
	}

	for _, tableName := range path {
		_, ok := d.DataModel.Tables[TableName(tableName)]
		if !ok {
			return fmt.Errorf("Table %s in path not found in data model: %w", tableName, operators.ErrDbReadInconsistentWithDataModel)
		}
	}

	lastTable := d.DataModel.Tables[TableName(path[len(path)-1])]
	_, ok := lastTable.Fields[FieldName(fieldName)]
	if !ok {
		return fmt.Errorf("Field %s in table %s not found in data model: %w", fieldName, lastTable.Name, operators.ErrDbReadInconsistentWithDataModel)
	}

	return nil
}
