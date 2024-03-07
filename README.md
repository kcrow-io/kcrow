# kcrow

[![Go Report Card](https://goreportcard.com/badge/github.com/yylt/kcrow)](https://goreportcard.com/report/github.com/yylt/kcrow)
[![CodeFactor](https://www.codefactor.io/repository/github/yylt/kcrow/badge)](https://www.codefactor.io/repository/github/yylt/kcrowl)
[![codecov](https://codecov.io/gh/yylt/kcrow/branch/main/graph/badge.svg?token=YKXY2E4Q8G)](https://codecov.io/gh/yylt/kcrow)

**English** | [**简体中文**](./README-zh.md)

## Overview
This project is developed based on the NRI interface and is used to implement multi-tenant resource control. The resources include ulimits, CPU, and memory cgroup settings, more detail about [nri](https://github.com/containerd/nri).

Kcrow supports adding annotations to pod, node and namespaces to complete the configuration. 

The annotation configuration priorities are **pod > node > namespace**.

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
git clone https://github.com/yylt/kcrow/
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
