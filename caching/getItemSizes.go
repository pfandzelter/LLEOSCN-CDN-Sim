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
