package stringers

import "github.com/google/uuid"

func UUIDEmpty(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}
