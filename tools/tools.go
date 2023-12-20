package tools

import (
	_ "github.com/gogo/protobuf/gogoproto" // Used for protobuf generation of pkg/k8s/types/slim/k8s
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
