# Gaws

OpenAPI (swagger) docs generator for Golang.

## Examples

```golang
package users

import (
	"net/http"
	"time"
)

type Group struct {
	Name string `json:"name"`
}

type User struct {
	_         struct{}  `json:"-" openapiDesc:"User"`
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email" openapiDesc:"User's email'"`
	IsAdmin   bool      `json:"is_admin"`
	Groups    []Group   `json:"groups"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status" openapiEnum:"new confirmed deleted"`
}

type createUserRequest struct {
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Password1 string   `json:"password1"`
	Password2 string   `json:"password2"`
	IsAdmin   bool     `json:"is_admin"`
	Groups    []string `json:"groups"`
}

type updateUserRequest struct {
	Email    *string   `json:"email,omitempty"`
	Name     *string   `json:"name,omitempty"`
	Password *string   `json:"password,omitempty"`
	IsAdmin  *bool     `json:"is_admin,omitempty"`
	Groups   *[]string `json:"groups,omitempty"`
}

/*
Users returns users list
@openapi GET /api/v1/users
@openapiParam q in=query, type=string, example=John
@openapiResponse 200 application/json {"users": []User}
*/
func Users(w http.ResponseWriter, r *http.Request) {
}

/*
CreateUser creates user
@openapi POST /api/v1/users
@openapiRequest application/json createUserRequest
@openapiResponse 400 application/json {"message": "email=email;name=required"}
@openapiResponse 200 application/json {"user": User}
*/
func CreateUser(w http.ResponseWriter, r *http.Request) {
}

/*
UpdateUser updates user
@openapi POST /api/v1/users/{id}
@openapiParam id in=path, type=int, example=56
@openapiRequest application/json updateUserRequest
@openapiResponse 404 application/json {"message": "Not Found"}
@openapiResponse 200 application/json {"user": User}
*/
func UpdateUser(w http.ResponseWriter, r *http.Request) {
}

/*
DeleteUser delete user
@openapi DELETE /api/v1/users/{id}
@openapiParam id in=path, type=int, example=56
@openapiResponse 404 application/json {"message": "Not Found"}
@openapiResponse 200 application/json {}
*/
func DeleteUser(w http.ResponseWriter, r *http.Request) {
}

```