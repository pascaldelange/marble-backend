package app

import (
	"log"
	"marble/marble-backend/app/operators"
	"time"
)

///////////////////////////////
// Rule
///////////////////////////////

type Rule struct {
	ID            string
	DisplayOrder  int
	Name          string
	Description   string
	Formula       operators.OperatorBool
	ScoreModifier int
	CreatedAt     time.Time
}

type GetScenarioIterationRulesFilters struct {
	ScenarioIterationID *string
}

type CreateRuleInput struct {
	ScenarioIterationID string
	DisplayOrder        int
	Name                string
	Description         string
	Formula             operators.OperatorBool
	ScoreModifier       int
}

type UpdateRuleInput struct {
	ID            string
	DisplayOrder  *int
	Name          *string
	Description   *string
	Formula       *operators.OperatorBool
	ScoreModifier *int
}

///////////////////////////////
// Rule Execution
///////////////////////////////

type RuleExecution struct {
	Rule                Rule
	Result              bool
	ResultScoreModifier int
	Error               RuleExecutionError
}

///////////////////////////////
// RuleExecutionError
///////////////////////////////

type RuleExecutionError int

const (
	FieldEmptyOrMissing RuleExecutionError = 200
)

func (r RuleExecutionError) String() string {
	switch r {
	case FieldEmptyOrMissing:
		return "A field in rule is empty or missing"
	}
	return ""
}

///////////////////////////////
//
///////////////////////////////

func (r Rule) Eval(dataAccessor operators.DataAccessor) (RuleExecution, error) {

	// Eval the Node
	res, err := r.Formula.Eval(dataAccessor)
	if err != nil {
		log.Printf("Error while evaluating rule %s: %v", r.Name, err)
		return RuleExecution{}, err
	}

	score := 0
	if res {
		score = r.ScoreModifier
	}

	re := RuleExecution{
		Rule:                r,
		Result:              res,
		ResultScoreModifier: score,
		// TODO error ?
	}

	//log.Printf("Rule %s is %v, score = %v", r.RootNode.Print(p), res, score)

	return re, nil
}
