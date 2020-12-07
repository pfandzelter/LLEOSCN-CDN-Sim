package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
)

func getRequests(pathFile string, numRequests int) *[]*request {
	r, err := os.Open(pathFile)

	if err != nil {
		panic(err)
	}

	defer r.Close()

	csvr := csv.NewReader(r)

	// skip header
	if _, err = csvr.Read(); err != nil {
		panic(err)
	}

	requests := make([]*request, 0, numRequests)

	for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

		// first item: requested item
		item, err := strconv.ParseInt(line[0], 10, 64)

		if err != nil {
			panic(err)
		}

		// second item: item size
		// this is a bit ugly
		// item size is stored in the form "123.0"
		// it's always ".0"
		// but since working with int64 is a lot more comfortable than working with float64, we cut this part off
		// sorry

		if line[1][len(line[1])-2:] != ".0" {
			panic("weird stuff: " + line[1])
		}

		size, err := strconv.ParseInt(line[1][:len(line[1])-2], 10, 64)

		if err != nil {
			panic(err)
		}

		// third item: path delimited by "|"
		path := strings.Split(line[2], "|")

		requests = append(requests, &request{
			item:      item,
			bandwidth: size,
			path:      *strToInt64(&path),
		})
	}

	return &requests
}
