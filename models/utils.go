package models

import (
	"github.com/decred/base58"
	subkey "github.com/vedhavyas/go-subkey"
)

func SS58Address(addr []byte) (string, error) {
	return subkey.SS58Address(addr, 42) // substrate
}

func SS58Addr(addr []byte) (out string) {
	out, _ = subkey.SS58Address(addr, 42) // substrate
	return
}

func DecodeSS58Address(ss58addr string) (addr []byte) {
	decoded := base58.Decode(ss58addr)
	addr = decoded[1 : len(decoded)-2]
	return
}
