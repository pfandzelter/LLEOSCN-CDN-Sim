package main

import (
	"encoding/csv"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/pelletier/go-toml"
	"github.com/schollz/progressbar/v3"
)

func main() {
	if len(os.Args) != 3 {
		panic("not enough arguments given")
	}

	conf := os.Args[1]
	cachingStrategy := os.Args[2]

	config, err := toml.LoadFile(conf)

	if err != nil {
		panic(err)
	}

	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	workloadFolder := path.Join(cwd, "workloads", config.Get("name").(string))

	workloadConf := path.Join(workloadFolder, "config.toml")

	workloadConfig, err := toml.LoadFile(workloadConf)

	if err != nil {
		panic(err)
	}

	steps := workloadConfig.Get("steps").(int64)
	stepLength := workloadConfig.Get("step_length").(int64)

	cacheFiles := path.Join(workloadFolder, "cache", "c.csv")

	loadFile := path.Join(workloadFolder, workloadConfig.Get("loadfile").(string))

	analysisFile := path.Join(workloadFolder, "analysis.csv"+cachingStrategy)

	f, err := os.Create(analysisFile)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("time" + "," + cachingStrategy + "TX" + "," + cachingStrategy + "STR" + "," + cachingStrategy + "HITS" + "," + cachingStrategy + "HOPS")

	pbar := progressbar.Default(steps)

	var time int64 = 0

	for ; time < steps*stepLength; time += stepLength {
		f.WriteString("\n")
		f.WriteString(strconv.FormatInt(time, 10))
		f.WriteString(",")

		// Analyze Bandwidth Use
		totalBW := 0.0

		bandwidthFile, err := os.Open(cacheFiles + strconv.FormatInt(time, 10) + cachingStrategy + "tx")

		if err != nil {
			panic(err)
		}

		defer bandwidthFile.Close()

		csvr := csv.NewReader(bandwidthFile)

		for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {
			bw, err := strconv.ParseFloat(line[2], 64)
			if err != nil {
				continue
			}
			totalBW += bw
		}

		bWinMBit := totalBW / 1000.0 / 1000.0

		f.WriteString(strconv.FormatFloat(bWinMBit, 'f', -1, 64))

		// Analyze Storage Use

		f.WriteString(",")

		// 1. find size for each item
		itemSize := make(map[int]float64)

		load, err := os.Open(loadFile)

		if err != nil {
			panic(err)
		}

		defer load.Close()

		csvr = csv.NewReader(load)

		for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

			item, err := strconv.Atoi(line[0])

			if err != nil {
				continue
			}

			size, err := strconv.ParseFloat(line[3], 64)

			if err != nil {
				continue
			}

			itemSize[item] = size
		}

		// 2. calculate storage per node
		storage := make(map[int]float64)

		totalBW = 0

		storageFile, err := os.Open(cacheFiles + strconv.FormatInt(time, 10) + cachingStrategy + "store")

		if err != nil {
			panic(err)
		}

		defer storageFile.Close()

		csvr = csv.NewReader(storageFile)

		for line, err := csvr.Read(); err == io.EOF; line, err = csvr.Read() {

			node, err := strconv.Atoi(line[0])

			if err != nil {
				continue
			}

			item, err := strconv.Atoi(line[1])

			if err != nil {
				continue
			}

			storage[node] += itemSize[item]
		}

		// 3. sum up storage

		totalStorage := 0.0

		for _, cacheSize := range storage {
			totalStorage += cacheSize
		}

		// 4. convert to MB

		totalStorage = totalStorage / 1000 / 1000

		// 5. and write

		f.WriteString(strconv.FormatFloat(totalStorage, 'f', -1, 64))

		// Analyze Cache Hits

		f.WriteString(",")

		hits := 0
		totalRequests := 0

		hitsFile, err := os.Open(cacheFiles + strconv.FormatInt(time, 10) + cachingStrategy + "cache")

		if err != nil {
			panic(err)
		}

		defer hitsFile.Close()

		csvr = csv.NewReader(hitsFile)

		for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

			hit, err := strconv.ParseBool(line[1])

			if err != nil {
				continue
			}

			totalRequests++

			if hit {
				hits++
			}
		}

		ratio := float64(hits) / float64(totalRequests)

		f.WriteString(strconv.FormatFloat(ratio, 'f', -1, 64))

		// Analyze Hops
		f.WriteString(",")

		totalHops := 0
		totalRequests = 0

		hopsFile, err := os.Open(cacheFiles + strconv.FormatInt(time, 10) + cachingStrategy + "hops")

		if err != nil {
			panic(err)
		}

		defer hopsFile.Close()

		csvr = csv.NewReader(hopsFile)

		for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {

			hops, err := strconv.Atoi(line[1])

			if err != nil {
				continue
			}

			totalRequests++

			totalHops += hops
		}

		avg := float64(totalHops) / float64(totalRequests)

		f.WriteString(strconv.FormatFloat(avg, 'f', -1, 64))

		pbar.Add(1)

	}
}
