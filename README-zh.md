# kcrow

[![Go Report Card](https://goreportcard.com/badge/github.com/yylt/kcrow)](https://goreportcard.com/report/github.com/yylt/kcrow)
[![CodeFactor](https://www.codefactor.io/repository/github/yylt/kcrow/badge)](https://www.codefactor.io/repository/github/yylt/kcrow)
[![codecov](https://codecov.io/gh/yylt/kcrow/branch/main/graph/badge.svg?token=YKXY2E4Q8G)](https://codecov.io/gh/yylt/kcrow)

[**English**](./README.md) | **简体中文**


## 概述

该项目基于 NRI 接口开发，用于实现多租户资源控制，资源包括 ulimit、CPU 和内存 cgroup 设置，更多 NRI 参考[这里](https://github.com/containerd/nri)

支持通过对节点和命名空间上添加注解完成配置

当前注解的配置优先级是 **pod > node > namespace**

## 快速开始

### 前提条件

- containerd 版本大于 1.7.7 
- 打开并配置 nri，通常 containerd 配置文件在 '/etc/containerd/config.toml'
```
  [plugins."io.containerd.nri.v1.nri"]
    disable = false
    disable_connections = false
    plugin_config_path = "/etc/nri/conf.d"
    plugin_path = "/opt/nri/plugins"
    plugin_registration_timeout = "5s"
    plugin_request_timeout = "2s"
    socket_path = "/var/run/nri/nri.sock"

```
### 安装

```bash
git clone https://github.com/yylt/kcrow/
helm install charts/kcrowdaemon kcrow -n kcrow  --create-namespace
```

### 使用示例

```yaml
// namespace 示例
apiVersion: v1
kind: Namespace
metadata:
  name: kcrowtest
  annotations:
    nofile.rlimit.kcrow.io: '{"hard":65535,"soft":65535}'
    cpu.cgroup.kcrow.io: '{"cpus":"0-2"}'

// node 示例
apiVersion: v1
kind: Node
metadata:
  name: node-1
  annotations:
    nofile.rlimit.kcrow.io: '{"hard":65535,"soft":65535}'
    cpu.cgroup.kcrow.io: '{"cpus":"0-2"}'
```

## 贡献

欢迎贡献代码和提出问题。请提交 issue 或 PR。

## 版权信息

该项目基于 [MIT 许可证](./LICENSE). 详情请参阅[许可证文件](./LICENSE)。

## 联系我们

如果您有任何疑问或建议，请联系我们。您可以在 [GitHub](https://github.com/yylt) 上找到我们。  


