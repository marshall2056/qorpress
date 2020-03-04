package helpers

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
)

func EncodeInterfaceToBytes(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func EncodeIntToBytes(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func Uin32tobytes(num uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(num))
	return b
}

func Bytestouint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func MarshalValue(value interface{}) ([]byte, error) {
	switch val := value.(type) {
	case []byte:
		return val, nil
	default:
		return json.Marshal(value)
	}
}
