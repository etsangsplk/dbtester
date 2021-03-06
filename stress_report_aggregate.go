// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbtester

import (
	"sort"
	"time"

	"github.com/coreos/dbtester/pkg/report"
)

// processTimeSeries sorts all data points by its timestamp.
// And then aggregate by the cumulative throughput,
// in order to map the number of keys to the average latency.
//
//	type DataPoint struct {
//		Timestamp  int64
//		MinLatency time.Duration
//		AvgLatency time.Duration
//		MaxLatency time.Duration
//		ThroughPut int64
//	}
//
// If unis is 1000 and the average throughput per second is 30,000
// and its average latency is 10ms, it will have 30 data points with
// latency 10ms.
func processTimeSeries(tss report.TimeSeries, unit int64, totalRequests int64) keyNumToAvgLatencys {
	sort.Sort(tss)

	cumulKeyN := int64(0)
	maxKey := int64(0)

	rm := make(map[int64]keyNumToAvgLatency)

	// this data is aggregated by second
	// and we want to map number of keys to latency
	// so the range is the key
	// and the value is the cumulative throughput
	for _, ts := range tss {
		cumulKeyN += ts.ThroughPut
		if cumulKeyN < unit {
			// not enough data points yet
			continue
		}

		// cumulKeyN >= unit
		for cumulKeyN > maxKey {
			maxKey += unit
			rm[maxKey] = keyNumToAvgLatency{minLat: ts.MinLatency, avgLat: ts.AvgLatency, maxLat: ts.MaxLatency}
		}
	}

	// fill-in empty rows
	for i := maxKey; i < totalRequests; i += unit {
		if _, ok := rm[i]; !ok {
			rm[i] = keyNumToAvgLatency{}
		}
	}
	if _, ok := rm[totalRequests]; !ok {
		rm[totalRequests] = keyNumToAvgLatency{}
	}

	kss := []keyNumToAvgLatency{}
	delete(rm, 0)
	for k, v := range rm {
		// make sure to use 'k' as keyNum
		kss = append(kss, keyNumToAvgLatency{keyNum: k, minLat: v.minLat, avgLat: v.avgLat, maxLat: v.maxLat})
	}
	sort.Sort(keyNumToAvgLatencys(kss))

	return kss
}

type keyNumToAvgLatency struct {
	keyNum int64
	minLat time.Duration
	avgLat time.Duration
	maxLat time.Duration
}

type keyNumToAvgLatencys []keyNumToAvgLatency

func (t keyNumToAvgLatencys) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t keyNumToAvgLatencys) Len() int           { return len(t) }
func (t keyNumToAvgLatencys) Less(i, j int) bool { return t[i].keyNum < t[j].keyNum }
