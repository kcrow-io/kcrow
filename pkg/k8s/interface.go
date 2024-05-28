package k8s

import (
	"os"
)

const (
	nodeNameEnv = "NODE_NAME"

	nameRuntimeAnnotationKey = "name.vm.kcrow.io"
)

var (
	nodeName = os.Getenv(nodeNameEnv)
)

type Event int

const (
	AddEvent = iota
	UpdateEvent
	DeleteEvent
)

type runtimeName string

const (
	kataRuntimeName runtimeName = "kata"
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

type VolumeRegister interface {
	Register
	VolumeUpdate(*VolumeItem)
}

type nullHandler struct{}

func (t *nullHandler) OnAdd(obj interface{}, isInInitialList bool) {}

func (t *nullHandler) OnUpdate(oldObj, newObj interface{}) {}

func (t *nullHandler) OnDelete(obj interface{}) {}
