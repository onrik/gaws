package tests

import "github.com/onrik/gaws/tests/nested/nestedv2/v2"

type VersionStruct struct {
	ID       int
	NestedV2 nestedv2.NestedStruct
}
