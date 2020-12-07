package main

type stepSet struct {
	time             int64
	shortestSatPaths *map[int64]map[int64]satPath
	gndSatLinks      *map[int64]gndSatLink
	requests         *[]*request
}

type writeSet struct {
	time         int64
	strategyName string
	txRecords    *[]txRecord
	storeRecords *[]storeRecord
	cacheRecords *[]cacheRecord
	hopsRecords  *[]hopsRecord
}

type satPath struct {
	path     *[]int64
	distance int64
}

type gndSatLink struct {
	sat      int64
	distance int64
}

type request struct {
	item      int64
	bandwidth int64
	path      []int64
}

type txRecord struct {
	source    int64
	target    int64
	bandwidth int64
}

type storeRecord struct {
	node int64
	item int64
}

type cacheRecord struct {
	item    int64
	success bool
}

type hopsRecord struct {
	item int64
	hops int64
}
