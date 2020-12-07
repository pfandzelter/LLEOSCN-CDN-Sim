package main

type noneCache struct{}

func newNone() *noneCache {
	return &noneCache{}
}

func (C *noneCache) getName() string {
	return "NONE"
}

func (C *noneCache) getStoreNodes() int64 {
	return 0
}

func (C *noneCache) stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord) {

	txRecords := []txRecord{}
	storeRecords := []storeRecord{}
	// we always need as many cache records as we have requests
	cacheRecords := make([]cacheRecord, 0, len(*requests))
	// same goes for hops records
	hopsRecords := make([]hopsRecord, 0, len(*requests))

	for _, req := range *requests {
		for i := 0; i < len(req.path)-1; i++ {
			source := req.path[i]
			target := req.path[i+1]

			txRecords = append(txRecords, txRecord{
				source:    source,
				target:    target,
				bandwidth: req.bandwidth,
			})
		}

		cacheRecords = append(cacheRecords, cacheRecord{
			item:    req.item,
			success: false,
		})

		hopsRecords = append(hopsRecords, hopsRecord{
			item: req.item,
			hops: int64(len(req.path)) - 1,
		})
	}

	return &txRecords, &storeRecords, &cacheRecords, &hopsRecords

}
