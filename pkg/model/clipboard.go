package model

import "time"

type Clipboard struct {
	ID        int64
	Mime      string
	Data      []byte
	CreatedAt time.Time
}
