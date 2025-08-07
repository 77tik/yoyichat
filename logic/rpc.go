package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"strconv"
	"time"
	"yoyichat/config"
	"yoyichat/logic/dao"
	"yoyichat/pb/logic_pb"
	"yoyichat/pb/task_pb"
	"yoyichat/tools"
)

// logic 层全都注册进rpc中

type RpcLogic struct{}

func (rpc *RpcLogic) Register(ctx context.Context, args *logic_pb.RegisterRequest, reply *logic_pb.RegisterReply) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	uData := u.CheckHaveUserName(args.Name)
	if uData.Id > 0 {
		return errors.New("this user name already have , please login !!!")
	}
	u.UserName = args.Name
	u.Password = args.Password
	userId, err := u.Add()
	if err != nil {
		logrus.Infof("register err:%s", err.Error())
		return err
	}
	if userId == 0 {
		return errors.New("register userId empty!")
	}

	// 设置会话令牌
	randToken := tools.GetRandomToken(32)
	sessionId := tools.CreateSessionId(randToken)
	userData := make(map[string]interface{})
	userData["userId"] = userId
	userData["userName"] = uData.UserName
	RedisSessClient.Do("MULTI")
	RedisSessClient.HMSet(sessionId, userData)
	RedisSessClient.Expire(sessionId, config.RedisBaseValidTime*time.Second)
	err = RedisSessClient.Do("EXEC").Err()
	if err != nil {
		logrus.Infof("register set redis token fail!")
		return err
	}

	// 操作成功，返回认证令牌
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}

// 1.通过用户key 拿到token，2.然后拼接token 为会话key，3.通过会话key就能拿到用户元信息
func (rpc *RpcLogic) Login(ctx context.Context, request *logic_pb.LoginRequest, reply *logic_pb.LoginResponse) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	username := request.Name
	password := request.Password

	data := u.CheckHaveUserName(username)
	if (data.Id == 0) || (password != data.Password) {
		return errors.New("username or password error")
	}

	// 获取会话ID
	loginSessionId := tools.GetSessionIdByUserId(data.Id)
	randToken := tools.GetRandomToken(32)
	sessionId := tools.CreateSessionId(randToken)
	userData := make(map[string]interface{})
	userData["userId"] = data.Id
	userData["userName"] = data.UserName

	token, _ := RedisSessClient.Get(loginSessionId).Result()

	if token != "" {
		oldSession := tools.CreateSessionId(token)
		err := RedisSessClient.Del(oldSession).Err()
		if err != nil {
			return errors.New("logout user fail token is:" + token)
		}
	}
	RedisSessClient.Do("MULTI")
	RedisSessClient.HMSet(sessionId, userData)
	RedisSessClient.Expire(sessionId, config.RedisBaseValidTime*time.Second)
	RedisSessClient.Set(loginSessionId, randToken, config.RedisBaseValidTime*time.Second)
	err = RedisSessClient.Do("EXEC").Err()
	if err != nil {
		logrus.Infof("login set redis token fail!")
		return err
	}

	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}

// 用用户Id获取用户
func (rpc *RpcLogic) GetUserInfoByUserId(ctx context.Context, req *logic_pb.GetUserInfoRequest, reply *logic_pb.GetUserInfoResponse) (err error) {
	reply.Code = config.FailReplyCode
	userId := req.UserId
	u := new(dao.User)
	userName := u.GetUserNameByUserId(int(userId))
	reply.UserName = userName
	reply.UserId = userId
	reply.Code = config.SuccessReplyCode
	return
}

// 用会话key拿到用户元信息，然后比对
func (rpc *RpcLogic) CheckAuth(ctx context.Context, req *logic_pb.CheckAuthRequest, reply *logic_pb.CheckAuthResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := req.AuthToken

	sessionName := tools.GetSessionName(authToken)

	userDataMap, err := RedisSessClient.HGetAll(sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail!， authToken is: %s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session! authToken is: %s", authToken)
		return
	}

	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	reply.Code = config.SuccessReplyCode
	reply.UserId = int32(intUserId)
	userName, _ := userDataMap["userName"]
	reply.UserName = userName
	return
}

