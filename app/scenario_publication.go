package app

import "time"

type PublicationAction int

const (
	Publish PublicationAction = iota
	Unpublish
	UnknownPublicationAction
)

// Provide a string value for each publication action
func (o PublicationAction) String() string {
	switch o {
	case Publish:
		return "publish"
	case Unpublish:
		return "unpublish"
	}
	return "unknown"
}

// Provide an PublicationAction from a string value
func PublicationActionFrom(s string) PublicationAction {
	switch s {
	case "publish":
		return Publish
	case "unpublish":
		return Unpublish
	case "unknown":
		return UnknownPublicationAction
	}
	return UnknownPublicationAction
}

type ScenarioPublication struct {
	ID    string
	OrgID string
	// UserID              string
	ScenarioID          string
	ScenarioIterationID string
	PublicationAction   PublicationAction
	CreatedAt           time.Time
}

type CreateScenarioPublication struct {
	// UserID              string
	ScenarioID          string
	ScenarioIterationID string
	PublicationAction   PublicationAction
}
