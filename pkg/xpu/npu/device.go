package npu

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/containerd/containerd/oci"
	"github.com/containerd/nri/pkg/api"
	"github.com/kcrow-io/kcrow/pkg/xpu/npu/dcmi"
	"github.com/opencontainers/runtime-spec/specs-go"
	"k8s.io/klog/v2"
)

const (
	// Atlas200ISoc Product name
	Atlas200ISoc = "Atlas 200I SoC A1"
	// Atlas200 Product name
	Atlas200 = "Atlas 200 Model 3000"
	// Ascend310 ascend 310 chip
	Ascend310 = "Ascend310"
	// Ascend310P ascend 310P chip
	Ascend310P = "Ascend310P"
	// Ascend310B ascend 310B chip
	Ascend310B = "Ascend310B"
	// Ascend910 ascend 910 chip
	Ascend910 = "Ascend910"

	maxDevice = 128

	maxCommandLength = 65535
	hookCli          = "ascend-docker-hook"
	destroyHookCli   = "ascend-docker-destroy"
	dockerRuncFile   = "docker-runc"
	runcFile         = "runc"
	envLength        = 2
	kvPairSize       = 2
	borderNum        = 2

	devicePath           = "/dev/"
	davinciName          = "davinci"
	virtualDavinciName   = "vdavinci"
	davinciManager       = "davinci_manager"
	davinciManagerDocker = "davinci_manager_docker"
	notRenameDeviceType  = ""
	devmmSvm             = "devmm_svm"
	hisiHdc              = "hisi_hdc"
	svm0                 = "svm0"
	tsAisle              = "ts_aisle"
	upgrade              = "upgrade"
	sys                  = "sys"
	vdec                 = "vdec"
	vpc                  = "vpc"
	pngd                 = "pngd"
	venc                 = "venc"
	dvppCmdList          = "dvpp_cmdlist"
	logDrv               = "log_drv"
	acodec               = "acodec"
	ai                   = "ai"
	ao                   = "ao"
	vo                   = "vo"
	hdmi                 = "hdmi"

	ascendVisibleDevices = "ASCEND_VISIBLE_DEVICES"
	ascend               = "Ascend"
	ascendRuntimeOptions = "ASCEND_RUNTIME_OPTIONS"
)

// getDeviceTypeByChipName get device type by chipName
func getDeviceTypeByChipName(chipName string) string {
	if strings.Contains(chipName, "310B") {
		return Ascend310B
	}
	if strings.Contains(chipName, "310P") {
		return Ascend310P
	}
	if strings.Contains(chipName, "310") {
		return Ascend310
	}
	if strings.Contains(chipName, "910") {
		return Ascend910
	}
	return ""
}

func addDeviceToSpec(adj *api.ContainerAdjustment, dPath string, deviceType string) error {
	device, err := oci.DeviceFromPath(dPath)
	if err != nil {
		return fmt.Errorf("failed to get %s info : %#v", dPath, err)
	}

	switch deviceType {
	case virtualDavinciName:
		vDeviceNumber := regexp.MustCompile("[0-9]+").FindAllString(dPath, -1)
		if len(vDeviceNumber) != 1 {
			return fmt.Errorf("invalid vdavinci path: %s", dPath)
		}
		device.Path = devicePath + davinciName + vDeviceNumber[0]
	case davinciManagerDocker:
		device.Path = devicePath + davinciManager
	default:
		// do nothing
	}
	apidevice := api.FromOCILinuxDevices([]specs.LinuxDevice{*device})
	adj.Linux.Devices = append(adj.Linux.Devices, apidevice[0])

	newDeviceCgroup := api.LinuxDeviceCgroup{
		Allow:  true,
		Type:   device.Type,
		Major:  &api.OptionalInt64{Value: device.Major},
		Minor:  &api.OptionalInt64{Value: device.Minor},
		Access: "rwm",
	}
	adj.Linux.Resources.Devices = append(adj.Linux.Resources.Devices, &newDeviceCgroup)
	return nil
}

func addManagerDevice(adj *api.ContainerAdjustment) error {
	chipName, err := dcmi.GetChipName()
	if err != nil {
		return fmt.Errorf("get chip name error: %#v", err)
	}
	devType := getDeviceTypeByChipName(chipName)
	klog.Infof("device type is: %s", devType)
	if devType == Ascend310B {
		return addAscend310BManagerDevice(adj)
	}

	if err := addDeviceToSpec(adj, devicePath+davinciManager, notRenameDeviceType); err != nil {
		return fmt.Errorf("add davinci_manager to spec error: %#v", err)
	}

	productType, err := dcmi.GetProductType(&dcmi.NpuWorker{})
	if err != nil {
		return fmt.Errorf("parse product type error: %#v", err)
	}
	klog.Infof("product type is %s", productType)

	switch productType {
	// do nothing
	case Atlas200ISoc, Atlas200:
	default:
		for _, device := range []string{devmmSvm, hisiHdc} {
			dPath := devicePath + device
			if err := addDeviceToSpec(adj, dPath, notRenameDeviceType); err != nil {
				return fmt.Errorf("failed to add common manage device to spec : %#v", err)
			}
		}
	}

	return nil
}

func addAscend310BManagerDevice(adj *api.ContainerAdjustment) error {
	var Ascend310BManageDevices = []string{
		svm0,
		tsAisle,
		upgrade,
		sys,
		vdec,
		vpc,
		pngd,
		venc,
		dvppCmdList,
		logDrv,
		acodec,
		ai,
		ao,
		vo,
		hdmi,
	}

	for _, device := range Ascend310BManageDevices {
		dPath := devicePath + device
		if err := addDeviceToSpec(adj, dPath, notRenameDeviceType); err != nil {
			klog.Warningf("failed to add %s to spec : %#v", dPath, err)
		}
	}

	davinciManagerPath := devicePath + davinciManagerDocker
	if _, err := os.Stat(davinciManagerPath); err != nil {
		klog.Warningf("failed to get davinci manager docker, err: %#v", err)
		davinciManagerPath = devicePath + davinciManager
		if _, err := os.Stat(davinciManagerPath); err != nil {
			return fmt.Errorf("failed to get davinci manager, err: %#v", err)
		}
	}
	return addDeviceToSpec(adj, davinciManagerPath, davinciManagerDocker)
}
