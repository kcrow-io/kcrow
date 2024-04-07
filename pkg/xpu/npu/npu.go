package npu

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/k8s"
	"github.com/kcrow-io/kcrow/pkg/oci"
	"github.com/kcrow-io/kcrow/pkg/util"
	"github.com/kcrow-io/kcrow/pkg/xpu/npu/dcmi"
	"k8s.io/klog/v2"
)

var (
	assendRe = regexp.MustCompile(`^Ascend(910|310|310B|310P)-(\d+)$`)

	name          = "ascendnpu"
	annotationKey = "ascend.npu.kcrow.io"
)

type npuPath struct {
	HookPath    string `json:"hookpath" yaml:"hookpath"`
	DestoryPath string `json:"destorypath" yaml:"destorypath"`
}

type Npu struct {
	rmctl *k8s.RuntimeManage
	// runtime - value
	runtime map[string]*npuPath

	mu sync.RWMutex
}

func NewNpu(rm *k8s.RuntimeManage) *Npu {
	npu := &Npu{
		runtime: map[string]*npuPath{},
		rmctl:   rm,
	}
	rm.Registe(npu)
	return npu
}

func (n *Npu) Name() string {
	return name
}

func (n *Npu) RuntimeUpdate(ri *k8s.RuntimeItem) {
	if ri == nil {
		return
	}
	var p = &npuPath{}

	for k, v := range ri.No.Annotations {
		if strings.ToLower(k) == annotationKey {
			err := json.Unmarshal([]byte(v), p)
			if err != nil {
				klog.Warningf("unmarshal runtime %s annotation %s failed: %v", ri.No.Name, annotationKey, err)
			} else {
				n.mu.Lock()
				n.runtime[ri.No.Name] = p
				n.mu.Unlock()
				return
			}
		}
	}
}

func (n *Npu) Process(ctx context.Context, im *oci.Item) error {
	if im == nil || im.Ct == nil {
		klog.Warningf("not found container info")
		return nil
	}
	var (
		ct = im.Ct
	)

	visibleDevices := util.GetValueFromEnvByKey(ct.Env, ascendVisibleDevices)
	if visibleDevices == "" {
		klog.V(2).Infof("no env %s found", ascendVisibleDevices)
		return nil
	}
	// TODO support more runtime
	klog.Infof("process %s device, in vm runtime: %v", name, n.rmctl.Isvm(im.Sb.RuntimeHandler))

	devices, err := parseDevices(visibleDevices)
	if err != nil {
		klog.Errorf("parse ascend device failed: %v", err)
		return err
	}
	klog.Infof("ascend devices info %v, start inject hook and devices", devices)
	if len(devices) != 0 {
		if idlist, err := addHook(im, devices); err != nil {
			klog.Errorf("failed to inject hook, err: %v", err)
			return fmt.Errorf("failed to inject hook, err: %v", err)
		} else {
			if idlist != nil {
				devices = idlist
			}
		}
		if err = addDevice(im, devices); err != nil {
			klog.Errorf("failed to add device, err: %v", err)
			return fmt.Errorf("failed to add device to env: %v", err)
		}
	}
	return nil
}

func addDevice(im *oci.Item, deviceidlist []int) error {
	deviceName := davinciName
	if strings.Contains(util.GetValueFromEnvByKey(im.Ct.Env, ascendRuntimeOptions), "VIRTUAL") {
		deviceName = virtualDavinciName
	}
	for _, deviceId := range deviceidlist {
		dPath := devicePath + deviceName + strconv.Itoa(deviceId)
		if err := addDeviceToSpec(im.Adjust, dPath, deviceName); err != nil {
			return fmt.Errorf("failed to add davinci device to spec: %v", err)
		}
	}

	if err := addManagerDevice(im.Adjust); err != nil {
		return fmt.Errorf("failed to add Manager device to spec: %v", err)
	}

	return nil
}

func addHook(im *oci.Item, deviceidlist []int) ([]int, error) {
	var (
		apihook   *api.Hooks = im.Adjust.Hooks
		oldenv               = im.Ct.Env
		newIdList []int
	)
	if im.Adjust.Hooks == nil {
		im.Adjust.Hooks = &api.Hooks{}
		apihook = im.Adjust.Hooks
	}

	needUpdate := true
	if len(apihook.CreateRuntime) > maxCommandLength {
		return nil, fmt.Errorf("too many items in Prestart ")
	}
	for _, hook := range apihook.CreateRuntime {
		if strings.Contains(hook.Path, hookCli) {
			needUpdate = false
			break
		}
	}
	if needUpdate {
		apihook.CreateRuntime = append(apihook.CreateRuntime, &api.Hook{
			Path: hookCli,
			Args: []string{hookCli},
		})
	}

	if len(oldenv) > maxCommandLength {
		return nil, fmt.Errorf("too many items in Env ")
	}

	// check virtual device or not.
	if strings.Contains(util.GetValueFromEnvByKey(im.Ct.Env, ascendRuntimeOptions), "VIRTUAL") {
		return nil, nil
	}

	vdevice, err := dcmi.CreateVDevice(&dcmi.NpuWorker{}, oldenv, deviceidlist)
	if err != nil {
		return nil, err
	}
	klog.Infof("vnpu split done: vdevice: %v", vdevice.VdeviceID)

	if vdevice.VdeviceID != -1 {
		newIdList = []int{int(vdevice.VdeviceID)}
		updateEnvAndPostHook(im, vdevice)
	}

	return newIdList, nil
}

