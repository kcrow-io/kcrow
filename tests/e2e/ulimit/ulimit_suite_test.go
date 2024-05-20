package ulimit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUlimit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ulimit Suite")
}
