# kcrow

[![Go Report Card](https://goreportcard.com/badge/github.com/kcrow-io/kcrow)](https://goreportcard.com/report/github.com/kcrow-io/kcrow)
[![CodeFactor](https://www.codefactor.io/repository/github/kcrow-io/kcrow/badge)](https://www.codefactor.io/repository/github/kcrow-io/kcrow)
[![codecov](https://codecov.io/gh/kcrow-io/kcrow/branch/main/graph/badge.svg?token=YKXY2E4Q8G)](https://codecov.io/gh/kcrow-io/kcrow)


[**English**](./README.md) | **简体中文**


## 介绍

kcrow 主要是完成多租户的资源管理，以及设备和运行时相关初始化功能。当前已有的能力如下

- 支持控制 ulimit 和 cpu/memory cgroup 资源

- 支持配置资源注解在命名空间，节点和容器多个级别上

- 支持优先级，当前资源设置优先级为 pod > 节点 > 命名空间

当遇到是否配置以及如何配置关于 cgroup 和 ulimit 信息， 可以参考 [cgroup](./docs/cgroup.md) 和 [ulimit](./docs/ulimit.md)，之后会通过更多的例子做说明

## Roadmap

| 功能                              | 状态  |
|----------------------------------|----------|
| Multi-tenant                  | Alpha    |
| Cpu cgroup                    | Alpha    |
| Memory cgroup                    | Alpha     |
| Ulimit                    | Alpha     |
| NPU/GPU runtime                    | In-plan     |
| NPU/GPU topology                    | In-plan     |

关于详细功能的规划，具体可参考 [roadmap](./docs/develop/roadmap-zh.md)。


## 应用场景

- 多租户计算资源隔离，通过 cgroup 或 ulimit 方式控制

- 中间件、数据存储、日志观测、AI 训练等网络 I/O 密集性应用，支持自定义 ulimit 配额

- AI 基础平台提升 NPU/GPU 调度和运行能力

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
git clone https://github.com/kcrow-io/kcrow/
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


