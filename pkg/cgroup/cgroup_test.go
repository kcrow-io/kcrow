package cgroup

import (
	"testing"

	"github.com/containerd/nri/pkg/api"
)

func ptrUint(i uint64) *api.OptionalUInt64 {
	return &api.OptionalUInt64{
		Value: i,
	}
}

func ptrInt(i int64) *api.OptionalInt64 {
	return &api.OptionalInt64{
		Value: i,
	}
}

func TestCgroupMerge(t *testing.T) {
	tests := []struct {
		name     string
		src, dst any
		override bool
		want     any
		wantErr  bool
	}{
		{
			name:     "merge cpu",
			src:      &api.LinuxCPU{Shares: ptrUint(10)},
			dst:      &api.LinuxCPU{Shares: ptrUint(20)},
			override: true,
			want:     &api.LinuxCPU{Shares: ptrUint(10)},
		},
		{
			name:     "merge memory",
			src:      &api.LinuxMemory{Limit: ptrInt(100)},
			dst:      &api.LinuxMemory{Limit: ptrInt(200)},
			override: true,
			want:     &api.LinuxMemory{Limit: ptrInt(100)},
		},
		{
			name:     "merge nil src",
			src:      nil,
			dst:      &api.LinuxCPU{Shares: ptrUint(20)},
			override: true,
			wantErr:  true,
		},
		{
			name:     "merge nil dst",
			src:      &api.LinuxCPU{Shares: ptrUint(10)},
			dst:      nil,
			override: true,
			wantErr:  true,
		},
		{
			name:     "merge different type",
			src:      &api.LinuxCPU{Shares: ptrUint(10)},
			dst:      &api.LinuxMemory{Limit: ptrInt(200)},
			override: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cgroupMerge(tt.src, tt.dst, tt.override); (err != nil) != tt.wantErr {
				t.Errorf("CgroupMerge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
