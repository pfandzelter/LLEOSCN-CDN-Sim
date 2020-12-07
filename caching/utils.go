package main

import "strconv"

func copyCache(cache map[int64]map[int64]struct{}) map[int64]map[int64]struct{} {
	cp := make(map[int64]map[int64]struct{})

	for node, localCache := range cache {

		cp[node] = make(map[int64]struct{})

		for item := range localCache {
			cp[node][item] = struct{}{}
		}
	}

	return cp
}

func strToInt64(items *[]string) *[]int64 {
	sp := make([]int64, len(*items))

	for i, n := range *items {
		id, err := strconv.ParseInt(n, 10, 64)

		if err != nil {
			panic(err)
		}

		sp[i] = id
	}

	return &sp
}
