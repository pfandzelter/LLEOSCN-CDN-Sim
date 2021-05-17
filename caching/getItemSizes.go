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

func getItemSizes(loadFile string) *map[int64]int64 {
	itemSizes := make(map[int64]int64)

	load, err := os.Open(loadFile)

	if err != nil {
		panic(err)
	}

	defer load.Close()

	csvr := csv.NewReader(load)

	// skip header
	if _, err = csvr.Read(); err != nil {
		panic(err)
	}

	for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

		item, err := strconv.ParseInt(line[0], 10, 64)

		if err != nil {
			continue
		}

		if line[3][len(line[3])-2:] != ".0" {
			panic("weird stuff: " + line[3])
		}

		size, err := strconv.ParseInt(line[3][:len(line[3])-2], 10, 64)

		if err != nil {
			continue
		}

		itemSizes[item] = size
	}

	return &itemSizes
}
