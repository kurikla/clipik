package codec

import "encoding/json"

type JSON struct{}

func (JSON) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSON) Decode(b []byte, v any) error {
	return json.Unmarshal(b, v)
}
