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
