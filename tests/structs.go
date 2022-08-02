package tests

import (
	json2 "encoding/json"
	"time"
)

type Group struct {
	Name  string
	Data  json2.RawMessage
	Admin User
}

type User struct {
	ID        int
	Name      string
	CreatedAt *time.Time
	Groups    []Group
	secret    string
}