func parseDevices(visibleDevices string) ([]int, error) {
	var (
		devicesList = strings.Split(visibleDevices, ",")
		devices     []int
		err         error
	)

	if strings.Contains(visibleDevices, ascend) {
		devices, err = parseAscendDevices(devicesList)
		if err != nil {
			return nil, err
		}
	} else {
		for _, d := range devicesList {
			d = strings.TrimSpace(d)
			if strings.Contains(d, "-") {
				borders := strings.Split(d, "-")
				if len(borders) != borderNum {
					return nil, fmt.Errorf("invalid device range: %s", d)
				}

				borders[0] = strings.TrimSpace(borders[0])
				borders[1] = strings.TrimSpace(borders[1])

				left, err := strconv.Atoi(borders[0])
				if err != nil || left < 0 {
					return nil, fmt.Errorf("invalid left boarder range parameter: %s", borders[0])
				}

				right, err := strconv.Atoi(borders[1])
				if err != nil || right > maxDevice {
					return nil, fmt.Errorf("invalid right boarder range parameter: %s", borders[1])
				}

				if left > right {
					return nil, fmt.Errorf("left boarder (%d) should not be larger than the right one(%d)", left, right)
				}

				for n := left; n <= right; n++ {
					devices = append(devices, n)
				}
			} else {
				n, err := strconv.Atoi(d)
				if err != nil {
					return nil, fmt.Errorf("invalid single device parameter: %s", d)
				}

				devices = append(devices, n)
			}
		}
	}
	sort.SliceStable(devices, func(i, j int) bool { return i < j })
	return removeDuplication(devices), nil
}

func parseAscendDevices(devicesList []string) ([]int, error) {
	devices := make([]int, 0, len(devicesList))
	chipType := ""

	chipName, err := dcmi.GetChipName()
	if err != nil {
		return nil, fmt.Errorf("get chip name error: %v", err)
	}
	for _, d := range devicesList {
		matchGroups := assendRe.FindStringSubmatch(strings.TrimSpace(d))
		if matchGroups == nil {
			return nil, fmt.Errorf("invalid device format: %s", d)
		}
		n, err := strconv.Atoi(matchGroups[2])
		if err != nil {
			return nil, fmt.Errorf("invalid device id: %s", d)
		}

		if chipType == "" {
			chipType = matchGroups[1]
		}
		if chipType != "" && chipType != matchGroups[1] {
			return nil, fmt.Errorf("invalid device chip type: %s", d)
		}

		devices = append(devices, n)
	}

	if ascend+chipType != getDeviceTypeByChipName(chipName) {
		return nil, fmt.Errorf("chip type not match really: %s", chipType)
	}
	return devices, nil
}

func removeDuplication(devices []int) []int {
	list := make([]int, 0, len(devices))
	prev := -1

	for _, device := range devices {
		if device == prev {
			continue
		}

		list = append(list, device)
		prev = device
	}

	return list
}

func updateEnvAndPostHook(it *oci.Item, vdevice dcmi.VDeviceInfo) {
	oldenv := it.Ct.Env
	newEnv := make([]string, 0, len(oldenv)+1)
	needAddVirtualFlag := true

	for _, line := range oldenv {
		words := strings.Split(line, "=")
		if len(words) == envLength && strings.TrimSpace(words[0]) == ascendRuntimeOptions {
			needAddVirtualFlag = false
			if strings.Contains(words[1], "VIRTUAL") {
				newEnv = append(newEnv, line)
				continue
			} else {
				newEnv = append(newEnv, strings.TrimSpace(line)+",VIRTUAL")
				continue
			}
		}
		newEnv = append(newEnv, line)
	}
	if needAddVirtualFlag {
		newEnv = append(newEnv, "ASCEND_RUNTIME_OPTIONS=VIRTUAL")
	}
	it.Ct.Env = newEnv
	it.Adjust.Hooks.Poststart = append(it.Adjust.Hooks.Poststart, &api.Hook{
		Path: destroyHookCli,
		Args: []string{destroyHookCli, fmt.Sprintf("%d", vdevice.CardID), fmt.Sprintf("%d", vdevice.DeviceID),
			fmt.Sprintf("%d", vdevice.VdeviceID)},
	})
}
