package helpers

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

func DecodeBytesToInt(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func DecodeBytesToInterface(b []byte, dest interface{}) error {
	return gob.NewDecoder(bytes.NewReader(b)).Decode(dest)
}
