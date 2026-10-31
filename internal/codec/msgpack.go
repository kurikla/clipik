package codec

import (
	"github.com/vmihailenco/msgpack/v5"
)

type Msgpack struct {}

func (Msgpack) Encode(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (Msgpack) Decode(b []byte, v any) error {
	return msgpack.Unmarshal(b, v)
}
