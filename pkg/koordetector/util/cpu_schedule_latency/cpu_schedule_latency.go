/*
Copyright 2022 The Koordinator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cpu_schedule_latency

import (
	"fmt"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf ../ebpf/cpu_schedule_latency/csl.bpf.c -- -I../ebpf/headers

type ProgObjects struct {
	Objs        *bpfObjects
	tracepoints []*link.Link
}

func NewCSLeBPFProg() (*ProgObjects, error) {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("lock memory error: %v", err)
	}
	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("load bpf objects error: %v", err)
	}
	tpWakeup, err := link.Tracepoint("sched", "sched_wakeup", objs.HandleSchedWakeup)
	if err != nil {
		return nil, fmt.Errorf("link tracepoint sched_wakeup error: %v", err)
	}
	tpWakeupNew, err := link.Tracepoint("sched", "sched_wakeup_new", objs.HandleSchedWakeupNew)
	if err != nil {
		return nil, fmt.Errorf("link tracepoint sched_wakeup_new error: %v", err)
	}
	tpSwitch, err := link.Tracepoint("sched", "sched_switch", objs.HandleSwitch)
	if err != nil {
		return nil, fmt.Errorf("link tracepoint sched_switch error: %v", err)
	}

	return &ProgObjects{
		Objs: &objs,
		tracepoints: []*link.Link{
			&tpWakeup,
			&tpWakeupNew,
			&tpSwitch,
		},
	}, nil
}

func (p *ProgObjects) DestroyEBPFProg() (err error) {
	for _, tracepoint := range p.tracepoints {
		newErr := (*tracepoint).Close()
		err = multierr.Append(err, newErr)
	}
	newErr := p.Objs.Close()
	err = multierr.Append(err, newErr)
	return
}

// GetCgroupScheduleLatencyAvg get cgroup delay and counter with filtering from cgroupNames.
// @delay is total delay in nanosecond for all pids within this cgroup in the last time window.
// @counter is total number of finish_task_switch() is called for all pids within this cgroup in tha last time window.
// @return each cgroup's average CPU schedule latency in the last time window as the result of @delay / @counter.
func (p *ProgObjects) GetCgroupScheduleLatencyAvg(cgroupNames []string) (map[string]float64, error) {
	cgroupLatencyAvg := map[string]float64{}
	for _, name := range cgroupNames {
		cgroupLatencyAvg[name] = float64(0)
	}
	var delay, counter uint64
	var cgroupNameArray []byte
	cgroupComputationMap := map[string][2]uint64{}
	// Use MapIterator.Next() instead of MapIterator.Lookup() due to problems of marshaling and searching bpf map key
	// of type char array.
	delayIterator := p.Objs.OutputCgroupDelay.Iterate()
	for delayIterator.Next(&cgroupNameArray, &delay) {
		nameFromKernel := unix.ByteSliceToString(cgroupNameArray)
		// filter cgroup name from eBPF by input cgroupNames
		if _, ok := cgroupLatencyAvg[nameFromKernel]; ok {
			cgroupComputationMap[nameFromKernel] = [2]uint64{delay, 0}
		}
	}
	err := delayIterator.Err()
	counterIterator := p.Objs.OutputCgroupDelay.Iterate()
	for counterIterator.Next(&cgroupNameArray, &counter) {
		nameFromKernel := unix.ByteSliceToString(cgroupNameArray)
		// filter cgroup name directly from cgroupComputationMap
		// iterator delay and counter may have different cgroup names, but technically not, this special case simply
		// make array[1] which is counter zero
		if array, ok := cgroupComputationMap[nameFromKernel]; ok {
			array[1] = counter
		}
	}
	for name, array := range cgroupComputationMap {
		if array[1] != 0 {
			cgroupLatencyAvg[name] = float64(array[0]) / float64(array[1])
		}
	}
	err = multierr.Append(err, counterIterator.Err())
	return cgroupLatencyAvg, err
}
