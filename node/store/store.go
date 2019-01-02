package store

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/symphonyprotocol/p2p/config"
	"log"
	"strings"
)

func GetLocalNodeKeyStr() (privKey string, pubKey string) {
	bytes, _ := getData("LocalNodeKey")
	value := string(bytes)
	if len(value) > 0 {
		vals := strings.Split(value, "#")
		if len(vals) != 2 {
			return "", ""
		}
		return vals[0], vals[1]
	}
	return "", ""
}

func SaveLocalNodeKey(privKey string, pubKey string) error {
	value := privKey + "#" + pubKey
	return saveData("LocalNodeKey", []byte(value))
}

func getData(key string) ([]byte, error) {
	db, err := leveldb.OpenFile(config.LEVEL_DB_FILE, nil)
	if err != nil {
		log.Fatalf("cannot open leveldb:", err)
	}
	defer db.Close()
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Printf("GetLocalNodePrivateKeyStr error:%v\n", err)
		return nil, err
	}
	return data, err
}

func saveData(key string, value []byte) error {
	db, err := leveldb.OpenFile(config.LEVEL_DB_FILE, nil)
	if err != nil {
		log.Fatalf("cannot open leveldb:", err)
	}
	defer db.Close()
	err = db.Put([]byte(key), value, nil)
	if err != nil {
		log.Printf("GetLocalNodePrivateKeyStr error:%v\n", err)
		return err
	}
	return nil
}
