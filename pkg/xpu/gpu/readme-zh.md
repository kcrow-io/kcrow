# nvidia gpu

- 本项目的类似项目是 nvidia-container-toolkit[1]，该项目当前是基于 runc 实现，但实际环境中已经会有多种 runtime，如 kuasar，runhcs，Kata 等，这些很难使用gpu设备，因此为适配更多runtime，有该plugin 出现。

- 当前 nvidia-container-runtime 的内容有包括 legacy，graphics 和 feature 这几种变更器，本项目只专注于 legacy 的变更。在使用本项目前，需安装 gpu-device-plugin[2] 用于设备资源的发现和申请，并映射 env 和 device 信息，架构图如下


# 使用

- 要求安装 driver，参考[这里](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#nvidia-drivers)

- 针对虚拟机 sandbox，在虚拟机镜像内需安装 [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#installation-guide)
    

# nvidia-container-toolkit

## nvidia-container-runtime

- 类似 containerd-shim-runc-v2
- 以 legacy 变更器为例，添加 preStart Hook 执行 nvidia-container-runtime-hook

## nvidia-container-runtime-hook

- TODO

# 参考
[1] nvidia-container-toolkit: https://github.com/NVIDIA/nvidia-container-toolkit
[2] gpu-device-plugin: https://github.com/NVIDIA/k8s-device-plugin