func (rpc *RpcLogic) Logout(ctx context.Context, req *logic_pb.LogoutRequest, reply *logic_pb.LogoutResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := req.AuthToken
	sessionName := tools.GetSessionName(authToken)
	userDataMap, err := RedisSessClient.HGetAll(sessionName).Result()
	if err != nil {
		logrus.Infof("logout fail! authToken is: %s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session! authToken is: %s", authToken)
		return
	}

	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	SessIdMap := tools.GetSessionIdByUserId(intUserId)
	// 删掉用户key
	err = RedisSessClient.Del(SessIdMap).Err()
	if err != nil {
		logrus.Infof("logout del token error ! authToken is: %s", authToken)
		return err
	}

	logic := new(Logic)
	// yoyichat_user01，删掉对应的这个是什么用户key？？
	serverIdKey := logic.getUserKey(fmt.Sprintf("%d", intUserId))
	err = RedisSessClient.Del(serverIdKey).Err()
	if err != nil {
		logrus.Infof("logout del server id error:%s", err.Error())
		return err
	}

	// 删掉会话
	err = RedisSessClient.Del(sessionName).Err()
	if err != nil {
		logrus.Infof("logout error:%s", err.Error())
		return err
	}
	reply.Code = config.SuccessReplyCode
	return
}

