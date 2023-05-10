package operators

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type DataAccessorImpl struct{}

func (d *DataAccessorImpl) GetPayloadField(fieldName string) interface{} {
	var val bool
	if fieldName == "true" {
		val = true
	} else if fieldName == "false" {
		val = false
	} else {
		return nil
	}
	return &val
}

func (d *DataAccessorImpl) GetDbField(path []string, fieldName string) (interface{}, error) {
	var val pgtype.Bool
	if fieldName == "true" {
		val = pgtype.Bool{Bool: true, Valid: true}
	} else if fieldName == "false" {
		val = pgtype.Bool{Bool: false, Valid: true}
	} else {
		val = pgtype.Bool{Bool: true, Valid: false}
	}
	return val, nil
}
func (d *DataAccessorImpl) ValidateDbFieldReadConsistency(path []string, fieldName string) error {
	return nil
}

func TestLogicEval(t *testing.T) {
	type testCase struct {
		name     string
		operator OperatorBool
		expected bool
	}
	dataAccessor := DataAccessorImpl{}

	cases := []testCase{
		{
			name: "1",
			operator: &EqBool{
				Left: &True{},
				Right: &EqBool{
					Left:  &False{},
					Right: &False{},
				},
			},
			expected: true,
		},
		{
			name: "2",
			operator: &EqBool{
				Left: &True{},
				Right: &EqBool{
					Left:  &False{},
					Right: &True{},
				},
			},
			expected: false},
		{
			name: "3",
			operator: &EqBool{
				Left: &True{},
				Right: &EqBool{
					Left:  &DbFieldBool{Path: []string{"a", "b"}, FieldName: "true"},
					Right: &True{},
				},
			},
			expected: true,
		},
		{
			name: "4",
			operator: &EqBool{
				Left:  &True{},
				Right: &False{},
			},
			expected: false,
		},
		{
			name:     "Payload true",
			operator: &PayloadFieldBool{FieldName: "true"},
			expected: true,
		},
		{
			name:     "Payload false",
			operator: &PayloadFieldBool{FieldName: "false"},
			expected: false,
		},
		{
			name: "variadic and: 3 ops, true",
			operator: &And{
				Operands: []OperatorBool{&True{}, &True{}, &True{}},
			},
			expected: true,
		},
		{
			name: "variadic and: 3 ops, false",
			operator: &And{
				Operands: []OperatorBool{&True{}, &True{}, &False{}},
			},
			expected: false,
		},
		{
			name: "variadic and: 1 ops, false",
			operator: &And{
				Operands: []OperatorBool{&False{}},
			},
			expected: false,
		},
		{
			name: "variadic or: 3 ops, true",
			operator: &Or{
				Operands: []OperatorBool{&False{}, &True{}, &False{}},
			},
			expected: true,
		},
		{
			name: "variadic and: 3 ops, false",
			operator: &Or{
				Operands: []OperatorBool{&False{}, &False{}, &False{}},
			},
			expected: false,
		},
		{
			name: "variadic and: 1 ops, false",
			operator: &And{
				Operands: []OperatorBool{&False{}},
			},
			expected: false,
		},
		{
			name: "NOT true",
			operator: &Not{
				Child: &True{},
			},
			expected: false,
		},
	}
	asserts := assert.New(t)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.operator.Eval(&dataAccessor)

			if err != nil {
				t.Errorf("error: %v", err)
			}

			asserts.Equal(c.expected, got)
		})
	}
}

func TestLogicEvalErrorCase(t *testing.T) {
	type testCase struct {
		name     string
		operator OperatorBool
	}
	dataAccessor := DataAccessorImpl{}

	cases := []testCase{
		{
			name:     "Payload nil",
			operator: &PayloadFieldBool{FieldName: "nil"},
		},
		{
			name:     "Payload nil",
			operator: &DbFieldBool{Path: []string{"table"}, FieldName: "nil"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := c.operator.Eval(&dataAccessor)

			if err == nil {
				t.Errorf("Was expecting an error reading a null field")
			}

		})
	}
}

func TestMarshalUnMarshal(t *testing.T) {
	type testCase struct {
		name     string
		operator OperatorBool
	}
	dataAccessor := DataAccessorImpl{}
	asserts := assert.New(t)

	cases := []testCase{
		{
			name: "Simple Equal",
			operator: &EqBool{
				Left:  &False{},
				Right: &True{},
			},
		},
		{
			name: "Larger tree",
			operator: &EqBool{
				Left: &True{},
				Right: &EqBool{
					Left:  &DbFieldBool{Path: []string{"a", "b"}, FieldName: "true"},
					Right: &True{},
				},
			},
		},
		{
			name: "Variadic and",
			operator: &And{
				Operands: []OperatorBool{&True{}, &True{}, &False{}},
			},
		},
		{
			name: "Variadic or",
			operator: &Or{
				Operands: []OperatorBool{&True{}, &True{}, &False{}},
			},
		},
		{
			name: "Not true",
			operator: &Not{
				Child: True{},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			JSONbytes, err := c.operator.MarshalJSON()
			if err != nil {
				t.Errorf("error marshaling operator: %v", err)
			}

			t.Log("JSONbytes", string(JSONbytes))

			rootOperator, err := UnmarshalOperatorBool(JSONbytes)
			if err != nil {
				t.Errorf("error unmarshaling operator: %v", err)
			}

			expected, err := c.operator.Eval(&dataAccessor)
			if err != nil {
				t.Errorf("error: %v", err)
			}
			got, err := rootOperator.Eval(&dataAccessor)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			asserts.Equal(expected, got, "evaluated operator should be the same as the original")

		})
	}

}

