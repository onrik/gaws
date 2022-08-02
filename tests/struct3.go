package tests

import (
	"encoding/json"
	json2 "encoding/json"
	"time"
)

type Type struct {
	Name string `json:"name"`
}

type Group2 struct {
	Name      string   `json:"name"`
	Paths     []string `json:"paths"`
	Admin     *User2   `json:"admin"`
	GroupType Type     `json:"groupType"`
	Data      []byte
}

type User2 struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Email       string           `json:"email"`
	IsAdmin     bool             `json:"is_admin"`
	Groups      []Group2         `json:"groups"`
	CreatedAt   time.Time        `json:"created_at"`
	AccountType Type             `json:"accountType"`
	UserType    Type             `json:"userType"`
	Manager     *User2           `json:"manager"`
	Json        json.RawMessage  `json:"json"`
	Json2       json2.RawMessage `json:"json2"`
	Bites       []byte           `json:"bites"`
}

/*
Users list
@openapi GET /i/v1/users
@openapiRequest multipart/form-data {"file": []byte}
@openapiResponse 200 application/json {"users": []User2, "user": User2}
@openapiSecurity api_key apiKey cookie AuthKey
*/
func Users() error {
	return nil
}
