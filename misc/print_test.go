package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testInnerStruct struct {
	S string
}

type testStruct struct {
	A string
	B string
	C []string
	D testInnerStruct
	E *testInnerStruct
}

func TestStructToString(t *testing.T) {
	t.Parallel()
	s := testStruct{"a", "b", []string{"c1", "c2"}, testInnerStruct{"d"}, &testInnerStruct{"e"}}
	assert.Equal(t, "\"a\"\t\"b\"\t\"c1,c2\"\t\"d\"\t\"e\"", StructToString(s))

	assert.Equal(t, "\"a\"\t\"b\"\t\"c1,c2\"\t\"d\"\t\"e\"", StructToString(&s))
}

func TestStructHeaderLine(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "A\tB\tC\tD.S\tE.S", StructHeaderLine(testStruct{}))
}
