// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package cgroup

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReliability(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cgroup Suite")
}

var _ = BeforeSuite(func() {

})
