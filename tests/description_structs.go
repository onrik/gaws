package tests

type DescriptionStruct struct {
	_  struct{} `json:"-" openapiDesc:"Test description"`
	ID int
}
