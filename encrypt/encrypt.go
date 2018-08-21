package encrypt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/big"
)

var (
	E_CURVE = elliptic.P256()
)

func GenerateNodeKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(E_CURVE, rand.Reader)
	if err != nil {
		log.Fatalf("GenerateNodeKey error:%v", err)
		return nil
	}
	return privKey
}

func FromPrivateKey(privKey *ecdsa.PrivateKey) string {
	return hex.EncodeToString(privKey.D.Bytes())
}

func ToPrivateKey(privKeyStr string, pubKey ecdsa.PublicKey) *ecdsa.PrivateKey {
	bytes, _ := hex.DecodeString(privKeyStr)
	d := new(big.Int)
	d.SetBytes(bytes)
	return &ecdsa.PrivateKey{
		PublicKey: pubKey,
		D:         d,
	}
}

func FromPublicKey(pubKey ecdsa.PublicKey) string {
	bytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	return hex.EncodeToString(bytes)
}

func ToPublicKey(pubKeyStr string) ecdsa.PublicKey {
	bytes, _ := hex.DecodeString(pubKeyStr)
	x, y := elliptic.Unmarshal(E_CURVE, bytes)
	return ecdsa.PublicKey{
		Curve: E_CURVE,
		X:     x,
		Y:     y,
	}
}

func PublicKeyToNodeId(pubKey ecdsa.PublicKey) []byte {
	bytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)
	id := make([]byte, 64)
	if len(bytes)-1 != len(id) {
		panic("err")
	}
	copy(id[:], bytes[1:])
	return ToSha256(id)
}

func ToSha256(source []byte) []byte {
	h := sha256.New()
	h.Write(source)
	bytes := h.Sum(nil)
	return bytes
}
