package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

func getGroundSatLinks(gslFile string) *map[int64]gndSatLink {
	gsl, err := os.Open(gslFile)

	if err != nil {
		panic(err)
	}

	defer gsl.Close()

	csvr := csv.NewReader(gsl)

	// skip header
	if _, err = csvr.Read(); err != nil {
		panic(err)
	}

	gndSatLinks := make(map[int64]gndSatLink)

	for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

		// first item: ground station
		gnd, err := strconv.ParseInt(line[0], 10, 64)

		if err != nil {
			panic(err)
		}

		// second item: nearest sat
		sat, err := strconv.ParseInt(line[1], 10, 64)

		if err != nil {
			panic(err)
		}

		// third item: distance
		distance, err := strconv.ParseInt(line[1], 10, 64)

		if err != nil {
			panic(err)
		}

		gndSatLinks[gnd] = gndSatLink{
			sat:      sat,
			distance: distance,
		}
	}

	return &gndSatLinks
}
