package nested

import "github.com/onrik/gaws/tests/nested/nested"

type NestedStruct struct {
	ID     int
	Nested nested.NestedStruct
}
