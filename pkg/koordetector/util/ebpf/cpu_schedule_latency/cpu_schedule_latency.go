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
	"C"
	"log"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"go.uber.org/multierr"
	"golang.org/x/sys/unix"
)

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf csl.bpf.c -- -I../headers

type ProgObjects struct {
	Objs        *bpfObjects
	tracepoints []*link.Link
}

func NewCSLeBPFProg() (*ProgObjects, error) {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}
	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	tpWakeup, err := link.Tracepoint("sched", "sched_wakeup", objs.HandleSchedWakeup, nil)
	if err != nil {
		log.Fatalf("opening tracepoint: %s", err)
	}
	tpWakeupNew, err := link.Tracepoint("sched", "sched_wakeup_new", objs.HandleSchedWakeupNew, nil)
	if err != nil {
		log.Fatalf("opening tracepoint: %s", err)
	}
	tpSwitch, err := link.Tracepoint("sched", "sched_switch", objs.HandleSwitch, nil)
	if err != nil {
		log.Fatalf("opening tracepoint: %s", err)
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

// GetCgroupDelayAndCount get cgroup delay and counter with filtering from cgroupNames.
// Use MapIterator.Next() instead of MapIterator.Lookup() way due to problems of marshalling and searching bpf map key
// of type char array.
// @delay is total delay in nanosecond for all pids within this cgroup in the last time window
// @counter is total number of finish_task_switch() is called for all pids within this cgroup in tha last time window
// To get the cgroup's average CPU schedule latency in the last time window, compute @delay / @counter for each cgroup.
func (p *ProgObjects) GetCgroupDelayAndCount(cgroupNames map[string][2]uint64) (err error) {
	var delay, counter uint64
	var cgroupNameArray []byte
	delayIterator := p.Objs.OutputCgroupDelay.Iterate()
	for delayIterator.Next(&cgroupNameArray, &delay) {
		nameFromKernel := unix.ByteSliceToString(cgroupNameArray)
		if array, ok := cgroupNames[nameFromKernel]; ok {
			array[0] = delay
		}
	}
	err = multierr.Append(err, delayIterator.Err())
	counterIterator := p.Objs.OutputCgroupDelay.Iterate()
	for counterIterator.Next(&cgroupNameArray, &counter) {
		nameFromKernel := unix.ByteSliceToString(cgroupNameArray)
		if array, ok := cgroupNames[nameFromKernel]; ok {
			array[1] = counter
		}
	}
	err = multierr.Append(err, counterIterator.Err())
	return
}
