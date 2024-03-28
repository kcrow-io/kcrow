## 介绍

- 这篇是介绍rlimit 有关的配置及其简易说明，更详细的请参考 [man7](https://man7.org/linux/man-pages/man2/getrlimit.2.html)

- 介绍的目的是方便了解当我们可以针对租户或容器级别配置这些参数时，会方便更好的控制行为’.


### 使用

- 以下是常见的配置信息，根据系统配置略有不同

```
$ ulimit -a
core file size              (blocks, -c) 0
data seg size               (kbytes, -d) unlimited
scheduling priority                 (-e) 0
file size                   (blocks, -f) unlimited
pending signals                     (-i) 62181
max locked memory           (kbytes, -l) 1999620
max memory size             (kbytes, -m) unlimited
open files                          (-n) 1024
pipe size                (512 bytes, -p) 8
POSIX message queues         (bytes, -q) 819200
real-time priority                  (-r) 0
stack size                  (kbytes, -s) 8192
cpu time                   (seconds, -t) unlimited
max user processes                  (-u) 62181
virtual memory              (kbytes, -v) unlimited
file locks                          (-x) unlimited
```

### 资源


- RLIMIT_AS

    这是进程虚拟内存（地址空间）的最大大小。限制以字节为单位指定，并按照系统页面大小向下取整。此限制影响对brk(2)、mmap(2)和mremap(2)的调用，超过此限制会导致ENOMEM错误。此外，自动堆栈扩展失败（如果未通过sigaltstack(2)提供备用堆栈，则会生成一个导致进程退出的SIGSEGV）。由于值是一个长整型，在拥有32位长整型的计算机上，此限制最大为2 GiB，或者此资源无限制。

- RLIMIT_CORE

    这是进程可以转储的核心文件（参见core(5)）的最大大小，以字节为单位。当为0时，不会创建核心转储文件。非零值时，较大的转储文件会被截断为此大小。

- RLIMIT_CPU

    这是进程可以消耗的CPU时间的限制，以秒为单位。当进程达到软限制时，会收到SIGXCPU信号。该信号的默认操作是终止进程。然而，该信号可以被捕获，处理程序可以将控制返回到主程序。如果进程继续消耗CPU时间，每秒将发送一个SIGXCPU，直到达到硬限制，此时会发送一个SIGKILL。（后者描述了Linux的行为，实现在处理继续消耗CPU时间的进程方面存在差异。需要捕获此信号的可移植应用程序应在首次收到SIGXCPU时执行有序终止。）

- RLIMIT_DATA

    这是进程数据段（初始化数据、未初始化数据和堆）的最大大小。限制以字节为单位指定，并按照系统页面大小向下取整。此限制影响对brk(2)、sbrk(2)和（自Linux 4.7起）mmap(2)的调用，遇到此资源的软限制时将返回ENOMEM错误。

- RLIMIT_FSIZE

    这是进程可以创建的文件的最大大小，以字节为单位。尝试将文件扩展超出此限制将导致传递SIGXFSZ信号。默认情况下，此信号终止进程，但进程可以捕获此信号，此时相关的系统调用（例如write(2)、truncate(2)）将返回EFBIG错误。

- RLIMIT_LOCKS（Linux 2.4.0至Linux 2.4.24）

    这是进程可以建立的flock(2)锁和fcntl(2)租约的组合数量的限制。

- RLIMIT_MEMLOCK

    这是可以锁定到RAM的内存字节数的最大限制。此限制实际上会按照系统页面大小向下取整。此限制影响mlock(2)、mlockall(2)和mmap(2)的MAP_LOCKED操作。自Linux 2.6.9以来，它还影响shmctl(2)的SHM_LOCK操作，在这里它设置了由调用进程的真实用户ID锁定的共享内存段中的字节的最大值（参见shmget(2)）。shmctl(2)的SHM_LOCK锁定是分别计算的，独立于由mlock(2)、mlockall(2)和mmap(2)的MAP_LOCKED建立的进程内存锁定；一个进程可以在这两种类别中的每一个中锁定到达这个限制的字节。

    在Linux 2.6.9之前，此限制控制特权进程可以锁定的内存量。自Linux 2.6.9以来，不会对特权进程可以锁定的内存量施加限制，而此限制改  

- RLIMIT_MSGQUEUE (自Linux 2.6.8)

    这是对调用进程的真实用户ID分配的POSIX消息队列的字节数的限制。此限强制应用于mq_open(3)。用户创建的每个消息队列都按照以下公式计算（直至删除为止）对此限的计数:


```
  自Linux 3.5：

      bytes = attr.mq_maxmsg * sizeof(struct msg_msg) +
              MIN(attr.mq_maxmsg, MQ_PRIO_MAX) *
                  sizeof(struct posix_msg_tree_node)+
                              /* 欠额 */
              attr.mq_maxmsg * attr.mq_msgsize;
                              /* 消息数据 */

  Linux 3.4及更早版本：

      bytes = attr.mq_maxmsg * sizeof(struct msg_msg *) +
                              /* 欠额 */
              attr.mq_maxmsg * attr.mq_msgsize;
                              /* 消息数据 */
```

    这里，attr是作为mq_open(3)的第四个参数指定的mq_attr结构，msg_msg和posix_msg_tree_node结构是内核内部结构。
    公式中的“欠额”项考虑了实现所需的额外字节，并确保用户无法创建无限数量的零长度消息（尽管这些消息每个仍然会消耗系统内存用于簿记欠额）。

- RLIMIT_NICE (自Linux 2.6.12，但请参见下面的BUGS)

    这指定了使用setpriority(2)或nice(2)可以将进程的nice值提高到的最大值。实际的nice值上限计算为20 - rlim_cur。因此，此限的有用范围为从1（对应于nice值19）到40（对应于nice值-20）。选择此不寻常范围是必要的，因为无法将负数指定为资源限制值，因为通常具有特殊含义。例如，RLIM_INFINITY通常与-1相同。有关nice值的更多细节，请参见sched(7)。  


- RLIMIT_NOFILE

    指定了一个比该进程可以打开的最大文件描述符号大的值。试图超出此限制的尝试（open(2)、pipe(2)、dup(2)等）将产生EMFILE错误。（在历史上，此限制称为- BSD上的RLIMIT_OFILE)。

    自Linux 4.5以来，此限制还定义了未特权进程（没有CAP_SYS_RESOURCE功能）可能“传输”到其他进程的最大文件描述符的数量，通过通过UNIX域套接字传递。此限制适用于sendmsg(2)系统调用。有关详细信息，请参阅unix(7)。

