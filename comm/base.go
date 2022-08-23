package comm

type KeyAble interface {
	uint8 | int8 | int16 | int32 | int64 | uint16 | uint32 | uint64 | string
}

type NumberAble interface {
	uint8 | int8 | int16 | int32 | int64 | uint16 | uint32 | uint64
}

type RespResult struct {
	ErrCode int
	ErrMsg  string
	Data    any
}
