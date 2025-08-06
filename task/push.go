package task

type PushParams struct {
	ServerId string // 与接受者连接的Connection层
	UserId   int    // 接受者
	Msg      []byte
	RoomId   int
}

var pushChannel []chan *PushParams

func init() {
	pushChannel = make([]chan *PushParams, 0)
}