// 在logic层构造请求发送到task层，接收task层回复
func (rpc *RpcLogic) Push(ctx context.Context, req *logic_pb.SendMsg, reply *task_pb.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := req
	bodyBytes, err := proto.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic layer push msg fail !!! err: %s", err.Error())
		return
	}
	// 获取接收者所在的Connection服务器层，这个存在redis中
	logic := new(Logic)
	userSidKey := logic.getUserKey(fmt.Sprintf("%d", sendData.ToUserId))
	serverIdStr := RedisSessClient.Get(userSidKey).Val()

	if err != nil {
		logrus.Errorf("logic,push parse int fail:%s", err.Error())
		return
	}

	// 推送到对应的队列中
	err = logic.RedisPublishSingleSend(serverIdStr, int(sendData.ToUserId), bodyBytes)
	if err != nil {
		logrus.Errorf("logic,redis publish err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) PushRoom(ctx context.Context, req *logic_pb.SendMsg, reply *task_pb.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := req
	roomId := sendData.RoomId
	logic := new(Logic)
	roomUserInfo := make(map[string]string)
	// yoyichat_room_room01
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(int(roomId)))
	roomUserInfo, err = RedisClient.HGetAll(roomUserKey).Result()
	if err != nil {
		logrus.Errorf("logic,PushRoom redis hGetAll err:%s", err.Error())
		return
	}
	// 我明白了，他是向把单聊消息改为群发
	var bodyBytes []byte
	sendData.RoomId = roomId
	sendData.Msg = req.Msg
	sendData.FromUserId = req.FromUserId
	sendData.FromUserName = req.FromUserName
	sendData.Op = config.OpRoomSend
	sendData.CreateTime = tools.GetNowDateTime()
	bodyBytes, err = proto.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic,PushRoom Marshal err:%s", err.Error())
		return
	}
	err = logic.RedisPublishRoomSend(int(roomId), len(roomUserInfo), roomUserInfo, bodyBytes)
	if err != nil {
		logrus.Errorf("logic,PushRoom err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

/*
*
get room online person count 获取房间在线人数信息
*/
func (rpc *RpcLogic) Count(ctx context.Context, args *logic_pb.SendMsg, reply *task_pb.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	roomId := args.RoomId
	logic := new(Logic)
	var count int
	count, err = RedisSessClient.Get(logic.getRoomOnlineCountKey(fmt.Sprintf("%d", roomId))).Int()
	err = logic.RedisPublishRoomCount(int(roomId), count)
	if err != nil {
		logrus.Errorf("logic,Count err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

/*
*
get room info
*/
func (rpc *RpcLogic) GetRoomInfo(ctx context.Context, args *logic_pb.SendMsg, reply *task_pb.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(int(roomId)))
	roomUserInfo, err = RedisClient.HGetAll(roomUserKey).Result()
	if len(roomUserInfo) == 0 {
		return errors.New("getRoomInfo no this user")
	}
	err = logic.RedisPublishRoomInfo(int(roomId), len(roomUserInfo), roomUserInfo)
	if err != nil {
		logrus.Errorf("logic,GetRoomInfo err:%s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

// 加入房间
func (rpc *RpcLogic) Connect(ctx context.Context, args *logic_pb.ConnectRequest, reply *logic_pb.ConnectReply) (err error) {
	if args == nil {
		logrus.Errorf("logic,connect args empty")
		return
	}

	// 验证会话
	logic := new(Logic)
	//key := logic.getUserKey(args.AuthToken)
	logrus.Infof("logic,authToken is:%s", args.AuthToken)
	key := tools.GetSessionName(args.AuthToken)
	userInfo, err := RedisClient.HGetAll(key).Result()
	if err != nil {
		logrus.Infof("RedisCli HGetAll key :%s , err:%s", key, err.Error())
		return err
	}
	if len(userInfo) == 0 {
		reply.UserId = 0
		return
	}
	userId, _ := strconv.Atoi(userInfo["userId"])
	reply.UserId = int32(userId)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(int(args.RoomId)))
	if reply.UserId != 0 {
		userKey := logic.getUserKey(fmt.Sprintf("%d", reply.UserId))
		logrus.Infof("logic redis set userKey:%s, serverId : %s", userKey, args.ServerId)
		validTime := config.RedisBaseValidTime * time.Second

		// 记录用户 - 服务器 映射
		err = RedisClient.Set(userKey, args.ServerId, validTime).Err()
		if err != nil {
			logrus.Warnf("logic set err:%s", err)
		}

		// 加入房间，人数加1，房间记录新用户
		if RedisClient.HGet(roomUserKey, fmt.Sprintf("%d", reply.UserId)).Val() == "" {
			RedisClient.HSet(roomUserKey, fmt.Sprintf("%d", reply.UserId), userInfo["userName"])
			// add room user count ++
			RedisClient.Incr(logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId)))
		}
	}
	logrus.Infof("logic rpc userId:%d", reply.UserId)
	return
}

// 离开房间
func (rpc *RpcLogic) DisConnect(ctx context.Context, args *logic_pb.DisConnectRequest, reply *logic_pb.DisConnectReply) (err error) {
	logic := new(Logic)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(int(args.RoomId)))
	// room user count -- 更新在线人数
	if args.RoomId > 0 {
		count, _ := RedisSessClient.Get(logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Int()
		if count > 0 {
			RedisClient.Decr(logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Result()
		}
	}
	// room login user-- 将用户从map中移除
	if args.UserId != 0 {
		err = RedisClient.HDel(roomUserKey, fmt.Sprintf("%d", args.UserId)).Err()
		if err != nil {
			logrus.Warnf("HDel getRoomUserKey err : %s", err)
		}
	}
	//below code can optimize send a signal to queue,another process get a signal from queue,then push event to websocket
	// 下方代码可优化为：发送信号到队列，再由另一个进程从队列获取信号并推送事件到WebSocket
	// 但是我看来这其实就已经推送到队列中，这个logic层的逻辑从来不自己处理逻辑，都是推送到队列中
	roomUserInfo, err := RedisClient.HGetAll(roomUserKey).Result()
	if err != nil {
		logrus.Warnf("RedisCli HGetAll roomUserInfo key:%s, err: %s", roomUserKey, err)
	}
	if err = logic.RedisPublishRoomSend(int(args.RoomId), len(roomUserInfo), roomUserInfo, nil); err != nil {
		logrus.Warnf("publish RedisPublishRoomCount err: %s", err.Error())
		return
	}
	return
}
