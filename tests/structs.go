package tests

import (
	"time"

	json2 "encoding/json"
)

type Group struct {
	Name string
	Data json2.RawMessage
}

type User struct {
	ID        int
	Name      string
	CreatedAt *time.Time
	Groups    []Group
	secret    string
}
