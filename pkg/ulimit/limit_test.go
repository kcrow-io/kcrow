package ulimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvrlimit(t *testing.T) {
	var num uint64 = 1024
	tests := []struct {
		value string
		want  rlimit
	}{
		{"1024", rlimit{Hard: &num, Soft: &num}},
		{"{\"hard\": 1024, \"soft\": 1024}", rlimit{Hard: &num, Soft: &num}},
	}
	got := &rlimit{}
	for _, test := range tests {
		err := resolvRlimit(test.value, got)
		if err != nil {
			t.Fatalf("resolvrlimit(%s) failed: %v", test.value, err)
		}

		assert.Equal(t, test.want, *got)
	}

	badValue := "foo"
	err := resolvRlimit(badValue, &rlimit{})
	if err == nil {
		t.Fatalf("resolvrlimit(%s) should fail", badValue)
	}
}
