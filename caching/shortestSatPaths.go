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

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
)

func getShortestSatPaths(sspFile string) *map[int64]map[int64]satPath {
	ssp, err := os.Open(sspFile)

	if err != nil {
		panic(err)
	}

	defer ssp.Close()

	csvr := csv.NewReader(ssp)

	// skip header
	if _, err = csvr.Read(); err != nil {
		panic(err)
	}

	shortestSatPaths := make(map[int64]map[int64]satPath)

	for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

		// first item: source sat
		source, err := strconv.ParseInt(line[0], 10, 64)

		if err != nil {
			panic(err)
		}

		// second item: target sat
		target, err := strconv.ParseInt(line[1], 10, 64)

		if err != nil {
			panic(err)
		}

		// third item: distance
		distance, err := strconv.ParseInt(line[2], 10, 64)

		if err != nil {
			panic(err)
		}

		// fourth item: path delimited by "|"
		path := strings.Split(line[2], "|")

		if _, ok := shortestSatPaths[source]; !ok {
			shortestSatPaths[source] = make(map[int64]satPath)
		}

		shortestSatPaths[source][target] = satPath{
			path:     strToInt64(&path),
			distance: distance,
		}
	}

	return &shortestSatPaths
}
