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

type satelliteCache struct {
	cache map[int64]map[int64]struct{}
}

func newSatellite() *satelliteCache {
	return &satelliteCache{
		cache: make(map[int64]map[int64]struct{}),
	}
}

func (C *satelliteCache) getName() string {
	return "SATELLITE"
}

func (C *satelliteCache) getStoreNodes() int64 {
	return 66 * 24
}

func (C *satelliteCache) stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord) {

	txRecords := []txRecord{}
	storeRecords := []storeRecord{}
	// we always need as many cache records as we have requests
	cacheRecords := make([]cacheRecord, 0, len(*requests))
	// same goes for hops records
	hopsRecords := make([]hopsRecord, 0, len(*requests))

	// prepare a copied cache so we can modify the real cache
	scache := copyCache(C.cache)

	for _, req := range *requests {
		// hops := int64(0)
		// success := false

		// for i := 0; i < len(req.path)-1; i++ {

		// 	source := req.path[i]
		// 	target := req.path[i+1]

		// 	if _, ok := scache[source]; ok {
		// 		if _, ok := scache[source][req.item]; ok {
		// 			success = true
		// 			break
		// 		}
		// 	}

		// 	txRecords = append(txRecords, txRecord{
		// 		source:    source,
		// 		target:    target,
		// 		bandwidth: req.bandwidth,
		// 	})

		// 	hops++
		// }

		// firstSat := req.path[1]

		hops := int64(0)
		success := false

		// check if the satellite that got  the request first has the item in cache
		firstSat := req.path[1]
		if _, ok := scache[firstSat]; ok {
			if _, ok := scache[firstSat][req.item]; ok {
				txRecords = append(txRecords, txRecord{
					source:    req.path[0],
					target:    firstSat,
					bandwidth: req.bandwidth,
				})

				hops = 1

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
		if _, ok := C.cache[firstSat]; !ok {
			C.cache[firstSat] = make(map[int64]struct{})
		}

		C.cache[firstSat][req.item] = struct{}{}
	}

	for node := range C.cache {
		for item := range C.cache[node] {
			storeRecords = append(storeRecords, storeRecord{
				node: node,
				item: item,
			})
		}
	}

	return &txRecords, &storeRecords, &cacheRecords, &hopsRecords

}
