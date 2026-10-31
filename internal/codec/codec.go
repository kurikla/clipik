package codec

import (
	"fmt"
	"github.com/kurikla/clipik-s/pkg/dto"
)

type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(b []byte, v any) error
}

func New(f dto.Format) (Codec, error) {
	switch f {
	case dto.FormatJSON:
		return JSON{}, nil
	case dto.FormatMsgpack:
		return Msgpack{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %q", f)
	}
}
