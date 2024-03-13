
# 功能

## multi-tenant

- 简化租户资源隔离配置，

- 使用 kubernetes node/namespace 资源绑定租户，通过设置 node/namespace 完成同租户下资源隔离


## Cpu/Memory cgroup

- kubernetes pod 可以配置cpu资源当前是  request 和 limit，这较 cpu cgoup 能力弱，参考[具体 cpu cgoup](https://www.kernel.org/doc/Documentation/cgroup-v1/)

- 支持多租户设置

## ulimit

- kubernetes pod 不支持配置 ulimit 等资源，通过配置注解支持

- 支持多租户设置

## NPU/GPU runtime 

- NPU/GPU 等设备在运行业务容器时，需要添加额外 hook 或映射 device 动作，虽然当前有 *-docker-runtime 支持，但未考虑其他运行时，如 kata, kuasar 等

- 实现 NPU/GPU runtime 逻辑，完成统一runtime能力

## NPU/GPU topology 

- 不同 NPU/GPU 在不同 zone，其网络能力差距巨大或者微小，如统一到同 zone 下，有效提升工作效率
