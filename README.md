# kcrow

[![Go Report Card](https://goreportcard.com/badge/github.com/kcrow-io/kcrow)](https://goreportcard.com/report/github.com/kcrow-io/kcrow)
[![CodeFactor](https://www.codefactor.io/repository/github/kcrow-io/kcrow/badge)](https://www.codefactor.io/repository/github/kcrow-io/kcrow)
[![codecov](https://codecov.io/gh/kcrow-io/kcrow/branch/main/graph/badge.svg?token=YKXY2E4Q8G)](https://codecov.io/gh/kcrow-io/kcrow)

**English** | [**简体中文**](./README-zh.md)

## Overview

kcrow is primarily responsible for multi-tenant resource management, as well as device and runtime-related initialization functions. The current capabilities are as follows:

- Support for controlling ulimit and cpu/memory cgroup resources

- Support for configuring resource annotations at multiple levels, including namespace, node, and container

- Support for priority, with the current resource setting priority being pod > node > namespace



## Roadmap

| Feature                              | Status  |
|----------------------------------|----------|
| Multi-tenant                  | Alpha    |
| Cpu cgroup                    | Alpha    |
| Memory cgroup                    | Alpha     |
| Ulimit                    | Alpha     |
| NPU/GPU runtime                    | In-plan     |
| NPU/GPU topology                    | In-plan     |

Regarding the detailed functional planning, you can refer to the following: [roadmap](./docs/develop/roadmap.md)。


## Scenarios:

- Multi-tenant compute resource isolation, controlled through cgroup or ulimit methods

- Network I/O intensive applications like middleware, data storage, log observability, AI training, etc., supporting customized ulimit quotas

- Enhancing scheduling and runtime capabilities for NPU/GPU in AI base platforms


## Quick Start

### Prerequisites

- containerd version is greater than 1.7.7
- Open and configure nri. Usually, the containerd configuration file is in '/etc/containerd/config.toml'
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
### install

```bash
git clone https://github.com/kcrow-io/kcrow/
helm install charts/kcrowdaemon kcrow -n kcrow  --create-namespace
```

### example

```yaml
// namespace 
apiVersion: v1
kind: Namespace
metadata:
  name: kcrowtest
  annotations:
    nofile.rlimit.kcrow.io: '{"hard":65535,"soft":65535}'
    cpu.cgroup.kcrow.io: '{"cpus":"0-2"}'

// node 
apiVersion: v1
kind: Node
metadata:
  name: node-1
  annotations:
    nofile.rlimit.kcrow.io: '{"hard":65535,"soft":65535}'
    cpu.cgroup.kcrow.io: '{"cpus":"0-2"}'
```

## Contributing
Contributions of code and issues are welcome. Please submit an issue or a pull request.

## License
This project is licensed under the [MIT License](./LICENSE). Please see the [license file](./LICENSE) for more details.

## Contact
If you have any questions or suggestions, please feel free to contact us. You can find us on [GitHub](https://github.com/yylt).
