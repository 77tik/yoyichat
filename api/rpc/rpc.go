package rpc

import (
	"context"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"yoyichat/config"
	"yoyichat/pb/logic_pb"
	"yoyichat/pb/task_pb"

	"sync"
	"time"
)

var LogicRpcClient client.XClient
var once sync.Once

type RpcLogic struct {
}

var RpcLogicObj *RpcLogic

func InitLogicRpcClient() {
	once.Do(func() {
		etcdConfigOption := &store.Config{
			ClientTLS:         nil,
			TLS:               nil,
			ConnectionTimeout: time.Duration(config.Conf.Common.CommonEtcd.ConnectionTimeout) * time.Second,
			Bucket:            "",
			PersistConnection: true,
			Username:          config.Conf.Common.CommonEtcd.Username,
			Password:          config.Conf.Common.CommonEtcd.Password,
		}
		d, err := etcdV3.NewEtcdV3Discovery(
			config.Conf.Common.CommonEtcd.BasePath,
			config.Conf.Common.CommonEtcd.ServerPathLogic,
			[]string{config.Conf.Common.CommonEtcd.Host},
			true,
			etcdConfigOption,
		)
		if err != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail:%s", err.Error())
		}
		LogicRpcClient = client.NewXClient(config.Conf.Common.CommonEtcd.ServerPathLogic, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		RpcLogicObj = new(RpcLogic)
	})
	if LogicRpcClient == nil {
		logrus.Fatalf("get logic rpc client nil")
	}
}

func (rpc *RpcLogic) Login(req *logic_pb.LoginRequest) (code int, authToken string, msg string) {
	reply := &logic_pb.LoginResponse{}
	err := LogicRpcClient.Call(context.Background(), "Login", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = int(reply.Code)
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) Register(req *logic_pb.RegisterRequest) (code int, authToken string, msg string) {
	reply := &logic_pb.RegisterReply{}
	err := LogicRpcClient.Call(context.Background(), "Register", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = int(reply.Code)
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) GetUserNameByUserId(req *logic_pb.GetUserInfoRequest) (code int, userName string) {
	reply := &logic_pb.GetUserInfoResponse{}
	LogicRpcClient.Call(context.Background(), "GetUserInfoByUserId", req, reply)
	code = int(reply.Code)
	userName = reply.UserName
	return
}

func (rpc *RpcLogic) CheckAuth(req *logic_pb.CheckAuthRequest) (code int, userId int, userName string) {
	reply := &logic_pb.CheckAuthResponse{}
	LogicRpcClient.Call(context.Background(), "CheckAuth", req, reply)
	code = int(reply.Code)
	userId = int(reply.UserId)
	userName = reply.UserName
	return
}

func (rpc *RpcLogic) Logout(req *logic_pb.LogoutRequest) (code int) {
	reply := &logic_pb.LogoutResponse{}
	LogicRpcClient.Call(context.Background(), "Logout", req, reply)
	code = int(reply.Code)
	return
}

func (rpc *RpcLogic) Push(req *logic_pb.SendMsg) (code int, msg string) {
	reply := &task_pb.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "Push", req, reply)
	code = int(reply.Code)
	msg = reply.Msg
	return
}

func (rpc *RpcLogic) PushRoom(req *logic_pb.SendMsg) (code int, msg string) {
	reply := &task_pb.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "PushRoom", req, reply)
	code = int(reply.Code)
	msg = reply.Msg
	return
}

func (rpc *RpcLogic) Count(req *logic_pb.SendMsg) (code int, msg string) {
	reply := &task_pb.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "Count", req, reply)
	code = int(reply.Code)
	msg = reply.Msg
	return
}

func (rpc *RpcLogic) GetRoomInfo(req *logic_pb.SendMsg) (code int, msg string) {
	reply := &task_pb.SuccessReply{}
	LogicRpcClient.Call(context.Background(), "GetRoomInfo", req, reply)
	code = int(reply.Code)
	msg = reply.Msg
	return
}
