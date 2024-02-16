package tests

import (
	json2 "encoding/json"
	"time"
)

type User4 struct {
	ID        int
	Name      string
	CreatedAt *time.Time
	Groups    []struct {
		Name  string
		Data  json2.RawMessage
		Admin *struct {
			Name      string
			something struct {
				something bool
			}
		}
	}
	secret string
}
