## introduce

- This is an introduction to RLIMIT configurations and a brief explanation. For more details, please refer to [man7.org](https://man7.org/linux/man-pages/man2/getrlimit.2.html).

- The purpose of the introduction is to facilitate understanding of how configuring these parameters at the tenant or container level can better control behavior.

### useage

- Below are common configuration information, slightly different depending on the system configuration.

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

### resource
- RLIMIT_AS

    This is the maximum size of the process's virtual memory
    (address space).  The limit is specified in bytes, and is
    rounded down to the system page size.  This limit affects
    calls to brk(2), mmap(2), and mremap(2), which fail with
    the error ENOMEM upon exceeding this limit.  In addition,
    automatic stack expansion fails (and generates a SIGSEGV
    that kills the process if no alternate stack has been made
    available via sigaltstack(2)).  Since the value is a long,
    on machines with a 32-bit long either this limit is at
    most 2 GiB, or this resource is unlimited.

- RLIMIT_CORE

    This is the maximum size of a core file (see core(5)) in
    bytes that the process may dump.  When 0 no core dump
    files are created.  When nonzero, larger dumps are
    truncated to this size.

- RLIMIT_CPU

    This is a limit, in seconds, on the amount of CPU time
    that the process can consume.  When the process reaches
    the soft limit, it is sent a SIGXCPU signal.  The default
    action for this signal is to terminate the process.
    However, the signal can be caught, and the handler can
    return control to the main program.  If the process
    continues to consume CPU time, it will be sent SIGXCPU
    once per second until the hard limit is reached, at which
    time it is sent SIGKILL.  (This latter point describes
    Linux behavior.  Implementations vary in how they treat
    processes which continue to consume CPU time after
    reaching the soft limit.  Portable applications that need
    to catch this signal should perform an orderly termination
    upon first receipt of SIGXCPU.)

- RLIMIT_DATA

    This is the maximum size of the process's data segment
    (initialized data, uninitialized data, and heap).  The
    limit is specified in bytes, and is rounded down to the
    system page size.  This limit affects calls to brk(2),
    sbrk(2), and (since Linux 4.7) mmap(2), which fail with
    the error ENOMEM upon encountering the soft limit of this
    resource.

- RLIMIT_FSIZE

    This is the maximum size in bytes of files that the
    process may create.  Attempts to extend a file beyond this
    limit result in delivery of a SIGXFSZ signal.  By default,
    this signal terminates a process, but a process can catch
    this signal instead, in which case the relevant system
    call (e.g., write(2), truncate(2)) fails with the error
    EFBIG.

- RLIMIT_LOCKS (Linux 2.4.0 to Linux 2.4.24)

    This is a limit on the combined number of flock(2) locks
    and fcntl(2) leases that this process may establish.

- RLIMIT_MEMLOCK

    This is the maximum number of bytes of memory that may be
    locked into RAM.  This limit is in effect rounded down to
    the nearest multiple of the system page size.  This limit
    affects mlock(2), mlockall(2), and the mmap(2) MAP_LOCKED
    operation.  Since Linux 2.6.9, it also affects the
    shmctl(2) SHM_LOCK operation, where it sets a maximum on
    the total bytes in shared memory segments (see shmget(2))
    that may be locked by the real user ID of the calling
    process.  The shmctl(2) SHM_LOCK locks are accounted for
    separately from the per-process memory locks established
    by mlock(2), mlockall(2), and mmap(2) MAP_LOCKED; a
    process can lock bytes up to this limit in each of these
    two categories.

    Before Linux 2.6.9, this limit controlled the amount of
    memory that could be locked by a privileged process.
    Since Linux 2.6.9, no limits are placed on the amount of
    memory that a privileged process may lock, and this limit
    instead governs the amount of memory that an unprivileged
    process may lock.

- RLIMIT_MSGQUEUE (since Linux 2.6.8)

    This is a limit on the number of bytes that can be
    allocated for POSIX message queues for the real user ID of
    the calling process.  This limit is enforced for
    mq_open(3).  Each message queue that the user creates
    counts (until it is removed) against this limit according
    to the formula:

        Since Linux 3.5:

            bytes = attr.mq_maxmsg * sizeof(struct msg_msg) +
                    MIN(attr.mq_maxmsg, MQ_PRIO_MAX) *
                        sizeof(struct posix_msg_tree_node)+
                                    /* For overhead */
                    attr.mq_maxmsg * attr.mq_msgsize;
                                    /* For message data */

        Linux 3.4 and earlier:

            bytes = attr.mq_maxmsg * sizeof(struct msg_msg *) +
                                    /* For overhead */
                    attr.mq_maxmsg * attr.mq_msgsize;
                                    /* For message data */

    where attr is the mq_attr structure specified as the
    fourth argument to mq_open(3), and the msg_msg and
    posix_msg_tree_node structures are kernel-internal
    structures.

    The "overhead" addend in the formula accounts for overhead
    bytes required by the implementation and ensures that the
    user cannot create an unlimited number of zero-length
    messages (such messages nevertheless each consume some
    system memory for bookkeeping overhead).

- RLIMIT_NICE (since Linux 2.6.12, but see BUGS below)

    This specifies a ceiling to which the process's nice value
    can be raised using setpriority(2) or nice(2).  The actual
    ceiling for the nice value is calculated as 20 - rlim_cur.
    The useful range for this limit is thus from 1
    (corresponding to a nice value of 19) to 40 (corresponding
    to a nice value of -20).  This unusual choice of range was
    necessary because negative numbers cannot be specified as
    resource limit values, since they typically have special
    meanings.  For example, RLIM_INFINITY typically is the
    same as -1.  For more detail on the nice value, see
    sched(7).

- RLIMIT_NOFILE

    This specifies a value one greater than the maximum file
    descriptor number that can be opened by this process.
    Attempts (open(2), pipe(2), dup(2), etc.)  to exceed this
    limit yield the error EMFILE.  (Historically, this limit
    was named - RLIMIT_OFILE on BSD.)

    Since Linux 4.5, this limit also defines the maximum
    number of file descriptors that an unprivileged process
    (one without the CAP_SYS_RESOURCE capability) may have "in
    flight" to other processes, by being passed across UNIX
    domain sockets.  This limit applies to the sendmsg(2)
    system call.  For further details, see unix(7).

- RLIMIT_NPROC

    This is a limit on the number of extant process (or, more
    precisely on Linux, threads) for the real user ID of the
    calling process.  So long as the current number of
    processes belonging to this process's real user ID is
    greater than or equal to this limit, fork(2) fails with
    the error EAGAIN.

    The - RLIMIT_NPROC limit is not enforced for processes that
    have either the CAP_SYS_ADMIN or the CAP_SYS_RESOURCE
    capability, or run with real user ID 0.

- RLIMIT_RSS

    This is a limit (in bytes) on the process's resident set
    (the number of virtual pages resident in RAM).  This limit
    has effect only in Linux 2.4.x, x < 30, and there affects
    only calls to madvise(2) specifying MADV_WILLNEED.

- RLIMIT_RTPRIO (since Linux 2.6.12, but see BUGS)

    This specifies a ceiling on the real-time priority that
    may be set for this process using sched_setscheduler(2)
    and sched_setparam(2).

    For further details on real-time scheduling policies, see
    sched(7)

- RLIMIT_RTTIME (since Linux 2.6.25)

    This is a limit (in microseconds) on the amount of CPU
    time that a process scheduled under a real-time scheduling
    policy may consume without making a blocking system call.
    For the purpose of this limit, each time a process makes a
    blocking system call, the count of its consumed CPU time
    is reset to zero.  The CPU time count is not reset if the
    process continues trying to use the CPU but is preempted,
    its time slice expires, or it calls sched_yield(2).

    Upon reaching the soft limit, the process is sent a
    SIGXCPU signal.  If the process catches or ignores this
    signal and continues consuming CPU time, then SIGXCPU will
    be generated once each second until the hard limit is
    reached, at which point the process is sent a SIGKILL
    signal.

    The intended use of this limit is to stop a runaway real-
    time process from locking up the system.

    For further details on real-time scheduling policies, see
    sched(7)

- RLIMIT_SIGPENDING (since Linux 2.6.8)

    This is a limit on the number of signals that may be
    queued for the real user ID of the calling process.  Both
    standard and real-time signals are counted for the purpose
    of checking this limit.  However, the limit is enforced
    only for sigqueue(3); it is always possible to use kill(2)
    to queue one instance of any of the signals that are not
    already queued to the process.

- RLIMIT_STACK

    This is the maximum size of the process stack, in bytes.
    Upon reaching this limit, a SIGSEGV signal is generated.
    To handle this signal, a process must employ an alternate
    signal stack (sigaltstack(2)).

    Since Linux 2.6.23, this limit also determines the amount
    of space used for the process's command-line arguments and
    environment variables; for details, see execve(2).