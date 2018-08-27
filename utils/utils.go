package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/satori/go.uuid"
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

func DiagramToBytes(diagram interface{}) []byte {
	data, _ := json.Marshal(diagram)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte(data))
	gz.Flush()
	gz.Close()
	return b.Bytes()
}

func BytesToUDPDiagram(data []byte, diagram interface{}) {
	rdata := bytes.NewReader(data)
	r, _ := gzip.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)
	err := json.Unmarshal(s, &diagram)
	if err != nil {
		log.Fatal(err)
	}
}

func NewUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		u, err = uuid.NewV1()
	}
	return u.String()
}