- RLIMIT_NPROC

    这是对调用进程的真实用户ID的现有进程（或更确切地说，在Linux上，线程）数量的限制。只要属于此进程的真实用户ID的当前进程数大于或等于此限制，fork(2)将返回错误EAGAIN。

    - RLIMIT_NPROC限制对具有CAP_SYS_ADMIN或CAP_SYS_RESOURCE功能或以真实用户ID 0运行的进程不强制执行。

- RLIMIT_RSS

    这是对进程常驻集的限制（常驻RAM中的虚拟页面数），以字节为单位计数。此限制仅在Linux 2.4.x，x < 30中有效，在那儿只对指定MADV_WILLNEED的madvise(2)调用产生影响。  


- RLIMIT_RTPRIO（自Linux 2.6.12，但参见BUGS）

    这指定了一个上限，即可以使用sched_setscheduler(2)和sched_setparam(2)为该进程设置的实时优先级。

    有关实时调度策略的进一步详细信息，请参见sched7)。

- RLIMIT_RTTIME（自Linux 2.6.25）

    这是对采用实时调度策略的进程在不进行阻塞系统调用的情况下可以消耗的CPU时间量的限制（以微秒为单位）。对于此限制，每次进程进行阻塞系统调用时，其消耗的CPU时间计数会被重置为零。如果进程继续尝试使用CPU但被抢占，其时间片到期或调用了sched_yield(2)，则不会重置CPU时间计数。

    达到软限制时，进程会收到一个SIGXCPU信号。如果进程捕获或忽略此信号并继续消耗CPU时间，则SIGXCPU将每秒生成一次，直到达到硬限制，此时进程将收到一个SIGKILL信号。

    此限制的预期用途是防止失控的实时进程锁定系统。

    有关实时调度策略的进一步详细信息，请参见sched(7)。  


- RLIMIT_SIGPENDING (自Linux 2.6.8)

    这是对调用进程的真实用户ID的挂起信号数量的限制。为此限制目的，标准信号和实时信号都会被计算在内。但是，此限制仅对sigqueue(3)生效；这意味着可以始终使用kill(2)将未排队到进程的任何信号的一个实例排队。

- RLIMIT_STACK

    这是进程堆栈的最大大小，以字节为单位计数。达到此限制时，会生成一个SIGSEGV信号。要处理此信号，进程必须使用替代信号栈(sigaltstack(2))。

    自Linux 2.6.23以来，此限制还确定了用于进程命令行参数和环境变量的空间量；有关详细信息，请参见execve(2)。  
