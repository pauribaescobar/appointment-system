package models

type Appointment struct {
	ID                         string
	CenterID                   string
	customerPhoneNumber        string
	confirmationExpirationDate int64
}
