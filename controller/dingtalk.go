package controller

import (
	"errors"
	"path"
	"strings"

	"github.com/fufuok/utils"
	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-message-center/conf"
	"github.com/fufuok/xy-message-center/middleware"
	"github.com/fufuok/xy-message-center/service/dingtalk"
)

// 上传媒体文件
func V1DDMediaUploadHandler(c *gin.Context) {
	mediaType := c.PostForm("type")
	file, err := c.FormFile("media")
	if err != nil {
		middleware.APIFailure(c, "文件上传失败", err.Error())
		return
	}

	mediaConf, ok := conf.Config.DD.MediaConf[mediaType]
	if !ok {
		mediaConf = conf.Config.DD.MediaConf["file"]
	}

	// 文件大小限制
	if file.Size > mediaConf.Size {
		middleware.APIFailure(c, "文件太大", "")
		return
	}

	// 文件格式检查, 仅检查后缀名
	if ext := path.Ext(file.Filename); ext == "" || !strings.Contains(mediaConf.Ext, ext) {
		middleware.APIFailure(c, "文件格式不支持", mediaConf.Ext)
		return
	}

	// 尝试打开文件
	mediaFile, err := file.Open()
	if err != nil {
		middleware.APIFailure(c, "文件格式异常", err.Error())
		return
	}
	defer func() {
		_ = mediaFile.Close()
	}()

	mediaID, err := dingtalk.MediaUpload(file.Filename, mediaType, mediaFile)
	if err != nil {
		middleware.APIFailure(c, "资源上传失败", err.Error())
		return
	}

	// 日志
	c.Set(conf.LogDataName, map[string]interface{}{
		"media_type":   mediaType,
		"media_name":   file.Filename,
		"media_size":   file.Size,
		"media_header": file.Header,
		"media_id":     mediaID,
	})

	middleware.APISuccess(c, mediaID, 0)
}

// 创建会话
func V1DDChatCreateHandler(c *gin.Context) {
	var data struct {
		Name       string   `json:"name"`
		Owner      string   `json:"owner"`
		UserIDList []string `json:"useridlist"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		middleware.APIFailure(c, "参数格式有误", err.Error())
		return
	}

	chatID, err := dingtalk.CreateChat(data.Name, data.Owner, data.UserIDList)
	if err != nil {
		middleware.APIFailure(c, "创建会话失败", err.Error())
		return
	}

	// 日志
	c.Set(conf.LogDataName, utils.MustJSONString(data))

	middleware.APISuccess(c, chatID, 0)
}

// 发送消息到企业群
func V1DDChatSendHandler(c *gin.Context) {
	var data struct {
		ChatID string                 `json:"chatid"`
		Msg    map[string]interface{} `json:"msg"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		middleware.APIFailure(c, "参数格式有误", err.Error())
		return
	}

	if errMsg, err := checkMessageData(data.Msg); err != nil {
		middleware.APIFailure(c, errMsg, err.Error())
		return
	}

	messageID, err := dingtalk.SendChatMessage(data.ChatID, data.Msg)
	if err != nil {
		middleware.APIFailure(c, "群消息发送失败", err.Error())
		return
	}

	// 日志
	c.Set(conf.LogDataName, utils.MustJSONString(data))

	middleware.APISuccess(c, messageID, 0)
}

// 发送工作通知消息
func V1DDTopAPIMessageHandler(c *gin.Context) {
	var data struct {
		UserList string                 `json:"userid_list"`
		DeptList string                 `json:"dept_id_list"`
		ToAll    bool                   `json:"to_all_user"`
		Msg      map[string]interface{} `json:"msg"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		middleware.APIFailure(c, "参数格式有误", err.Error())
		return
	}
	if errMsg, err := checkMessageData(data.Msg); err != nil {
		middleware.APIFailure(c, errMsg, err.Error())
		return
	}

	taskID, err := dingtalk.SendTopAPIMessage(data.UserList, data.DeptList, data.ToAll, data.Msg)
	if err != nil {
		middleware.APIFailure(c, "工作通知发送失败", err.Error())
		return
	}

	// 日志
	c.Set(conf.LogDataName, utils.MustJSONString(data))

	middleware.APISuccess(c, taskID, 0)
}

// 校验消息体
func checkMessageData(msg map[string]interface{}) (string, error) {
	msgtype, ok := msg["msgtype"]
	if !ok {
		return "消息类型无效: [msgtype]", errors.New("miss msgtype")
	}
	msgType := msgtype.(string)
	_, ok = msg[msgType]
	if !ok {
		return "消息内容无效: [" + msgType + "]", errors.New("miss content")
	}

	return "", nil
}
