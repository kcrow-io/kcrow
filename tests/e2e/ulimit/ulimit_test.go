package ulimit_test

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
	noAnno bool

	annotations map[string]string
}

type podExpect struct {
	po types.NamespacedName
	// expect core-filesize
	coreSizeExpect string
}

func coreSizeAnnotations(v string) map[string]string {
	if v == "" {
		return nil
	}
	return map[string]string{
		"core.rlimit.kcrow.io": v,
	}
}

var _ = Describe("Namespace rlimit", func() {
	var (
		nscore     = "1"
		pocore     = "2"
		namespaces = map[string]testItem{
			"norlimit": {
				noAnno:      true,
				annotations: nil,
			},
			"rlimit": {
				noAnno:      false,
				annotations: coreSizeAnnotations(nscore),
			},
		}
		expects []podExpect

		hostCoreSize string
		err          error
	)

	// create namespace before
	BeforeSuite(func() {
		for k, v := range namespaces {
			Expect(k8cli.CreateOrUpdate(bgctx, common.FakeNamespace(k, v.annotations))).To(BeNil())
		}
		var hostcmd = "containerd"
		hostCoreSize, err = common.CoreSizeCommand(hostcmd)
		fmt.Fprintf(GinkgoWriter, "hostCommand '%s' coresize: %s\n", hostcmd, hostCoreSize)
		Expect(err).To(BeNil())
	})

	Context("rlimit annotation namespace", func() {
		// create no annotations pod in has rlimit annotation namespace.
		// expect pod injected by namespace annotations.
		It("no rlimit annotation pod", func() {
			for k, v := range namespaces {
				if v.noAnno {
					continue
				}
				poname := fmt.Sprintf("no%s", k)
				po := common.FakePod(poname, k, nil)
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:             types.NamespacedName{Namespace: k, Name: poname},
					coreSizeExpect: nscore,
				})
			}
		})

		// create annotations pod in has rlimit annotation namespace.
		// expect pod injected by pod annotations
		It("rlimit annotation pod", func() {

			for k, v := range namespaces {
				if v.noAnno {
					continue
				}
				name := fmt.Sprintf("rl%s", k)
				po := common.FakePod(name, k, coreSizeAnnotations(pocore))
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:             types.NamespacedName{Namespace: k, Name: name},
					coreSizeExpect: pocore,
				})
			}
		})
	})

	Context("no rlimit annotation namespace", func() {
		// create no annotations pod in has rlimit annotation namespace.
		// expect pod equal host cpuset.
		It("no rlimit annotation pod", func() {
			for k, v := range namespaces {
				if !v.noAnno {
					continue
				}
				name := fmt.Sprintf("no%s", k)
				po := common.FakePod(name, k, nil)
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:             types.NamespacedName{Namespace: k, Name: name},
					coreSizeExpect: hostCoreSize,
				})
			}
		})

		// create annotations pod in has rlimit annotation namespace.
		// expect pod injected by pod annotations
		It("rlimit annotation pod", func() {
			for k, v := range namespaces {
				if !v.noAnno {
					continue
				}
				name := fmt.Sprintf("rl%s", k)
				po := common.FakePod(name, k, coreSizeAnnotations(pocore))
				Expect(k8cli.CreateOrUpdate(bgctx, po)).To(BeNil())
				expects = append(expects, podExpect{
					po:             types.NamespacedName{Namespace: k, Name: name},
					coreSizeExpect: pocore,
				})
			}
		})
	})

	AfterSuite(func() {
		ctx, cancle := context.WithCancel(context.Background())
		commands := []string{
			"sh", "-c",
			"cat /proc/1/limits | awk  '/core file size/ {print $5}'",
		}
		for _, v := range expects {
			out, err := common.PodExec(ctx, *k8cli, v.po, commands)
			out = bytes.TrimSuffix(out, []byte("\n"))
			fmt.Fprintf(GinkgoWriter, "pod '%s' coresize: %s, expect: %s\n", v.po, string(out), v.coreSizeExpect)
			Expect(err).To(BeNil())
			Expect(string(out)).To(BeEquivalentTo(v.coreSizeExpect))
		}
		cancle()

		for k, v := range namespaces {
			Expect(k8cli.Delete(bgctx, common.FakeNamespace(k, v.annotations))).To(BeNil())
		}
	})
})
