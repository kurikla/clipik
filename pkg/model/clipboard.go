package model

import "github.com/google/uuid"

type Clipboard struct {
	ID        uuid.UUID
	Mime      string
	Data      []byte
	CreatedAt string
}
