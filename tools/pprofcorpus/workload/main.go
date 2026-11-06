// SPDX-License-Identifier: MIT
// Provenance: original

// Command workload is the program that tools/pprofcorpus profiles: hashing,
// sorting and string building under the pprof label work, for the CPU
// profile, then allocations of three sizes, for the heap profile. It writes
// cpu.prof and heap.prof into the directory given as its argument.
package main

import (
	"context"
	"crypto/sha256"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
)

// sink keeps the allocations alive until the heap profile is written.
var sink [][]byte

func hashLoop(n int) {
	b := []byte("seed")
	for range n {
		s := sha256.Sum256(b)
		b = s[:]
	}
}

func sortLoop(n int) {
	for range n {
		xs := make([]int, 2000)
		for i := range xs {
			xs[i] = (i * 7919) % 2003
		}
		slices.Sort(xs)
	}
}

func stringLoop(n int) {
	for range n {
		var sb strings.Builder
		for j := range 200 {
			sb.WriteString(strings.Repeat("x", j%17))
		}
		_ = strings.ToUpper(sb.String())
	}
}

func alloc(size, count int) {
	for range count {
		sink = append(sink, make([]byte, size))
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: workload DIR")
	}
	dir := os.Args[1]
	runtime.MemProfileRate = 1
	cpu, err := os.Create(filepath.Join(dir, "cpu.prof"))
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		log.Fatal(err)
	}
	for _, w := range []struct {
		name string
		f    func()
	}{
		{"hash", func() { hashLoop(400000) }},
		{"sort", func() { sortLoop(3000) }},
		{"string", func() { stringLoop(3000) }},
	} {
		pprof.Do(context.Background(), pprof.Labels("work", w.name), func(context.Context) { w.f() })
	}
	pprof.StopCPUProfile()
	if err := cpu.Close(); err != nil {
		log.Fatal(err)
	}
	alloc(64, 2000)
	alloc(4096, 300)
	alloc(1<<20, 5)
	heap, err := os.Create(filepath.Join(dir, "heap.prof"))
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.WriteHeapProfile(heap); err != nil {
		log.Fatal(err)
	}
	if err := heap.Close(); err != nil {
		log.Fatal(err)
	}
}