func TestMarshalContracts(t *testing.T) {
	for typeKey, creatorFunc := range operatorFromType {
		testname := typeKey
		t.Run(testname, func(t *testing.T) {

			op := creatorFunc()
			JSONop, err := op.MarshalJSON()
			if err != nil {
				t.Errorf("unable to marshal operator to JSON")
			}

			var mapFormatOp map[string]interface{}
			err = json.Unmarshal(JSONop, &mapFormatOp)
			fmt.Println(mapFormatOp)
			for k := range mapFormatOp {
				if k != "type" && k != "staticData" && k != "children" {
					t.Errorf("marshaled operator contains unexpected key: %s", k)
				}
			}
			_, ok := mapFormatOp["type"]
			if !ok {
				t.Errorf(`marshaled operator does not contain mandatory field "type"`)
			}
		})
	}
}

func TestMarshallBoolOperators(t *testing.T) {
	type testCase struct {
		name     string
		operator OperatorBool
		expected string
	}
	asserts := assert.New(t)
	cases := []testCase{
		{
			name:     "true",
			operator: True{},
			expected: `{"type":"TRUE"}`,
		},
		{
			name:     "false",
			operator: False{},
			expected: `{"type":"FALSE"}`,
		},
		{
			name: "equal",
			operator: &EqBool{
				Left:  &True{},
				Right: &False{},
			},
			expected: `{"type":"EQUAL_BOOL","children":[{"type":"TRUE"},{"type":"FALSE"}]}`,
		},
		{
			name: "db field bool",
			operator: &DbFieldBool{
				Path:      []string{"a", "b"},
				FieldName: "c",
			},
			expected: `{"type":"DB_FIELD_BOOL","staticData":{"path":["a","b"],"fieldName":"c"}}`,
		},
		{
			name: "variadic and",
			operator: &And{
				Operands: []OperatorBool{&True{}, &True{}, &False{}},
			},
			expected: `{"type":"AND","children":[{"type":"TRUE"},{"type":"TRUE"},{"type":"FALSE"}]}`,
		},
		{
			name: "variadic or",
			operator: &Or{
				Operands: []OperatorBool{&True{}, &True{}, &False{}},
			},
			expected: `{"type":"OR","children":[{"type":"TRUE"},{"type":"TRUE"},{"type":"FALSE"}]}`,
		},
		{
			name: "not true",
			operator: &Not{
				Child: True{},
			},
			expected: `{"type":"NOT","children":[{"type":"TRUE"}]}`,
		},
		{
			name: "eq with null",
			operator: &EqBool{
				Left:  &True{},
				Right: nil,
			},
			expected: `{"type":"EQUAL_BOOL","children":[{"type":"TRUE"},null]}`,
		},
		{
			name: "or with null",
			operator: &Or{
				Operands: []OperatorBool{&True{}, nil, &False{}},
			},
			expected: `{"type":"OR","children":[{"type":"TRUE"},null,{"type":"FALSE"}]}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			JSONbytes, err := c.operator.MarshalJSON()
			if err != nil {
				t.Errorf("error marshaling operator: %v", err)
			}
			asserts.Equal(c.expected, string(JSONbytes))
		})
	}
}

func TestUnmarshallBoolOperators(t *testing.T) {
	type testCase struct {
		name     string
		expected OperatorBool
		json     string
	}
	asserts := assert.New(t)
	cases := []testCase{
		{
			name:     "true",
			expected: &True{},
			json:     `{"type":"TRUE"}`,
		},
		{
			name:     "false",
			expected: &False{},
			json:     `{"type":"FALSE"}`,
		},
		{
			name: "equal",
			expected: &EqBool{
				Left:  &True{},
				Right: &False{},
			},
			json: `{"type":"EQUAL_BOOL","children":[{"type":"TRUE"},{"type":"FALSE"}]}`,
		},
		{
			name: "equal",
			expected: &DbFieldBool{
				Path:      []string{"a", "b"},
				FieldName: "c",
			},
			json: `{"type":"DB_FIELD_BOOL","staticData":{"path":["a","b"],"fieldName":"c"}}`,
		},
		{
			name: "eq with null",
			expected: &EqBool{
				Left:  &True{},
				Right: nil,
			},
			json: `{"type":"EQUAL_BOOL","children":[{"type":"TRUE"},null]}`,
		},
		{
			name: "or with null",
			expected: &Or{
				Operands: []OperatorBool{&True{}, nil, &False{}},
			},
			json: `{"type":"OR","children":[{"type":"TRUE"},null,{"type":"FALSE"}]}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			operator, err := UnmarshalOperatorBool([]byte(c.json))
			if err != nil {
				t.Errorf("error marshaling operator: %v", err)
			}
			asserts.Equal(c.expected, operator)
		})
	}
}

func TestInvalidOperators(t *testing.T) {
	type testCase struct {
		name     string
		operator OperatorBool
	}

	cases := []testCase{
		{
			name:     "empty and",
			operator: &And{},
		},
		{
			name:     "and with null",
			operator: &And{Operands: []OperatorBool{&True{}, nil}},
		},
		{
			name:     "and with null first",
			operator: &And{Operands: []OperatorBool{nil, &True{}, &False{}}},
		},
		{
			name: "eq",
			operator: &EqBool{
				Left:  &True{},
				Right: nil,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.operator.IsValid() {
				t.Errorf("operator should be invalid")
			}
		})
	}
}
