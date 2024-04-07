# Introduction

- Currently, there are a large number of AI devices used in the Kubernetes platform, and device management is generally done in a plugin manner. For more details on specific plugins, see [here](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#grpc-endpoint-getallocatableresources).

- However, the Kubernetes platform is also not limited to one runtime, but currently the provided runtimes for AI devices are generally developed based on the default OCI runtime, which is not very friendly to other runtimes

- To solve more runtime problems, and also to hope to manage such devices in a unified way, so xpu support is added.

## GPU

- Currently supports NVIDIA GPUs, and the supported mode is also the [legacy mode](https://github.com/NVIDIA/k8s-device-plugin?tab=readme-ov-file#as-command-line-flags-or-envvars), which is the most common way of passing device information through envvars

## NPU

- Currently supports [Ascend NPU](https://gitee.com/ascend/ascend-device-plugin).