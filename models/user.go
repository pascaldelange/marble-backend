package models

type UserId string

type User struct {
	UserId         UserId
	Email          string
	FirebaseUid    string
	Role           Role
	OrganizationId string
}

type CreateUser struct {
	Email          string
	Role           Role
	OrganizationId string
}
