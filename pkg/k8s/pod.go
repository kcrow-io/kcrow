package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type PodManage struct {
	ctx context.Context

	reader cache.Cache
}

func NewPodControl(ctx context.Context, reader cache.Cache) *PodManage {
	if reader == nil {
		panic(fmt.Errorf("reader cannot be nil"))
	}
	nr := &PodManage{
		ctx:    ctx,
		reader: reader,
	}
	if err := nr.probe(); err != nil {
		panic(err)
	}
	return nr
}

func (nr *PodManage) Pod(nsname types.NamespacedName) (*corev1.Pod, error) {
	po := &corev1.Pod{}
	err := nr.reader.Get(nr.ctx, nsname, po)
	if err != nil {
		return nil, err
	}
	return po, nil
}

func (nr *PodManage) probe() error {
	var (
		po = &corev1.Pod{}
	)
	informer, err := nr.reader.GetInformer(nr.ctx, po)
	if err != nil {
		return err
	}
	if nodeName == "" {
		klog.Warningf("recommand set environment '%s', otherwise all pod will cache", nodeNameEnv)
	}
	evHandler := toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			v, ok := obj.(*corev1.Pod)
			if !ok {
				return false
			}

			if nodeName != "" && v.Spec.NodeName != nodeName {
				return false
			}

			return true
		},
		Handler: &nullHandler{},
	}

	_, err = informer.AddEventHandler(evHandler)
	return err
}

// try to analyze the bellow situations
// *.[kind] | [kind] | [container_name].[kind] | [container_index].[kind]
func TryParseContainer(po *corev1.Pod, cntname string, s string) (string, bool) {
	slist := strings.Split(s, ".")
	switch len(slist) {
	case 1:
		return s, true
	case 2:
		switch slist[0] {
		case "*":
			return s, true
		default:
			num, err := strconv.Atoi(s)
			if err != nil {
				return slist[1], s == cntname
			}
			for i := range po.Spec.Containers {
				if i != num {
					continue
				}
				if po.Spec.Containers[i].Name == cntname {
					return slist[1], true
				}
			}
			return slist[1], false
		}

	default:
		return s, false
	}
}

func TransPod(in interface{}) (out interface{}, err error) {
	v, ok := in.(*corev1.Pod)
	if ok {
		v.ManagedFields = nil
		v.Status = corev1.PodStatus{}
		return v, nil
	}
	return nil, fmt.Errorf("not pod type")
}
