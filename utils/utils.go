package utils

import (
	"fmt"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/satori/go.uuid"
	"github.com/symphonyprotocol/log"
)

var utilsLogger = log.GetLogger("utils")
var _mtx sync.RWMutex

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
	if data == nil || len(data) == 0 {
		return fmt.Errorf("Nil/Empty data (%v) cannot be converted to Diagram", data)
	}

	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return err
	}
	s, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(s, &diagram)
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
