## 介绍

- 主要介绍 cgroup v1 的相关配置，对于掌握 cgroup 是一个不错的开始

- 因为只在一篇中介绍，相对是比较简陋，如要查阅更详细的，可以参考 [cgroupv1](https://www.kernel.org/doc/Documentation/cgroup-v1/)

## blockio

- cgroup 子系统“blkio”实现块 io 控制器，可用于指定设备上的 IO 速率上限。可以实施于通用块层，可用于叶节点以及更高层级别逻辑设备，如设备映射器

- 依赖 kernel

    - CONFIG_BLK_CGROUP=y  启用块 IO 控制器
    - CONFIG_BLK_DEV_THROTTLING=y  在块层启用节流

- 设置
    - 指定特定设备上的带宽速率，格式 “<major>:<minor> <bytes_per_second>”，当 bytes_per_second 为 0 时，清除限制
    - 例子
        - echo “8:16 1048576”> /sys/fs/cgroup/blkio/blkio.throttle.read_bps_device

- 可限制的内容

    - blkio.throttle.read_bps_device
    - blkio.throttle.write_bps_device
    - blkio.throttle.read_iops_device
    - blkio.throttle.write_iops_device

- 关于
    - iops：IOPS (Input/Output Per Second)即每秒的读写次数，指的是系统在单位时间内能处理的最大的I/O频度，I/O请求通常为读或写数据操作请求。对于随机读写频繁的应用，如OLTP(Online Transaction Processing)，IOPS是关键衡量指标
    - throughput：单位时间内最大的I/O流量；一些大量的顺序文件访问


## cpuset

