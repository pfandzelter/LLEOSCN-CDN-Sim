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
	"os"
	"path"
	"strconv"

	"github.com/pelletier/go-toml"
	"github.com/schollz/progressbar/v3"
)

func main() {

	strategies := []string{"NONE", "GROUND-STATION", "SATELLITE", "SATELLITE-TIMEOUT", "SATELLITE-VIRTUAL"}

	maxClientsPerGST := []int64{10000, 100, 10}

	if len(os.Args) != 2 {
		panic("not enough arguments given")
	}

	conf := os.Args[1]

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
	//steps := int64(3)
	stepLength := workloadConfig.Get("step_length").(int64)
	numRequest := int(workloadConfig.Get("requestamount").(int64))

	err = os.MkdirAll(path.Join(workloadFolder, "cache"), os.ModePerm)

	if err != nil {
		panic(err)
	}

	cacheFiles := path.Join(workloadFolder, "cache", "c.csv")

	resultFiles := path.Join(workloadFolder, "results", "r.csv")

	loadFile := path.Join(workloadFolder, workloadConfig.Get("loadfile").(string))

	cityFile := path.Join(workloadFolder, workloadConfig.Get("cities").(string))

	itemSizes := getItemSizes(loadFile)

	C := make([]strategy, len(strategies)+len(maxClientsPerGST)-1)

	i := 0
	for _, s := range strategies {
		switch s {
		case "NONE":
			C[i] = newNone()
		case "GROUND-STATION":
			gstPopulation := getGSTPopulation(cityFile)
			for _, m := range maxClientsPerGST {
				C[i] = newGroundstation(m, *gstPopulation)
				i++
			}
			i = i - 1
		case "SATELLITE":
			C[i] = newSatellite()
		case "SATELLITE-VIRTUAL":
			C[i] = newSatelliteVirtual(itemSizes)
		case "SATELLITE-TIMEOUT":
			C[i] = newSatelliteTimeout(itemSizes)
		default:
			panic("Unknown caching strategy: " + s)
		}

		i++
	}

	fileWriteC := make(chan writeSet)

	storeNodesPerStrategy := make(map[string]int64)

	for _, c := range C {
		storeNodesPerStrategy[c.getName()] = c.getStoreNodes()
	}

	go newAvgWriter(cacheFiles, itemSizes, fileWriteC, storeNodesPerStrategy)
	// go newFileWriter(cacheFiles, fileWriteC)

	pbar := progressbar.Default(steps)

	var time int64 = 0

	coord := make(chan struct{}, len(C))

	for range C {
		coord <- struct{}{}
	}

	for ; time < steps*stepLength; time += stepLength {

		// 1. read shortest_sat_paths
		shortestSatPaths := getShortestSatPaths(resultFiles + strconv.FormatInt(time, 10) + "shortest_sat_paths")

		// 2. read gnd_sat_links
		gndSatLinks := getGroundSatLinks(resultFiles + strconv.FormatInt(time, 10) + "gnd_sat_links")

		// 3. read paths/requests
		requests := getRequests(resultFiles+strconv.FormatInt(time, 10)+"paths", numRequest)

		for range C {
			<-coord
		}

		for i, c := range C {

			go func(strategy string, cache *strategy, c *chan struct{}, t int64) {
				// 5. pass variables to caching strategy
				txRecords, storeRecords, cacheRecords, hopsRecords := (*cache).stepTo(t, shortestSatPaths, gndSatLinks, requests)

				// 6. write returns
				fileWriteC <- writeSet{
					time:         t,
					strategyName: strategy,
					txRecords:    txRecords,
					storeRecords: storeRecords,
					cacheRecords: cacheRecords,
					hopsRecords:  hopsRecords,
				}

				*c <- struct{}{}

			}(c.getName(), &C[i], &coord, time)
		}

		pbar.Add(1)
	}

	// only finish when everything is done
	for range C {
		<-coord
	}

	fileWriteC <- writeSet{time: -1}
}
