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
)

func getGSTPopulation(cityFile string) *map[int64]int64 {
	populations := make(map[int64]int64)

	cities, err := os.Open(cityFile)

	if err != nil {
		panic(err)
	}

	defer cities.Close()

	csvr := csv.NewReader(cities)

	// skip header
	if _, err = csvr.Read(); err != nil {
		panic(err)
	}

	var id int64 = -1

	for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

		// city id is the number in file -1 * -1, starts with -1 (line 1)
		pop, err := strconv.ParseInt(line[1], 10, 64)

		if err != nil {
			continue
		}

		populations[id] = pop

		id--
	}

	return &populations
}
