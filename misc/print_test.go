package misc

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type testStruct struct {
	A string
	B string
}

func TestFlatten(t *testing.T) {
	assert.Equal(t,
		[]string{"test1\ta,b", "test2\t"},
		flatten(map[string][]string{"test1": []string{"a", "b"}, "test2": []string{""}}),
	)
	assert.True(t,
		reflect.DeepEqual([]string{"A\t1", "B\t2"},
			flatten(&testStruct{A: "1", B: "2"})))

}
