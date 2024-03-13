# Features

## Multi-tenant

- Simplify tenant resource isolation configuration

- Use Kubernetes node/namespace resources to bind tenants, and achieve resource isolation within the same tenant by setting node/namespace

## CPU/Memory cgroup

- Kubernetes pods can currently configure CPU resources with requests and limits, which are weaker than CPU cgroup capabilities. Refer to [specific CPU cgroup](https://www.kernel.org/doc/Documentation/cgroup-v1/)

- Support multi-tenant settings

## Ulimit

- Kubernetes pods do not support configuring ulimit and other resources. Support through configuring annotations

- Support multi-tenant settings

## NPU/GPU runtime

- When running business containers with NPU/GPU and other devices, additional hooks or device mapping actions need to be added, although currently *-docker-runtime is supported, but other runtimes like kata, kuasar, etc. are not considered

- Implement NPU/GPU runtime logic to achieve unified runtime capabilities

## NPU/GPU topology

- Different NPU/GPUs in different zones have vastly different or slightly different network capabilities. By consolidating them into the same zone, work efficiency can be effectively improved