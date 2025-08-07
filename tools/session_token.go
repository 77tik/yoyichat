package tools

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"io"
	"time"
)

const SessionPrefix = "session_"

func GetSnowflakeId() string {
	node, _ := snowflake.NewNode(1)
	id := node.Generate().String()
	return id
}

func GetRandomToken(length int) string {
	r := make([]byte, length)
	io.ReadFull(rand.Reader, r)
	return base64.URLEncoding.EncodeToString(r)
}

// 由token 获取 会话key，和创建差不多
func CreateSessionId(sessionId string) string {
	return SessionPrefix + sessionId
}

// 由用户ID 获取 用户key，用户key可以获取token，然后拼接token就能拿到会话key
func GetSessionIdByUserId(userId int) string {
	return fmt.Sprintf("%s_map_%d", SessionPrefix, userId)
}

// 用token获取会话key，使用会话key就能拿到用户元信息
func GetSessionName(sessionId string) string {
	return SessionPrefix + sessionId
}

func Sha1(s string) (str string) {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func GetNowDateTime() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}
