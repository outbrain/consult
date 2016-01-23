package misc

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type testStruct struct {
	A string
	B string
}

func TestFlatten(t *testing.T) {
	var res []string

	res = flatten(map[string][]string{"test1": []string{"a", "b"}, "test2": []string{""}})
	sort.Strings(res)
	assert.Equal(t, []string{"test1\ta,b", "test2\t"}, res)

	res = flatten(&testStruct{A: "1", B: "2"})
	sort.Strings(res)
	assert.Equal(t, []string{"A\t1", "B\t2"}, res)
}
