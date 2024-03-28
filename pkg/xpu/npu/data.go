package npu

import "strings"

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
