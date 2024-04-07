# 介绍

- 当前有大量的 AI 设备用在 kubernetes 平台中，设备管理通常使用方式为 插件方式，具体更多插件的介绍见[这里](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#grpc-endpoint-getallocatableresources)

- 但同样 kubernetes 平台中也不局限于一种运行时，但通常目前提供的 AI 设备的运行时都是基于默认 oci runtime 开发的，这对其他运行时很不友好

- 为解决更多运行时问题，同时也希望能统一管理这类设备，因为加入 xpu 支持


## gpu

- 当前以 nvidia gpu 进行支持，并且支持的模式也是 [legacy](https://github.com/NVIDIA/k8s-device-plugin?tab=readme-ov-file#as-command-line-flags-or-envvars) 方式，也是最常见的通过 envvar 方式传递设备信息

## npu

- 当前以 [ascend npu]((https://gitee.com/ascend/ascend-device-plugin)) 进行支持