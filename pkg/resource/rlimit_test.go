package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvRlimit(t *testing.T) {
	var num uint64 = 1024
	tests := []struct {
		value string
		want  Rlimit
	}{
		{"1024", Rlimit{Hard: &num, Soft: &num}},
		{"{\"hard\": 1024, \"soft\": 1024}", Rlimit{Hard: &num, Soft: &num}},
	}
	got := &Rlimit{}
	for _, test := range tests {
		err := resolvRlimit(test.value, got)
		if err != nil {
			t.Fatalf("resolvRlimit(%s) failed: %v", test.value, err)
		}

		assert.Equal(t, test.want, *got)
	}

	badValue := "foo"
	err := resolvRlimit(badValue, &Rlimit{})
	if err == nil {
		t.Fatalf("resolvRlimit(%s) should fail", badValue)
	}
}
