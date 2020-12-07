package main

import (
	"bufio"
	"os"
	"strconv"
)

type fw struct {
	cacheStrategy string
	filename      string
}

func newFileWriter(filename string, c <-chan writeSet) {

	f := fw{
		filename: filename,
	}

	for w := range c {

		if w.time == -1 {
			break
		}

		f.write(w.time, w.strategyName, w.txRecords, w.storeRecords, w.cacheRecords, w.hopsRecords)
	}
}

func (f *fw) write(time int64, strategyName string, txRecords *[]txRecord, storeRecords *[]storeRecord, cacheRecords *[]cacheRecord, hopsRecords *[]hopsRecord) {
	baseFilename := f.filename + strconv.FormatInt(time, 10) + strategyName
	f.writeTX(baseFilename+"tx", txRecords)
	f.writeStore(baseFilename+"store", storeRecords)
	f.writeCache(baseFilename+"cache", cacheRecords)
	f.writeHops(baseFilename+"hops", hopsRecords)
}

func (f *fw) writeTX(filename string, records *[]txRecord) {
	txFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer txFile.Close()

	buf := bufio.NewWriter(txFile)

	buf.WriteString("source,target,bandwidth\n")

	for _, r := range *records {
		source := r.source
		target := r.target

		if source > target {
			source = r.target
			target = r.source
		}

		buf.WriteString(strconv.FormatInt(source, 10))
		buf.WriteString(",")
		buf.WriteString(strconv.FormatInt(target, 10))
		buf.WriteString(",")
		buf.WriteString(strconv.FormatInt(r.bandwidth, 10))
		buf.WriteString("\n")
	}

	buf.Flush()
}

func (f *fw) writeStore(filename string, records *[]storeRecord) {
	storeFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer storeFile.Close()

	buf := bufio.NewWriter(storeFile)

	buf.WriteString("node,item\n")

	for _, r := range *records {
		buf.WriteString(strconv.FormatInt(r.node, 10))
		buf.WriteString(",")
		buf.WriteString(strconv.FormatInt(r.item, 10))
		buf.WriteString("\n")
	}

	buf.Flush()
}

func (f *fw) writeCache(filename string, records *[]cacheRecord) {
	cacheFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer cacheFile.Close()

	buf := bufio.NewWriter(cacheFile)

	buf.WriteString("item,success\n")

	for _, r := range *records {

		buf.WriteString(strconv.FormatInt(r.item, 10))
		buf.WriteString(",")
		buf.WriteString(strconv.FormatBool(r.success))
		buf.WriteString("\n")
	}

	buf.Flush()
}

func (f *fw) writeHops(filename string, records *[]hopsRecord) {
	hopsFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}

	defer hopsFile.Close()

	buf := bufio.NewWriter(hopsFile)

	buf.WriteString("item,hops\n")

	for _, r := range *records {
		buf.WriteString(strconv.FormatInt(r.item, 10))
		buf.WriteString(",")
		buf.WriteString(strconv.FormatInt(r.hops, 10))
		buf.WriteString("\n")
	}

	buf.Flush()
}
