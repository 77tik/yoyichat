package logic

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
	"yoyichat/config"
	"yoyichat/pb/task_pb"
	"yoyichat/tools"
)

var RedisClient *redis.Client
var RedisSessClient *redis.Client

// 消息队列客户端初始化
func (l *Logic) InitPublishRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisCli Ping Result pong: %s,  err: %s", pong, err)
	}

	// 会话客户端通用
	RedisSessClient = RedisClient
	return err
}

// rpc 服务端 供给给api层调用，初始化
// 为配置中的每个地址:端口 都启动Server
func (l *Logic) InitRpcServer() (err error) {
	var network, addr string
	rpcAddressList := strings.Split(config.Conf.Logic.LogicBase.RpcAddress, ",")
	for _, bind := range rpcAddressList {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogicRpc ParseNetwork error : %s", err.Error())
		}
		logrus.Infof("logic start run at-->%s:%s", network, addr)
		go l.createRpcServer(network, addr)
	}
	return
}
func (l *Logic) createRpcServer(network, addr string) {
	s := server.NewServer() // rpcx 方法

	// 添加etcd注册
	l.addRegistryPlugin(s, network, addr)
	// ServerId 必须不一样
	// 注册到内存映射表中，没有超出代码
	// etcd中: /yoyichat/services/<ServerPathLogic>/<ServerId>
	// 例如: /yoyichat/services/logic-service/server-01
	// 注册进etcd 关键步骤
	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", l.ServerId))
	if err != nil {
		logrus.Errorf("Register logic failed err:%s", err.Error())
	}

	// 关闭时触发：
	//从内存中清除所有注册的方法
	//etcd 插件会自动删除该服务的注册节点
	//心跳停止后 etcd 租约到期自动清理

	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}

func (l *Logic) addRegistryPlugin(s *server.Server, network string, addr string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

// 单聊消息发布
func (l *Logic) RedisPublishSingleSend(serverId string, toUserId int, msg []byte) (err error) {
	redisMsg := task_pb.RedisMsg{
		Op:       config.OpSingleSend,
		ServerId: serverId,
		Msg:      msg,
		UserId:   int32(toUserId),
	}

	redisMsgBytes, err := proto.Marshal(&redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishChannel Marshal err:%s", err.Error())
		return err
	}

	// 将二进制数组push进队列下了吗？
	redisChannel := config.QueueName
	if err := RedisClient.LPush(redisChannel, redisMsgBytes).Err(); err != nil {
		logrus.Errorf("logic,RedisPublishChannel LPush err:%s", err.Error())
		return err
	}
	return
}

func (l *Logic) RedisPublishRoomSend(roomId int, count int, RoomUserInfo map[string]string, msg []byte) (err error) {
	var redisMsg = &task_pb.RedisMsg{
		Op:           config.OpRoomSend,
		RoomId:       int32(roomId),
		Count:        int32(count),
		Msg:          msg,
		RoomUserInfo: RoomUserInfo,
	}

	redisMsgBytes, err := proto.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(config.QueueName, redisMsgBytes).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

// 查询房间人数
func (l *Logic) RedisPublishRoomCount(roomId int, count int) (err error) {
	var redisMsg = &task_pb.RedisMsg{
		Op:     config.OpRoomCountSend,
		RoomId: int32(roomId),
		Count:  int32(count),
	}
	redisMsgBytes, err := proto.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(config.QueueName, redisMsgBytes).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	return
}

// 查询房间元信息
func (l *Logic) RedisPublishRoomInfo(roomId int, count int, roomUserInfo map[string]string) (err error) {
	var redisMsg = &task_pb.RedisMsg{
		Op:           config.OpRoomInfoSend,
		RoomId:       int32(roomId),
		Count:        int32(count),
		RoomUserInfo: roomUserInfo,
	}
	redisMsgBytes, err := proto.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.LPush(config.QueueName, redisMsgBytes).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

// 键命名规范
func (logic *Logic) getRoomUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getRoomOnlineCountKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
