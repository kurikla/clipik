package dto

type Ping struct {
	Ts int64 `json:"ts" msgpack:"ts"`
}

type Pong struct {
	PingTs int64 `json:"ping_ts" msgpack:"ping_ts"`
	PongTs int64 `json:"pong_ts" msgpack:"pong_ts"`
}
