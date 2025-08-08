package config

import (
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

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

// 差个站点层
type Config struct {
	Common  Common
	Connect ConnectConfig
	Logic   LogicConfig
	Task    TaskConfig
	Api     ApiConfig
	Client  ClientConfig
}

func init() {
	Init()
}

func getCurDir() string {
	_, filename, _, _ := runtime.Caller(1)
	aPath := strings.Split(filename, "/")
	dir := strings.Join(aPath[:len(aPath)-1], "/")
	return dir
}

func GetMode() string {
	env := os.Getenv("RUN_MODE")
	if env == "" {
		return "dev"
	}
	return env
}

func Init() {
	once.Do(func() {
		env := GetMode()

		dir := getCurDir()

		configFilePath := path.Join(dir, env)
		// 指定配置文件格式
		viper.SetConfigType("toml")
		// 添加配置文件搜索路径
		viper.AddConfigPath(configFilePath)
		// 指定文件名 1
		viper.SetConfigName("/connect")
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("/common")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("/task")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("/logic")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("/api")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName("/client")
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		Conf = new(Config)
		viper.Unmarshal(&Conf.Connect)
		viper.Unmarshal(&Conf.Common)
		viper.Unmarshal(&Conf.Api)
		viper.Unmarshal(&Conf.Logic)
		viper.Unmarshal(&Conf.Task)
		viper.Unmarshal(&Conf.Client)

	})
}

// 获取GIN运行Mode
func GetGinRunMode() string {
	env := GetMode()
	//gin have debug,test,release mode
	if env == "dev" {
		return "debug"
	}
	if env == "test" {
		return "debug"
	}
	if env == "prod" {
		return "release"
	}
	return "release"
}

type CommonEtcd struct {
	Host              string `mapstructure:"host"`
	BasePath          string `mapstructure:"basePath"`
	ServerPathLogic   string `mapstructure:"serverPathLogic"`
	ServerPathConnect string `mapstructure:"serverPathConnect"`
	Username          string `mapstructure:"userName"`
	Password          string `mapstructure:"password"`
	ConnectionTimeout int    `mapstructure:"connectionTimeout"`
}

type CommonRedis struct {
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	Db            int    `mapstructure:"db"`
}

type Common struct {
	CommonEtcd  CommonEtcd  `mapstructure:"common-etcd"`
	CommonRedis CommonRedis `mapstructure:"common-redis"`
}

// 这是干啥的
type ConnectBase struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}

type ConnectRpcAddressWebsockts struct {
	Address string `mapstructure:"address"`
}

type ConnectRpcAddressTcp struct {
	Address string `mapstructure:"address"`
}

type ConnectBucket struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	Channel       int    `mapstructure:"channel"`
	Room          int    `mapstructure:"room"`
	SrvProto      int    `mapstructure:"srvProto"`
	RoutineAmount uint64 `mapstructure:"routineAmount"` // 我记得这好像是叫广播协程数量
	RoutineSize   int    `mapstructure:"routineSize"`
}

type ConnectWebsocket struct {
	ServerId string `mapstructure:"serverId"`
	Bind     string `mapstructure:"bind"`
}

type ConnectTcp struct {
	ServerId      string `mapstructure:"serverId"`
	Bind          string `mapstructure:"bind"`
	SendBuf       int    `mapstructure:"sendbuf"`
	ReceiveBuf    int    `mapstructure:"receivebuf"`
	KeepAlive     bool   `mapstructure:"keepalive"`
	Reader        int    `mapstructure:"reader"`
	ReadBuf       int    `mapstructure:"readBuf"`
	ReadBufSize   int    `mapstructure:"readBufSize"`
	Writer        int    `mapstructure:"writer"`
	WriterBuf     int    `mapstructure:"writerBuf"`
	WriterBufSize int    `mapstructure:"writeBufSize"`
}

type ConnectConfig struct {
	ConnectBase                ConnectBase                `mapstructure:"connect-base"`
	ConnectRpcAddressWebSockts ConnectRpcAddressWebsockts `mapstructure:"connect-rpcAddress-websockts"`
	ConnectRpcAddressTcp       ConnectRpcAddressTcp       `mapstructure:"connect-rpcAddress-tcp"`
	ConnectBucket              ConnectBucket              `mapstructure:"connect-bucket"`
	ConnectWebsocket           ConnectWebsocket           `mapstructure:"connect-websocket"`
	ConnectTcp                 ConnectTcp                 `mapstructure:"connect-tcp"`
}

type LogicBase struct {
	ServerId   string `mapstructure:"serverId"`
	CpuNum     int    `mapstructure:"cpuNum"`
	RpcAddress string `mapstructure:"rpcAddress"`
	CertPath   string `mapstructure:"certPath"`
	KeyPath    string `mapstructure:"keyPath"`
}

type LogicConfig struct {
	LogicBase LogicBase `mapstructure:"logic-base"`
}

type TaskBase struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	RedisAddr     string `mapstructure:"redisAddr"`
	RedisPassword string `mapstructure:"redisPassword"`
	RpcAddress    string `mapstructure:"rpcAddress"`
	PushChan      int    `mapstructure:"pushChan"`
	PushChanSize  int    `mapstructure:"pushChanSize"`
}

type TaskConfig struct {
	TaskBase TaskBase `mapstructure:"task-base"`
}

type ApiBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

type ApiConfig struct {
	ApiBase ApiBase `mapstructure:"api-base"`
}

type ClientBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

type ClientConfig struct {
	ClientBase ClientBase `mapstructure:"client-base"`
}
