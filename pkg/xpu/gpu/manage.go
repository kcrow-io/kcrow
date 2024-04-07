package gpu

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"github.com/kcrow-io/kcrow/pkg/util"
	"k8s.io/klog/v2"
)

var (
	name          = "nvidiagpu"
	annotationKey = "nvidia.gpu.kcrow.io"
)

type gpuPath struct {
	HookPath string `json:"hookpath" yaml:"hookpath"`
	LibPath  string `json:"libpath" yaml:"libpath"`
}

type Gpu struct {
	rmctl *k8s.RuntimeManage
	// runtime - value
	runtime map[string]*gpuPath

	mu sync.RWMutex
}

func New(rm *k8s.RuntimeManage) *Gpu {
	gpu := &Gpu{
		runtime: map[string]*gpuPath{},
		rmctl:   rm,
	}
	rm.Registe(gpu)
	return gpu
}

func (g *Gpu) Name() string {
	return name
}

func (g *Gpu) RuntimeUpdate(ri *k8s.RuntimeItem) {
	if ri == nil {
		return
	}
	var p = &gpuPath{}

	for k, v := range ri.No.Annotations {
		if strings.ToLower(k) == annotationKey {
			err := json.Unmarshal([]byte(v), p)
			if err != nil {
				klog.Warningf("unmarshal runtime %s annotation %s failed: %v", ri.No.Name, annotationKey, err)
			} else {
				g.mu.Lock()
				g.runtime[ri.No.Name] = p
				g.mu.Unlock()
				return
			}
		}
	}
}

// TODO
func (g *Gpu) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Ct == nil {
		klog.Warningf("not found container info")
		return nil
	}
	var (
		ct = im.Ct
	)

	visibleDevices := util.GetValueFromEnvByKey(ct.Env, visibleDevicesEnvvar)
	if visibleDevices == "" {
		klog.V(2).Infof("no env %s found", visibleDevicesEnvvar)
		return nil
	}
	// TODO support more runtime
	klog.Infof("process nvidiagpu device, in vm runtime: %v", g.rmctl.Isvm(im.Sb.RuntimeHandler))
	return nil

}
