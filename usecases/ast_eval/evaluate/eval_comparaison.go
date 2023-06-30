package evaluate

import (
	"fmt"
	"marble/marble-backend/models/ast"
)

type Comparison struct {
	Function ast.Function
}

func NewComparison(f ast.Function) Comparison {
	return Comparison{
		Function: f,
	}
}

func (f Comparison) Evaluate(arguments ast.Arguments) (any, error) {
	// promote to float64
	operandsFloat, err := promoteOperandsToFloat64(arguments.Args, f.Function)
	if err != nil {
		return nil, err
	}
	return f.comparisonFunction(operandsFloat)
}

func (f Comparison)comparisonFunction(arguments []float64) (bool, error) {
	l, r, err := leftAndRight(arguments)
	if err != nil {
		return false, err
	}

	if f.Function == ast.FUNC_GREATER {
		return l > r, nil
	}
	if f.Function == ast.FUNC_LESS {
		return l < r, nil
	}
	if f.Function == ast.FUNC_EQUAL {
		// comparing float64 is not smart, but not illegal
		return l == r, nil
	}

	return false, fmt.Errorf("Comparison does not support %s function", f.Function.DebugString())
}
