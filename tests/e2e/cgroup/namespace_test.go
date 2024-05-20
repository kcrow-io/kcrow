package cgroup

import (
	"bytes"
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kcrow-io/kcrow/tests/e2e/common"
)

var (
	k8cli = common.MustGetClient()
	bgctx = context.Background()
)

type testItem struct {
	nocgAnno bool

	annotations map[string]string
}

type podExpect struct {
	po types.NamespacedName
	// expect cpuset
	cpusetExpect string
}

func cpusetAnnotations(v string) map[string]string {
	if v == "" {
		return nil
	}
	return map[string]string{
		"cpu.cgroup.kcrow.io": fmt.Sprintf("{\"cpus\":\"%s\"}", v),
	}
}

var _ = Describe("Namespace cgroup", func() {
	var (
		nscpus     = "0"
		namespaces = map[string]testItem{
			"nocgroup": {
				nocgAnno:    true,
				annotations: nil,
			},
			"cgroup": {
				nocgAnno:    false,
				annotations: cpusetAnnotations(nscpus),
			},
		}
		expects []podExpect
	)

	// create namespace before
	BeforeSuite(func() {
		for k, v := range namespaces {
			Expect(k8cli.CreateOrUpdate(bgctx, common.FakeNamespace(k, v.annotations))).To(BeNil())
		}
	})

	Context("cgroup annotation namespace", func() {
		// create no annotations pod in has cgroup annotation namespace.
		// expect pod injected by namespace annotations.
		It("no cgroup annotation pod", func() {
			for k, v := range namespaces {
				if v.nocgAnno {
					continue
				}
				poname := fmt.Sprintf("nocg%s", k)
				po := common.FakePod(poname, k, nil)
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:           types.NamespacedName{Namespace: k, Name: poname},
					cpusetExpect: nscpus,
				})
			}
		})

		// create annotations pod in has cgroup annotation namespace.
		// expect pod injected by pod annotations
		It("cgroup annotation pod", func() {
			var pocpus = "1"
			for k, v := range namespaces {
				if v.nocgAnno {
					continue
				}
				name := fmt.Sprintf("cg%s", k)
				po := common.FakePod(name, k, cpusetAnnotations(pocpus))
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:           types.NamespacedName{Namespace: k, Name: name},
					cpusetExpect: pocpus,
				})
			}
		})
	})

	Context("no cgroup annotation namespace", func() {
		// create no annotations pod in has cgroup annotation namespace.
		// expect pod equal host cpuset.
		It("no cgroup annotation pod", func() {
			for k, v := range namespaces {
				if !v.nocgAnno {
					continue
				}
				name := fmt.Sprintf("nocg%s", k)
				po := common.FakePod(name, k, nil)
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:           types.NamespacedName{Namespace: k, Name: name},
					cpusetExpect: common.HostCpuSet(),
				})
			}
		})

		// create annotations pod in has cgroup annotation namespace.
		// expect pod injected by pod annotations
		It("cgroup annotation pod", func() {
			var pocpus = "1"
			for k, v := range namespaces {
				if !v.nocgAnno {
					continue
				}
				name := fmt.Sprintf("cg%s", k)
				po := common.FakePod(name, k, cpusetAnnotations(pocpus))
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:           types.NamespacedName{Namespace: k, Name: name},
					cpusetExpect: pocpus,
				})
			}
		})
	})

	AfterSuite(func() {
		ctx, cancle := context.WithCancel(context.Background())
		commands := []string{
			"sh", "-c",
			"cat /sys/fs/cgroup/cpuset.cpus.effective 2>/dev/null || cat /sys/fs/cgroup/cpuset/cpuset.effective_cpus 2>/dev/null",
		}
		for _, v := range expects {
			out, err := common.PodExec(ctx, *k8cli, v.po, commands)
			out = bytes.TrimSuffix(out, []byte("\n"))
			fmt.Fprintf(GinkgoWriter, "pod '%s' cpuset: %s, expect: %s\n", v.po, string(out), v.cpusetExpect)
			Expect(err).To(BeNil())
			Expect(string(out)).To(BeEquivalentTo(v.cpusetExpect))
		}
		cancle()
		for k, v := range namespaces {
			Expect(k8cli.Delete(bgctx, common.FakeNamespace(k, v.annotations))).To(BeNil())
		}
	})
})
