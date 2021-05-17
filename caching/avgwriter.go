/*
* This file is part of LLEOSCN-CDN-Sim (https://github.com/pfandzelter/LLEOSCN-CDN-Sim).
* Copyright (c) 2020 Tobias Pfandzelter.
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation, version 3.
*
* This program is distributed in the hope that it will be useful, but
* WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
* General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this program. If not, see <http://www.gnu.org/licenses/>.
**/

package main

import (
	"bufio"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type aw struct {
	cacheStrategy string
	filename      string
	itemSizes     *map[int64]int64

	// performance optimization
	// for ground stations, where new items only come but are never
	// deleted from cache, store only the additional stored items
	// instead of iterating over the whole array again
	// ...
	// i'm sorry
	cachedStoreRecords    map[string]map[int64]int64
	cachedStoreRecordsNum map[string]int64

	storeNodesPerStrategy map[string]int64
}

func newAvgWriter(filename string, itemSizes *map[int64]int64, c <-chan writeSet, storeNodesPerStrategy map[string]int64) {

	f := aw{
		filename:              filename,
		itemSizes:             itemSizes,
		cachedStoreRecords:    make(map[string]map[int64]int64),
		cachedStoreRecordsNum: make(map[string]int64),
		storeNodesPerStrategy: storeNodesPerStrategy,
	}

	for w := range c {
		if w.time == -1 {
			break
		}

		f.write(w.time, w.strategyName, w.txRecords, w.storeRecords, w.cacheRecords, w.hopsRecords)
	}
}

// assumes val is sorted from low to high
func (f *aw) calcPercentile(val *[]int, p int64) float64 {

	// from https://github.com/aclements/go-moremath/blob/master/stats/sample.go#L232

	q := float64(p) / 100.0

	N := float64(len(*val))

	n := 1/3.0 + q*(N+1/3.0) // R8

	kf, frac := math.Modf(n)

	k := int(kf)

	if k <= 0 {
		return float64((*val)[0])
	} else if k >= len(*val) {
		return float64((*val)[len(*val)-1])
	}
	return float64((*val)[k-1]) + frac*(float64((*val)[k]-(*val)[k-1]))
}

func (f *aw) write(time int64, strategyName string, txRecords *[]txRecord, storeRecords *[]storeRecord, cacheRecords *[]cacheRecord, hopsRecords *[]hopsRecord) {
	baseFilename := f.filename + strconv.FormatInt(time, 10) + strategyName
	f.writeTX(baseFilename+"tx", txRecords)
	f.writeStore(baseFilename+"store", strategyName, storeRecords)
	f.writeCache(baseFilename+"cache", cacheRecords)
	f.writeHops(baseFilename+"hops", hopsRecords)
}

// writeTX writes the following to file:
// * total data flow in system
// * max data flow per sat
// * min data flow per sat
// * avg data flow per sat
// * median data flow per sat
// * 95th pcntl data flow per sat
// * 99th pcntl data flow per sat
func (f *aw) writeTX(filename string, records *[]txRecord) {

	var total int64

	flowPerSat := make(map[int64]int64)

	for _, r := range *records {
		total += r.bandwidth

		// ignore gst
		if r.source >= 0 {
			flowPerSat[r.source] += r.bandwidth
		}

		if r.target >= 0 {
			flowPerSat[r.target] += r.bandwidth
		}
	}

	var maxFlow int
	var minFlow int
	var avgFlow float64
	var medianFlow float64
	var p95 float64
	var p99 float64

	// no flow per sat? everything is 0
	if len(flowPerSat) > 0 {
		// Convert map to slice of values
		flowVals := make([]int, len(flowPerSat))
		i := 0
		totalPerSat := 0

		for _, flow := range flowPerSat {
			flowVals[i] = int(flow)
			totalPerSat += int(flow)
			i++
		}

		// sort flow vals
		sort.Ints(flowVals)

		maxFlow = flowVals[len(flowVals)-1]
		minFlow = flowVals[0]

		avgFlow = float64(totalPerSat) / float64(len(flowVals))

		medianFlow = f.calcPercentile(&flowVals, 50)
		p95 = f.calcPercentile(&flowVals, 95)
		p99 = f.calcPercentile(&flowVals, 99)
	}

	txFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer txFile.Close()

	buf := bufio.NewWriter(txFile)

	// * total data flow in system
	buf.WriteString("total,")
	buf.WriteString(strconv.FormatInt(total, 10))
	buf.WriteString("\n")

	// * max data flow per sat
	buf.WriteString("max,")
	buf.WriteString(strconv.Itoa(maxFlow))
	buf.WriteString("\n")

	// * min data flow per sat
	buf.WriteString("min,")
	buf.WriteString(strconv.Itoa(minFlow))
	buf.WriteString("\n")

	// * avg data flow per sat
	buf.WriteString("avg,")
	buf.WriteString(strconv.FormatFloat(avgFlow, 'f', -1, 64))
	buf.WriteString("\n")

	// * median data flow per sat
	buf.WriteString("median,")
	buf.WriteString(strconv.FormatFloat(medianFlow, 'f', -1, 64))
	buf.WriteString("\n")

	// * 95th pcntl data flow per sat
	buf.WriteString("95th,")
	buf.WriteString(strconv.FormatFloat(p95, 'f', -1, 64))
	buf.WriteString("\n")

	// * 99th pcntl data flow per sat
	buf.WriteString("99th,")
	buf.WriteString(strconv.FormatFloat(p99, 'f', -1, 64))
	buf.WriteString("\n")

	buf.Flush()
}

// writeStore writes the following to file
// * total storage per store node
// * max storage use per store node
// * min storage use per store node
// * avg storage use per store node
// * median storage use per store node
// * 95th pcntl storage use per store node
// * 99th pcntl storage use per store node
// * amount of nodes
// * amount of nodes without store
func (f *aw) writeStore(filename string, strategyName string, records *[]storeRecord) {

	strPerNode := make(map[int64]int64)

	if strings.Contains(strategyName, "GROUND-STATION") {
		if _, ok := f.cachedStoreRecords[strategyName]; !ok {
			f.cachedStoreRecords[strategyName] = make(map[int64]int64)
		}

		strPerNode = f.cachedStoreRecords[strategyName]

		for _, r := range *records {
			strPerNode[r.node] += (*f.itemSizes)[r.item]
		}

		f.cachedStoreRecords[strategyName] = strPerNode

	} else {
		for _, r := range *records {
			strPerNode[r.node] += (*f.itemSizes)[r.item]
		}
	}

	var total int64

	var maxStore int
	var minStore int
	var avgStore float64
	var medianStore float64
	var p95 float64
	var p99 float64
	var noStoreNodes int64

	// build store array
	storePerNode := make([]int, f.storeNodesPerStrategy[strategyName])

	i := 0
	for _, store := range strPerNode {
		storePerNode[i] = int(store)
		total += store
		i++
	}

	noStoreNodes = f.storeNodesPerStrategy[strategyName] - int64(i)

	// no store per node? everything is 0
	if len(storePerNode) > 0 {
		// sort flow vals
		sort.Ints(storePerNode)

		maxStore = storePerNode[len(storePerNode)-1]
		minStore = storePerNode[0]

		avgStore = float64(total) / float64(f.storeNodesPerStrategy[strategyName])

		medianStore = f.calcPercentile(&storePerNode, 50)
		p95 = f.calcPercentile(&storePerNode, 95)
		p99 = f.calcPercentile(&storePerNode, 99)
	}

	storeFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer storeFile.Close()

	buf := bufio.NewWriter(storeFile)

	// * total storage per store node
	buf.WriteString("total,")
	buf.WriteString(strconv.FormatInt(total, 10))
	buf.WriteString("\n")

	// * max storage use per store node
	buf.WriteString("max,")
	buf.WriteString(strconv.Itoa(maxStore))
	buf.WriteString("\n")

	// * min storage use per store node
	buf.WriteString("min,")
	buf.WriteString(strconv.Itoa(minStore))
	buf.WriteString("\n")

	// * avg storage use per store node
	buf.WriteString("avg,")
	buf.WriteString(strconv.FormatFloat(avgStore, 'f', -1, 64))
	buf.WriteString("\n")

	// * median storage use per store node
	buf.WriteString("median,")
	buf.WriteString(strconv.FormatFloat(medianStore, 'f', -1, 64))
	buf.WriteString("\n")

	// * 95th pcntl storage use per store node
	buf.WriteString("95th,")
	buf.WriteString(strconv.FormatFloat(p95, 'f', -1, 64))
	buf.WriteString("\n")

	// * 99th pcntl storage use per store node
	buf.WriteString("99th,")
	buf.WriteString(strconv.FormatFloat(p99, 'f', -1, 64))
	buf.WriteString("\n")

	// * amount of nodes that can store data
	buf.WriteString("numnodes,")
	buf.WriteString(strconv.FormatInt(f.storeNodesPerStrategy[strategyName], 10))
	buf.WriteString("\n")

	// * amount of nodes without store
	buf.WriteString("numnostorenodes,")
	buf.WriteString(strconv.FormatInt(noStoreNodes, 10))
	buf.WriteString("\n")

	buf.Flush()
}

// writeCache writes the following to file
// * cache hit ratio
// * number of requests
func (f *aw) writeCache(filename string, records *[]cacheRecord) {
	numSuccess := 0
	numRequests := 0

	for _, r := range *records {
		if r.success {
			numSuccess++
		}
		numRequests++
	}

	ratio := float64(numSuccess) / float64(numRequests)

	cacheFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer cacheFile.Close()

	buf := bufio.NewWriter(cacheFile)

	buf.WriteString("ratio,")
	buf.WriteString(strconv.FormatFloat(ratio, 'f', -1, 64))
	buf.WriteString("\n")

	buf.WriteString("num_requests,")
	buf.WriteString(strconv.Itoa(numRequests))
	buf.WriteString("\n")

	buf.Flush()
}

// writeHops writes the following to file
// * max hops for requests
// * min hops for requests
// * avg hops for requests
// * median hops for requests
// * 95th pcntl hops for requests
// * 99th pcntl hops for requests
func (f *aw) writeHops(filename string, records *[]hopsRecord) {

	hops := make([]int, len(*records))

	totalHops := 0

	for i, r := range *records {
		hops[i] = int(r.hops)
		totalHops += int(r.hops)
	}

	// sort flow vals
	sort.Ints(hops)

	maxHops := hops[len(hops)-1]
	minHops := hops[0]

	avgHops := float64(totalHops) / float64(len(hops))

	medianHops := f.calcPercentile(&hops, 50)
	p95 := f.calcPercentile(&hops, 95)
	p99 := f.calcPercentile(&hops, 99)

	hopsFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer hopsFile.Close()

	buf := bufio.NewWriter(hopsFile)

	// * total data flow in system
	buf.WriteString("total,")
	buf.WriteString(strconv.Itoa(totalHops))
	buf.WriteString("\n")

	// * max data flow per sat
	buf.WriteString("max,")
	buf.WriteString(strconv.Itoa(maxHops))
	buf.WriteString("\n")

	// * min data flow per sat
	buf.WriteString("min,")
	buf.WriteString(strconv.Itoa(minHops))
	buf.WriteString("\n")

	// * avg data flow per sat
	buf.WriteString("avg,")
	buf.WriteString(strconv.FormatFloat(avgHops, 'f', -1, 64))
	buf.WriteString("\n")

	// * median data flow per sat
	buf.WriteString("median,")
	buf.WriteString(strconv.FormatFloat(medianHops, 'f', -1, 64))
	buf.WriteString("\n")

	// * 95th pcntl data flow per sat
	buf.WriteString("95th,")
	buf.WriteString(strconv.FormatFloat(p95, 'f', -1, 64))
	buf.WriteString("\n")

	// * 99th pcntl data flow per sat
	buf.WriteString("99th,")
	buf.WriteString(strconv.FormatFloat(p99, 'f', -1, 64))
	buf.WriteString("\n")

	buf.Flush()

	buf.Flush()
}
