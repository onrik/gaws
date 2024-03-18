package tests

type ExtensionsStruct struct {
	ID int `openapiExt:"x-test-ext=test,x-test-ext2=1"`
}
