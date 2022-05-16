package accounts

import "github.com/google/uuid"

type Account struct {
	ID   uuid.UUID
	Name string
}
