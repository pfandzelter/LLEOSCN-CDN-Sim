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
