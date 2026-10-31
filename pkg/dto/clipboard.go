package dto

type ClipboardType string

const (
	ClipboardTypeData ClipboardType = "data"
	ClipboardTypeFile ClipboardType = "file"
)

type ClipboardFile struct {
	ID   string  `json:"id" msgpack:"id"`
	Name string  `json:"name" msgpack:"name"`
}

type Clipboard struct {
	Type ClipboardType `json:"type" msgpack:"type"`
	Data *[]byte `json:"data" msgpack:"data"`
	File *ClipboardFile `json:"file" msgpack:"file"`
}
