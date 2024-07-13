// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file benchmarks Marshal[ToBytes] and Unmarshal[Bytes].

package umsgpack_test

import (
	"testing"
	"time"

	. "github.com/viettrungluu/umsgpack"
)

var benchmarkMarshalCorpus = []any{
	map[any]any{
		"foo": map[string]any{
			"zero_int":      0,
			"small_pos_int": 123,
			"small_neg_int": -5,
			"large_pos_int": 1234567890,
		},
		"bar": []any{uint(0), uint(123), uint(1234567890), "abc", "de", false},
		"baz": true,
		"quux": map[any]any{
			"float64":       123.45,
			"data":          fillerBytes(67),
			"small_string":  string(fillerChars(34)),
			"medium_string": string(fillerChars(123)),
		},
		"time_1":        time.Unix(0x12345678, 0),
		"time_2":        time.Unix(0x23456789a, 123456789),
		"another_array": []string{"a", "bc", "d", "ef", "ghi", "jk", "l", "mn", "opq", "rstu", "vwx", "yz"},
		"another_map": map[int]float64{
			12: 34.5,
			6:  7.8,
			9:  0,
		},
	},
}

// Filled lazily.
var benchmarkUnmarshalCorpus [][]byte

func ensureBenchmarkUnmarshalCorpus(b *testing.B) {
	if benchmarkUnmarshalCorpus == nil {
		for _, obj := range benchmarkMarshalCorpus {
			if encoded, err := MarshalToBytes(nil, obj); err != nil {
				b.Fatalf("MarshalToBytes failed: %v", err)
			} else {
				benchmarkUnmarshalCorpus = append(benchmarkUnmarshalCorpus, encoded)
			}
		}
	}
	b.ResetTimer()
}

var benchmarkMarshalToBytesSink []byte

func BenchmarkMarshalToBytes(b *testing.B) {
	for i := 0; i < b.N; i += 1 {
		obj := benchmarkMarshalCorpus[i%len(benchmarkMarshalCorpus)]
		if encoded, err := MarshalToBytes(nil, obj); err != nil {
			b.Fatalf("MarshalToBytes failed: %v", err)
		} else {
			benchmarkMarshalToBytesSink = encoded
		}
	}
}

var benchmarkUnmarshalBytesSink any

func BenchmarkUnmarshalBytes(b *testing.B) {
	ensureBenchmarkUnmarshalCorpus(b)
	for i := 0; i < b.N; i += 1 {
		encoded := benchmarkUnmarshalCorpus[i%len(benchmarkUnmarshalCorpus)]
		if obj, err := UnmarshalBytes(nil, encoded); err != nil {
			b.Fatalf("UnmarshalBytes failed: %v", err)
		} else {
			benchmarkUnmarshalBytesSink = obj
		}
	}
}
