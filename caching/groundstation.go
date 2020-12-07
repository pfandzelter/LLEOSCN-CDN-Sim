package main

import (
	"math/rand"
	"strconv"
)

// offset should be greater than the number of ground stations
const offset int64 = 1000000

type groundstationCache struct {
	name             string
	cache            map[int64]map[int64]map[int64]struct{}
	maxClientsPerGST int64
	gstPopulation    map[int64]int64
	nodes            []int64
}

func newGroundstation(maxClientsPerGST int64, gstPopulation map[int64]int64) *groundstationCache {

	rand.Seed(0)

	cache := make(map[int64]map[int64]map[int64]struct{})
	// test gst set
	nodes := make([]int64, 0)

	for gst, pop := range gstPopulation {
		numGst := pop/maxClientsPerGST + 1
		cache[gst] = make(map[int64]map[int64]struct{})

		for i := int64(0); i < numGst; i++ {
			x := offset*int64(-i) + gst

			if x%offset != gst {
				panic("offset too small!")
			}

			nodes = append(nodes, x)
			cache[gst][x] = make(map[int64]struct{})
		}
	}

	return &groundstationCache{
		name:             "GROUND-STATION" + "-" + strconv.FormatInt(maxClientsPerGST, 10),
		cache:            cache,
		gstPopulation:    gstPopulation,
		maxClientsPerGST: maxClientsPerGST,
		nodes:            nodes,
	}
}

func (C *groundstationCache) getStoreNodes() int64 {
	return int64(len(C.nodes))
}

func (C *groundstationCache) getRandInGST(gst int64) int64 {
	i := rand.Intn(int(C.gstPopulation[gst]/C.maxClientsPerGST + 1))
	return offset*int64(-i) + gst

}

func (C *groundstationCache) getName() string {
	return C.name
}

func (C *groundstationCache) getStore(scache map[int64]map[int64]map[int64]struct{}) *[]storeRecord {

	storeRecords := []storeRecord{}

	for gst := range scache {
		for node := range scache[gst] {
			for item := range scache[gst][node] {
				storeRecords = append(storeRecords, storeRecord{
					node: node,
					item: item,
				})
			}
		}
	}

	return &storeRecords
}

func (C *groundstationCache) stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord) {

	txRecords := []txRecord{}
	// we always need as many cache records as we have requests
	cacheRecords := make([]cacheRecord, 0, len(*requests))
	// same goes for hops records
	hopsRecords := make([]hopsRecord, 0, len(*requests))

	// prepare a new cache that will store additions to the cache
	scache := make(map[int64]map[int64]map[int64]struct{})

	// for gst := range C.cache {
	// 	scache[gst] = copyCache(C.cache[gst])
	// }

	for _, req := range *requests {
		hops := int64(0)
		success := false

		// check if the ground station that makes the request has the item in cache
		// the ground station is a random one found in gstSet[<this_ground_station>]
		actualGST := req.path[0]
		cacheGst := C.getRandInGST(actualGST)

		if _, ok := C.cache[actualGST][cacheGst]; ok {
			if _, ok := C.cache[actualGST][cacheGst][req.item]; ok {
				success = true
			}
		}

		// if it didn't: request to origin server
		if !success {
			for i := 0; i < len(req.path)-1; i++ {
				source := req.path[i]
				target := req.path[i+1]

				txRecords = append(txRecords, txRecord{
					source:    source,
					target:    target,
					bandwidth: req.bandwidth,
				})

				hops++
			}
		}

		cacheRecords = append(cacheRecords, cacheRecord{
			item:    req.item,
			success: success,
		})

		hopsRecords = append(hopsRecords, hopsRecord{
			item: req.item,
			hops: hops,
		})

		// write that item into the cache for the next round
		if _, ok := scache[actualGST]; !ok {
			scache[actualGST] = make(map[int64]map[int64]struct{})
		}

		if _, ok := scache[actualGST][cacheGst]; !ok {
			scache[actualGST][cacheGst] = make(map[int64]struct{})
		}

		scache[actualGST][cacheGst][req.item] = struct{}{}
	}

	// transfer the items from the temporary scache into the main cache for next round
	for actualGST := range scache {
		for cacheGst := range scache[actualGST] {
			for item := range scache[actualGST][cacheGst] {
				C.cache[actualGST][cacheGst][item] = struct{}{}
			}
		}
	}

	return &txRecords, C.getStore(scache), &cacheRecords, &hopsRecords

}
