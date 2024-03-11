package tests

type UUIDUser struct {
	Name        string `json:"name"`
	ID          string `json:"id" openapi:"required,format=uuid"`
	Group       string `json:"group" openapi:"required,default=user" openapiEnum:"admin,manager,user"`
	Description string `json:"description" openapi:"example=testExample" openapiDesc:"testDescription"`
}
