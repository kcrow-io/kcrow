package tools

import (
	_ "github.com/gogo/protobuf/gogoproto"              // nolint Used for protobuf generation of pkg/k8s/types/slim/k8s
	_ "golang.org/x/tools/cmd/goimports"                // nolint
	_ "k8s.io/code-generator"                           // nolint
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" // nolint
)
