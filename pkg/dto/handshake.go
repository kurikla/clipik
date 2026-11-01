package dto

const (
	Version = "1.0.0"
	Protocol = "CLIPIK"
)

type Format string

const (
	FormatJSON    Format = "json"
	FormatMsgpack Format = "msgpack"
)

func (f Format) Valid() bool {
	switch f {
	case FormatJSON, FormatMsgpack:
		return true
	}
	return false
}

type Handshake struct {
	Protocol  string `json:"protocol"`
	Version   string `json:"version"`
	Format    Format `json:"format"`
	ClientPub []byte `json:"client_pub"`
	ClientID  []byte `json:"client_id"`
	ClientSig []byte `json:"client_sig"`
}

type HandshakeAck struct {
	Protocol     string `json:"protocol"`
	Version      string `json:"version"`
	Format       Format `json:"format"`
	ServerPub    []byte `json:"server_pub"`
	ServerID     []byte `json:"server_id"`
	ServerSig    []byte `json:"server_sig"`
	SessionToken string `json:"session_token"`
}
