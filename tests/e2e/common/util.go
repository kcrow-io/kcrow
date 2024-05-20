package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"time"

	"github.com/containerd/cgroups/v3"
	"github.com/kcrow-io/kcrow/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/prometheus/procfs"
)

var (
	defaultCgroup = "/sys/fs/cgroup"

	bgctx = context.Background()
)

const (
	unlimited = "unlimited"
)

func BashImage() string {
	return fmt.Sprintf("%s/library/bash:5", DockerMirror)
}

// fake namespace
func FakeNamespace(name string, annotation map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotation,
		},
	}
}

// fake pod
func FakePod(name string, namespace string, annotation map[string]string) *corev1.Pod {
	var ns string = namespace
	if namespace == "" {
		ns = "default"
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: annotation,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    name,
					Image:   BashImage(),
					Command: []string{"sleep", "1d"},
				},
			},
		},
	}
}

// NOTICE. get cpuset in current node.
func HostCpuSet() string {
	var cpus string
	if cgroups.Mode() == cgroups.Unified {
		// cgroup2
		cpus = path.Join(defaultCgroup, "cpuset.cpus.effective")
	} else {
		cpus = path.Join(defaultCgroup, "cpuset", "cpuset.effective_cpus")
	}
	data, err := os.ReadFile(cpus)
	if err != nil {
		panic(err)
	}

	return string(bytes.TrimSuffix(data, []byte("\n")))
}

// NOTICE. get CoreSize in current node.
func CoreSizeCommand(cmd string) (string, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return "", err
	}
	procs, err := fs.AllProcs()
	if err != nil {
		return "", err
	}
	var limitstr string
	for _, p := range procs {
		c, err := p.Comm()
		if err != nil {
			continue
		}
		if cmd == c {
			limit, err := p.Limits()
			if err != nil {
				return "", err
			}
			if limit.CoreFileSize == math.MaxUint64 {
				limitstr = unlimited
			} else {
				limitstr = fmt.Sprintf("%d", limit.CoreFileSize)
			}
			return limitstr, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func waitPodPahseRunning(cli k8sclient, nsname types.NamespacedName) (container string, err error) {
	var (
		pod = &corev1.Pod{}
	)
	err = util.TimeBackoff(func() error {
		err := cli.Get(bgctx, nsname, pod)
		if err != nil {
			return err
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			container = pod.Spec.Containers[0].Name
			return nil
		default:
			return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
		}
	}, time.Second*10)
	return
}

// exec command in default container
func PodExec(ctx context.Context, cli k8sclient, nsname types.NamespacedName, cmd []string) ([]byte, error) {
	var (
		stdout = util.GetBuf()
		stderr = util.GetBuf()
	)
	defer util.PutBuf(stdout)
	defer util.PutBuf(stderr)

	_, err := waitPodPahseRunning(cli, nsname)
	if err != nil {
		log.Default().Println("wait failed: ", err)
		return nil, err
	}
	req := cli.CoreV1().RESTClient().Post().Resource("pods").Name(nsname.Name).
		Namespace(nsname.Namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdout:  true,
		Stderr:  true,
	}
	req.VersionedParams(option, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(cli.RestConfig(), "POST", req.URL())
	if err != nil {
		log.Default().Println("spdy exector failed: ", err)
		return nil, err
	}
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		log.Default().Println("stream failed: ", err)
		return nil, err
	}
	if stderr.Len() != 0 {
		return stdout.Bytes(), fmt.Errorf("stderr: %s", stderr.String())
	}

	return stdout.Bytes(), nil
}
