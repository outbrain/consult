package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	A string
	B string
	C []string
}

func TestStructToString(t *testing.T) {
	t.Parallel()
	s := testStruct{"a", "b", []string{"c1", "c2"}}
	assert.Equal(t, "a\tb\tc1,c2", StructToString(s))

	assert.Equal(t, "a\tb\tc1,c2", StructToString(&s))
}
