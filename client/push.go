package client

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"yoyichat/logic/dao"
)

// api层的push请求体
type formPush struct {
	Msg       string `form:"msg" json:"msg" binding:"required"`
	ToUserId  string `form:"toUserId" json:"toUserId" binding:"required"`
	RoomId    int    `form:"roomId" json:"roomId" binding:"required"`
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

// 发送消息
func (m *model) pushMessage() {
	content := strings.TrimSpace(m.input.Value())
	if content == "" {
		return
	}

	// 处理退出命令
	if content == "/exit" {
		os.Exit(0)
	}

	// 处理私聊消息
	recipient := ""
	if strings.HasPrefix(content, "@") {
		parts := strings.SplitN(content[1:], " ", 2)
		if len(parts) == 2 {
			recipient = parts[0]
			content = parts[1]
		}
	}

	// 本地显示：
	msg := Message{
		Sender:    m.username,
		Recipient: recipient,
		Content:   content,
		Timestamp: time.Now(),
	}

	// 请求体发送：
	toUserIdStr := strconv.Itoa(new(dao.User).GetUserIdByUserName(m.recipient))
	fp := &formPush{
		Msg:       content,
		ToUserId:  toUserIdStr,
		RoomId:    0,
		AuthToken: "",
	}
	msgData, _ := json.Marshal(fp)

	// 通过 HTTP 发送消息
	// TODO: api层接口调用
	go func() {
		req, _ := http.NewRequest("POST", apiBase+pushPath, strings.NewReader(string(msgData)))
		req.Header.Set("Authorization", "Bearer "+m.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()

	// 如果是私聊消息，在UI中保存
	if recipient != "" {
		m.recipient = recipient
	}

	// 添加消息到本地显示

	m.messages = append(m.messages, msg)
	m.input.SetValue("")
}

// 群聊消息
type formRoom struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
	Msg       string `form:"msg" json:"msg" binding:"required"`
	RoomId    int    `form:"roomId" json:"roomId" binding:"required"`
}
