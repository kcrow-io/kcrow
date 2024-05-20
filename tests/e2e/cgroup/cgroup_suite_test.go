package cgroup

import (
	"testing"

	_ "github.com/kcrow-io/kcrow/tests/e2e/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCgroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cgroup Suite")
}
