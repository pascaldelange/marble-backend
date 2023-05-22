package pg_repository

import (
	"context"
	"log"

	"marble/marble-backend/app"
	"marble/marble-backend/app/operators"
	"marble/marble-backend/models"
	"marble/marble-backend/utils"
)

func (r *PGRepository) Seed() {

	///////////////////////////////
	// Organizations
	///////////////////////////////

	organizations, err := r.GetOrganizations(context.TODO())
	if err != nil {
		log.Printf("error getting organizations: %v", err)
	}
	var testOrg *models.Organization
	for _, org := range organizations {
		if org.Name == "Test organization" {
			testOrg = &org
		}
	}
	if testOrg != nil {
		log.Printf("test organization already exists, skip inserting the rest of the seed data")
		return
	}

	org, err := r.CreateOrganization(context.TODO(), models.CreateOrganizationInput{
		Name:         "Test organization",
		DatabaseName: "test_1",
	})
	if err != nil {
		log.Printf("error creating organization: %v", err)
	}

	///////////////////////////////
	// Tokens
	///////////////////////////////

	token, err := r.CreateToken(context.TODO(), CreateToken{
		OrgID: org.ID,
		Token: "token12345",
	})
	if err != nil {
		log.Printf("error creating token: %v", err)
	}

	///////////////////////////////
	// Create and store a data model
	///////////////////////////////
	r.CreateDataModel(context.TODO(), org.ID, app.DataModel{
		Tables: map[app.TableName]app.Table{
			"transactions": {
				Name: "transactions",
				Fields: map[app.FieldName]app.Field{
					"object_id": {
						DataType: app.String,
					},
					"updated_at":  {DataType: app.Timestamp},
					"account_id":  {DataType: app.String},
					"bic_country": {DataType: app.String},
					"country":     {DataType: app.String},
					"description": {DataType: app.String},
					"direction":   {DataType: app.String},
					"status":      {DataType: app.String},
					"title":       {DataType: app.String},
					"amount":      {DataType: app.Float},
				},
				LinksToSingle: map[app.LinkName]app.LinkToSingle{
					"accounts": {
						LinkedTableName: "accounts",
						ParentFieldName: "object_id",
						ChildFieldName:  "account_id"},
				},
			},
			"accounts": {
				Name: "accounts",
				Fields: map[app.FieldName]app.Field{
					"object_id": {
						DataType: app.String,
					},
					"updated_at": {DataType: app.Timestamp},
					"balance":    {DataType: app.Float},
					"company_id": {DataType: app.String},
					"name":       {DataType: app.String},
					"currency":   {DataType: app.String},
					"is_frozen":  {DataType: app.Bool},
				},
			},
			"companies": {
				Name: "companies",
				Fields: map[app.FieldName]app.Field{
					"object_id": {
						DataType: app.String,
					},
					"updated_at": {DataType: app.Timestamp},
					"name":       {DataType: app.String},
				},
			},
		},
	})
	///////////////////////////////
	// Create and store a scenario
	///////////////////////////////
	createScenarioInput := app.CreateScenarioInput{
		Name:              "test name",
		Description:       "test description",
		TriggerObjectType: "transactions",
	}
	scenario, err := r.CreateScenario(context.TODO(), org.ID, createScenarioInput)
	if err != nil {
		log.Printf("error creating scenario: %v", err)
	}

	createScenarioIterationInput := app.CreateScenarioIterationInput{
		ScenarioID: scenario.ID,
		Body: &app.CreateScenarioIterationBody{
			TriggerCondition:     &operators.True{},
			ScoreReviewThreshold: utils.Ptr(10),
			ScoreRejectThreshold: utils.Ptr(30),
			Rules: []app.CreateRuleInput{
				{
					Formula:       &operators.True{},
					ScoreModifier: 2,
					Name:          "Rule 1 Name",
					Description:   "Rule 1 Desc",
				},
				{
					Formula:       &operators.False{},
					ScoreModifier: 2,
					Name:          "Rule 2 Name",
					Description:   "Rule 2 Desc",
				},
				{
					Formula:       &operators.EqBool{Left: &operators.True{}, Right: &operators.True{}},
					ScoreModifier: 2,
					Name:          "Rule 3 Name",
					Description:   "Rule 3 Desc",
				},
				{
					Formula:       &operators.EqBool{Left: &operators.True{}, Right: &operators.EqBool{Left: &operators.False{}, Right: &operators.False{}}},
					ScoreModifier: 2,
					Name:          "Rule 4 Name",
					Description:   "Rule 4 Desc",
				},
			},
		},
	}

	scenarioIteration, err := r.CreateScenarioIteration(context.TODO(), org.ID, createScenarioIterationInput)
	if err != nil {
		log.Printf("error creating scenario iteration: %v", err)
	}
	_, err = r.CreateScenarioPublication(context.TODO(), org.ID, app.CreateScenarioPublicationInput{
		ScenarioIterationID: scenarioIteration.ID,
		PublicationAction:   app.Publish,
	})
	if err != nil {
		log.Printf("error publishing scenario iteration: %v", err)
	}

	///////////////////////////////
	// Also create the demo scenario
	///////////////////////////////
	demoScenario, err := r.CreateScenario(context.TODO(), org.ID, app.CreateScenarioInput{
		Name:              "Demo scenario",
		Description:       "Demo scenario",
		TriggerObjectType: "transactions",
	})
	if err != nil {
		log.Printf("error creating demo scenario: %v", err)
	}

	createDemoScenarioIterationInput := app.CreateScenarioIterationInput{
		ScenarioID: demoScenario.ID,
		Body: &app.CreateScenarioIterationBody{
			TriggerCondition: &operators.And{
				Operands: []operators.OperatorBool{
					&operators.EqString{
						Left:  &operators.PayloadFieldString{FieldName: "direction"},
						Right: &operators.StringValue{Text: "payout"},
					},
					&operators.EqString{
						Left:  &operators.PayloadFieldString{FieldName: "status"},
						Right: &operators.StringValue{Text: "pending"},
					},
				},
			},
			ScoreReviewThreshold: utils.Ptr(20),
			ScoreRejectThreshold: utils.Ptr(30),
			Rules: []app.CreateRuleInput{
				{
					Formula: &operators.And{
						Operands: []operators.OperatorBool{
							&operators.GreaterOrEqualFloat{
								Left:  &operators.PayloadFieldFloat{FieldName: "amount"},
								Right: &operators.FloatValue{Value: 10000},
							},
							&operators.LesserFloat{
								Left:  &operators.PayloadFieldFloat{FieldName: "amount"},
								Right: &operators.FloatValue{Value: 100000},
							},
						},
					},
					ScoreModifier: 10,
					Name:          "Medium amount",
				},
				{
					Formula: &operators.GreaterOrEqualFloat{
						Left:  &operators.PayloadFieldFloat{FieldName: "amount"},
						Right: &operators.FloatValue{Value: 100000},
					},
					ScoreModifier: 20,
					Name:          "High amount",
				},
				{
					Formula: &operators.StringIsInList{
						Str: &operators.PayloadFieldString{FieldName: "bic_country"},
						List: &operators.StringListValue{
							Strings: []string{"HU", "IT", "PO", "IR"},
						},
					},
					ScoreModifier: 10,
					Name:          "Medium risk country",
				},
				{
					Formula: &operators.StringIsInList{
						Str: &operators.PayloadFieldString{FieldName: "bic_country"},
						List: &operators.StringListValue{
							Strings: []string{"RO", "RU", "LT"},
						},
					},
					ScoreModifier: 20,
					Name:          "High risk country",
				},
				{
					Formula: &operators.EqString{
						Left: &operators.PayloadFieldString{FieldName: "bic_country"},
						Right: &operators.StringValue{
							Text: "FR",
						},
					},
					ScoreModifier: -10,
					Name:          "Low risk country",
				},
				{
					Formula: &operators.StringIsInList{
						Str: &operators.PayloadFieldString{FieldName: "bic"},
						List: &operators.StringListValue{
							Strings: []string{"FRTRZOFRPP", "FPPRPFFXXX"},
						},
					},
					ScoreModifier: 10,
					Name:          "High risk BIC",
				},
				{
					Formula: &operators.DbFieldBool{
						FieldName:        "is_frozen",
						TriggerTableName: "transactions",
						Path:             []string{"account"},
					},
					ScoreModifier: 100,
					Name:          "Frozen account",
				},
			},
		},
	}
	demoScenarioIteration, err := r.CreateScenarioIteration(context.TODO(), org.ID, createDemoScenarioIterationInput)
	if err != nil {
		log.Printf("error creating demo scenario iteration: %v", err)
	}
	_, err = r.CreateScenarioPublication(context.TODO(), org.ID, app.CreateScenarioPublicationInput{
		ScenarioIterationID: demoScenarioIteration.ID,
		PublicationAction:   app.Publish,
	})
	if err != nil {
		log.Printf("error publishing demo scenario iteration: %v", err)
	}

	log.Println("")
	log.Println("Finish to Seed the DB :")
	log.Printf("\t- $SCENARIO_ID: %s\n", scenario.ID)
	log.Printf("\t- $TOKEN: %s\n", token.Token)
	log.Println("")
}
