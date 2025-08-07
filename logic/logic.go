package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"runtime"
	"yoyichat/config"
)

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	//read config
	logicConfig := config.Conf.Logic

	runtime.GOMAXPROCS(logicConfig.LogicBase.CpuNum)
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())
	//init publish redis 这应该才是rpc => 消息队列的rpc客户端才对
	if err := logic.InitPublishRedisClient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail,err:%s", err.Error())
	}

	//init rpc server 这里是logic => 消息队列的rpc吗？ 不对，应该是作为api => logic的rpc服务器
	// 没想到吧，其实是connect层调用的
	// 还有个问题，它是怎么把服务注册到etcd上的？
	// 破案了，在正下方这个初始化里面就已经把服务注册进etcd了
	// 这个只是注册方法，所以如果没有人调用那岂不是就尬在这里了
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init rpc server fail")
	}
}
