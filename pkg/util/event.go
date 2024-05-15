package util

import (
	"fmt"
	"time"
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

func Rand(n int) string {
	return fmt.Sprintf("%v", time.Now().Unix())
}
