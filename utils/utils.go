package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
)

func ZipDiagramToBytes(diagram interface{}) []byte {
	data, err := json.Marshal(diagram)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write(data)
	gz.Flush()
	gz.Close()
	return b.Bytes()
}

func UnzipDiagramBytes(data []byte) []byte {
	rdata := bytes.NewReader(data)
	r, _ := gzip.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)
	return s
}
