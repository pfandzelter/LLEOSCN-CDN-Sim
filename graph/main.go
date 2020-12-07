package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/pelletier/go-toml"
	"github.com/schollz/progressbar/v3"
)

func getTx(txFile string, cacheFiles string, strategies []string, steps int64, stepLength int64) {

	attr := []string{"total", "max", "min", "avg", "median", "95th", "99th"}

	bufs := make(map[string]*bufio.Writer)

	for _, a := range attr {
		f, err := os.Create(txFile + a + ".csv")

		if err != nil {
			panic(err)
		}

		defer f.Close()
		buf := bufio.NewWriter(f)

		buf.WriteString("time")

		for _, s := range strategies {
			buf.WriteString(",")
			buf.WriteString(s)
		}

		buf.WriteString("\n")

		bufs[a] = buf
		buf.Flush()
	}

	pbar := progressbar.Default(steps)

	var time int64

	for ; time < steps*stepLength; time += stepLength {
		ts := strconv.FormatInt(time, 10)
		for _, buf := range bufs {
			buf.WriteString(ts)
		}

		for _, s := range strategies {
			for _, buf := range bufs {
				buf.WriteString(",")
			}

			file := cacheFiles + ts + s + "tx"
			c, err := os.Open(file)

			if err != nil {
				panic(err)
			}

			csvr := csv.NewReader(c)

			for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {
				bufs[line[0]].WriteString(line[1])
			}

			c.Close()
		}

		for _, buf := range bufs {
			buf.WriteString("\n")
			buf.Flush()
		}
		pbar.Add(1)

	}
}

func getStore(storeFile string, cacheFiles string, strategies []string, steps int64, stepLength int64) {

	attr := []string{"total", "max", "min", "avg", "median", "95th", "99th", "numnodes", "numnostorenodes"}

	bufs := make(map[string]*bufio.Writer)

	for _, a := range attr {
		f, err := os.Create(storeFile + a + ".csv")

		if err != nil {
			panic(err)
		}

		defer f.Close()
		buf := bufio.NewWriter(f)

		buf.WriteString("time")

		for _, s := range strategies {
			buf.WriteString(",")
			buf.WriteString(s)
		}

		buf.WriteString("\n")

		bufs[a] = buf
		buf.Flush()
	}

	pbar := progressbar.Default(steps)

	var time int64

	for ; time < steps*stepLength; time += stepLength {
		ts := strconv.FormatInt(time, 10)
		for _, buf := range bufs {
			buf.WriteString(ts)
		}

		for _, s := range strategies {
			for _, buf := range bufs {
				buf.WriteString(",")
			}

			file := cacheFiles + ts + s + "store"
			c, err := os.Open(file)

			if err != nil {
				panic(err)
			}

			csvr := csv.NewReader(c)

			for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {
				bufs[line[0]].WriteString(line[1])
			}

			c.Close()
		}

		for _, buf := range bufs {
			buf.WriteString("\n")
			buf.Flush()
		}
		pbar.Add(1)

	}
}

func getCache(cacheFile string, cacheFiles string, strategies []string, steps int64, stepLength int64) {

	attr := []string{"ratio", "num_requests"}

	bufs := make(map[string]*bufio.Writer)

	for _, a := range attr {
		f, err := os.Create(cacheFile + a + ".csv")

		if err != nil {
			panic(err)
		}

		defer f.Close()
		buf := bufio.NewWriter(f)

		buf.WriteString("time")

		for _, s := range strategies {
			buf.WriteString(",")
			buf.WriteString(s)
		}

		buf.WriteString("\n")

		bufs[a] = buf
		buf.Flush()
	}

	pbar := progressbar.Default(steps)

	var time int64

	for ; time < steps*stepLength; time += stepLength {
		ts := strconv.FormatInt(time, 10)
		for _, buf := range bufs {
			buf.WriteString(ts)
		}

		for _, s := range strategies {
			for _, buf := range bufs {
				buf.WriteString(",")
			}

			file := cacheFiles + ts + s + "cache"
			c, err := os.Open(file)

			if err != nil {
				panic(err)
			}

			csvr := csv.NewReader(c)

			for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {
				bufs[line[0]].WriteString(line[1])
			}

			c.Close()
		}

		for _, buf := range bufs {
			buf.WriteString("\n")
			buf.Flush()
		}
		pbar.Add(1)

	}
}

func getHops(hopsFile string, cacheFiles string, strategies []string, steps int64, stepLength int64) {

	attr := []string{"total", "max", "min", "avg", "median", "95th", "99th"}

	bufs := make(map[string]*bufio.Writer)

	for _, a := range attr {
		f, err := os.Create(hopsFile + a + ".csv")

		if err != nil {
			panic(err)
		}

		defer f.Close()
		buf := bufio.NewWriter(f)

		buf.WriteString("time")

		for _, s := range strategies {
			buf.WriteString(",")
			buf.WriteString(s)
		}

		buf.WriteString("\n")

		bufs[a] = buf
		buf.Flush()
	}

	pbar := progressbar.Default(steps)

	var time int64

	for ; time < steps*stepLength; time += stepLength {
		ts := strconv.FormatInt(time, 10)
		for _, buf := range bufs {
			buf.WriteString(ts)
		}

		for _, s := range strategies {
			for _, buf := range bufs {
				buf.WriteString(",")
			}

			file := cacheFiles + ts + s + "hops"
			c, err := os.Open(file)

			if err != nil {
				panic(err)
			}

			csvr := csv.NewReader(c)

			for line, err := csvr.Read(); err != io.EOF; line, err = csvr.Read() {
				bufs[line[0]].WriteString(line[1])
			}

			c.Close()
		}

		for _, buf := range bufs {
			buf.WriteString("\n")
			buf.Flush()
		}
		pbar.Add(1)

	}
}

func main() {
	if len(os.Args) != 2 {
		panic("not enough arguments given")
	}

	strategies := []string{"NONE", "GROUND-STATION-1000000000", "GROUND-STATION-10000", "GROUND-STATION-100", "GROUND-STATION-10", "SATELLITE", "SATELLITE-TIMEOUT", "SATELLITE-VIRTUAL"}

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
	stepLength := workloadConfig.Get("step_length").(int64)

	cacheFiles := path.Join(workloadFolder, "cache", "c.csv")

	err = os.MkdirAll(path.Join(workloadFolder, "data"), os.ModePerm)

	if err != nil {
		panic(err)
	}

	dataFiles := path.Join(workloadFolder, "data", "data.csv")

	getTx(dataFiles+"tx", cacheFiles, strategies, steps, stepLength)
	getStore(dataFiles+"store", cacheFiles, strategies, steps, stepLength)
	getCache(dataFiles+"cache", cacheFiles, strategies, steps, stepLength)
	getHops(dataFiles+"hops", cacheFiles, strategies, steps, stepLength)

}
