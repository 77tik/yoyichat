package config

import "sync"

var once sync.Once
var Conf *Config

const (
	SuccessReplyCode      = 0
	FailReplyCode         = 1
	SuccessReplyMsg       = "success"
	QueueName             = "yoyichat_queue"
	RedisBaseValidTime    = 86400 // 这是有效时间吗，难道是Redis中消息队列的有效时间？
	RedisPrefix           = "yoyichat_"
	RedisRoomPrefix       = "yoyichat_room_"
	RedisRoomOnlinePrefix = "yoyichat_room_online_count_"
	MsgVersion            = 1
	OpSingleSend          = 2 // single user
	OpRoomSend            = 3 // send to room
	OpRoomCountSend       = 4 // get online user count
	OpRoomInfoSend        = 5 // send info to room
	OpBuildTcpConn        = 6 // build tcp conn
)

type Config struct {
}
