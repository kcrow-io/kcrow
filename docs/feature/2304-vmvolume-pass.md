# vm network storage passthrough

## Summary

在 Guest VM(如 Kata) 内挂载逻辑卷时，默认使用 Virtiofs 文件系统方式，该方式对 NAS 存储性能上有很大影响

对于 NAS 存储，可以通过在 Guest 中进行挂载操作绕过 virtiofs 的读写



## Goals

- 提升持久卷在虚拟机内的读写性能提升

## Non-Goals

- TODO

## Proposal

### User Stories

- [] 识别是否为使用虚拟化的容器

- [] 识别持久卷的挂载类型及其选项，在支持范围内或者否

- [] 删除原有的持久卷挂载信息从 spec.mounts，并改为 spec.hooks 内命令
