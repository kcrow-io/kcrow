package k8s

import (
	"os"
)

var (
	nodeName = os.Getenv("NODE_NAME")
)

type Event int

const (
	AddEvent = iota
	UpdateEvent
	DeleteEvent
)

type Register interface {
	Name() string
}

type NodeRegister interface {
	Register
	NodeUpdate(*NodeItem)
}

type NamespaceRegister interface {
	Register
	NamespaceUpdate(*NsItem)
}

type RuntimeRegister interface {
	Register
	RuntimeUpdate(*RuntimeItem)
}
