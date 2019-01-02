package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"

	"github.com/satori/go.uuid"
	"github.com/symphonyprotocol/log"
)

var utilsLogger = log.GetLogger("utils")

func ZipDiagramToBytes(diagram interface{}) []byte {
	data, err := json.Marshal(diagram)
	if err != nil {
		utilsLogger.Fatal("%v", err)
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

func BytesToUDPDiagram(data []byte, diagram interface{}) error {
	rdata := bytes.NewReader(data)
	r, _ := gzip.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)
	err := json.Unmarshal(s, &diagram)
	if err != nil {
		utilsLogger.Fatal("%v", err)
	}

	return err
}

func NewUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		u, err = uuid.NewV1()
	}
	return u.String()
}
