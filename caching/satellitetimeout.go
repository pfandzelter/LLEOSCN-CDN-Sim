package main

type satelliteTimeoutCache struct {
	lastUpdate   int64
	satsPerPlane int64
	numPlanes    int64
	cache        map[int64]map[int64]struct{}

	itemSizes map[int64]int64
}

func newSatelliteTimeout(itemSizes *map[int64]int64) *satelliteTimeoutCache {
	return &satelliteTimeoutCache{
		satsPerPlane: 66,
		numPlanes:    24,
		cache:        make(map[int64]map[int64]struct{}),
		itemSizes:    *itemSizes,
	}
}

func (C *satelliteTimeoutCache) getName() string {
	return "SATELLITE-TIMEOUT"
}

func (C *satelliteTimeoutCache) getStoreNodes() int64 {
	return 66 * 24
}

func (C *satelliteTimeoutCache) stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord) {

	txRecords := []txRecord{}
	storeRecords := []storeRecord{}
	cacheRecords := []cacheRecord{}
	hopsRecords := []hopsRecord{}

	// every 87 seconds: invalidate everything
	// 5730s / 66 = 86.8

	if time-C.lastUpdate >= 87 {
		C.lastUpdate = time
		C.cache = make(map[int64]map[int64]struct{})
	}

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
