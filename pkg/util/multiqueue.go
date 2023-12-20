package util

import (
	sync "github.com/yylt/kcrow/pkg/lock"
)

// 有名链表队列，可以指定队列名称
type MultiQueue[T any] struct {
	mu     sync.RWMutex
	queues map[string]*List[T]
}

func NewMultiQueue[T any]() *MultiQueue[T] {
	pc := &MultiQueue[T]{
		queues: map[string]*List[T]{},
	}
	return pc
}

// 压入队列，如queue指定，则压入指定队列中
// 未指定queue时，压入全局队列中
func (p *MultiQueue[T]) Push(queue string, value T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.queues[queue]
	if ok {
		v.Append(value)
	} else {
		p.queues[queue] = New[T](value)
	}
}

// 弹出可用的数值，如queue指定，则尝试从queue中弹出，若queue为空
// 则将从全局队列弹出
func (p *MultiQueue[T]) Pop(queue string) (t T, exist bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.queues[queue]
	if ok {
		return v.Pop()
	}
	return
}

// 迭代所有有名链表，注意使用的是原始数据
func (p *MultiQueue[T]) Iter(fn func(string, []T) error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for q, l := range p.queues {
		err := fn(q, l.Values())
		if err != nil {
			return
		}
	}
}
