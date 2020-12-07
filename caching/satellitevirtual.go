package main

type satelliteVirtualCache struct {
	lastIntra    int64
	lastCross    int64
	satsPerPlane int64
	numPlanes    int64
	cache        map[int64]map[int64]struct{}
	itemSizes    map[int64]int64
}

func newSatelliteVirtual(itemSizes *map[int64]int64) *satelliteVirtualCache {
	return &satelliteVirtualCache{
		satsPerPlane: 66,
		numPlanes:    24,
		cache:        make(map[int64]map[int64]struct{}),
		itemSizes:    *itemSizes,
	}
}

func (C *satelliteVirtualCache) getName() string {
	return "SATELLITE-VIRTUAL"
}

func (C *satelliteVirtualCache) getStoreNodes() int64 {
	return 66 * 24
}

func (C *satelliteVirtualCache) stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord) {

	txRecords := []txRecord{}
	// at least as much as satellites already caching stuff
	storeRecords := make([]storeRecord, 0, len(C.cache))
	// we always need as many cache records as we have requests
	cacheRecords := make([]cacheRecord, 0, len(*requests))
	// same goes for hops records
	hopsRecords := make([]hopsRecord, 0, len(*requests))

	// every 86 seconds: intra-plane backward propagation
	if time-C.lastIntra >= 87 {
		newCache := make(map[int64]map[int64]struct{})

		for sat, cache := range C.cache {
			posInPlane := sat % C.satsPerPlane
			thisPlane := (sat - posInPlane) / C.satsPerPlane

			nextSatPosInPlane := posInPlane - 1
			if nextSatPosInPlane <= 0 {
				nextSatPosInPlane = C.satsPerPlane
			}

			// it's possible that we haven't actually calculated the path between these nodes
			// but that's ok!
			// we only have one intra-plane link for intra-plane backward propagation

			// propagate one sat back
			satToPropagateTo := (thisPlane * C.satsPerPlane) + nextSatPosInPlane
			path := []int64{sat, satToPropagateTo}

			if sat > satToPropagateTo {
				path = []int64{satToPropagateTo, sat}
			}

			newCache[satToPropagateTo] = make(map[int64]struct{})

			for item := range cache {
				newCache[satToPropagateTo][item] = struct{}{}

				if _, ok := C.cache[satToPropagateTo]; ok {
					if _, ok := C.cache[satToPropagateTo][item]; ok {
						// item is in cache already, more efficient
						continue
					}
				}

				source := path[0]
				target := path[1]

				txRecords = append(txRecords, txRecord{
					source:    source,
					target:    target,
					bandwidth: C.itemSizes[item],
				})

			}
		}

		C.lastIntra = time
		C.cache = newCache
	}

	// every 3600 seconds: cross-plane forward propagation
	// the first timestamp where both cross- and intra-plane propagation will occur at the same time is 104400s, which is ok for our simulation
	// in theory, if both occur at the same time, there would be no need to first to intra- and then cross-plane, instead both could be merged into one
	if time-C.lastCross >= 3600 {
		newCache := make(map[int64]map[int64]struct{})

		for sat, cache := range C.cache {
			posInPlane := sat % C.satsPerPlane
			thisPlane := (sat - posInPlane) / C.satsPerPlane

			nextPlane := thisPlane + 1
			if nextPlane >= C.numPlanes {
				nextPlane = 0
			}
			// it's possible that we haven't actually calculated the path between these nodes
			// but that's ok!
			// we only have one for cross-plane propagation

			// if that's not the one we want, add another cross-plane hops
			satToPropagateTo := (nextPlane * C.satsPerPlane) + posInPlane

			path := []int64{sat, satToPropagateTo}

			if sat > satToPropagateTo {
				path = []int64{satToPropagateTo, sat}
			}

			newCache[satToPropagateTo] = make(map[int64]struct{})

			for item := range cache {
				newCache[satToPropagateTo][item] = struct{}{}

				if _, ok := C.cache[satToPropagateTo]; ok {
					if _, ok := C.cache[satToPropagateTo][item]; ok {
						// item is in cache already, more efficient
						continue
					}
				}

				source := path[0]
				target := path[1]

				txRecords = append(txRecords, txRecord{
					source:    source,
					target:    target,
					bandwidth: C.itemSizes[item],
				})

			}
		}

		C.lastCross = time
		C.cache = newCache
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
