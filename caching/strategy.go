package main

type strategy interface {
	getName() string
	getStoreNodes() int64
	stepTo(time int64, shortestSatPaths *map[int64]map[int64]satPath, gndSatLinks *map[int64]gndSatLink, requests *[]*request) (*[]txRecord, *[]storeRecord, *[]cacheRecord, *[]hopsRecord)
}
