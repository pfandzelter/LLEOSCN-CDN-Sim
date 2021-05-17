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
