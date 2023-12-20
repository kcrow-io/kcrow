package util

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/hashset"
)

var (
	randseed = rand.NewSource(37)
)

type CallbackFn func(any)

type Event int

const (
	UpdateE Event = iota
	DeleteE
	CreateE
)

func Ipv4Family() *int64 {
	var v int64 = 4
	return &v
}

func ClearMap(m map[string]struct{}) {
	for k := range m {
		delete(m, k)
	}
}

func SetAddList(s *hashset.Set, l []string) {
	if s == nil || len(l) == 0 {
		return
	}
	for _, v := range l {
		s.Add(v)
	}
}

func IterMap(s *hashmap.Map, fn func(k, v interface{}) error) {
	if s == nil || fn == nil {
		return
	}
	var val interface{}
	keys := s.Keys()
	for _, k := range keys {
		v, ok := s.Get(k)
		if !ok {
			val = nil
		} else {
			val = v
		}
		err := fn(k, val)
		if err != nil {
			return
		}
	}
}

func Rand(n int) string {
	return fmt.Sprintf("%v", time.Now().Unix())
}
