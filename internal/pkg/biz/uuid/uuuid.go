package uuid

import uuid "github.com/satori/go.uuid"

func GetUUID() string {
	uuid := uuid.NewV4()
	return uuid.String()
}
