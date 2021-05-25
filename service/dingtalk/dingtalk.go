package dingtalk

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/imroc/req"

	"github.com/fufuok/xy-message-center/conf"
)

type ErrResult struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

var (
	accessTokenExpires = 7200
	timingAdvance      = 200
	baseUrl            = "https://oapi.dingtalk.com"
)

func init() {
	expires, err := RefreshAccessToken()
	if err != nil || conf.Config.DD.AccessToken == "" {
		log.Fatalln("Failed to refresh AccessToken:", err, "\nbye.")
	}

	// 保活, 有效期(默认为 7200 秒)内重复获取返回相同结果, 并会自动续期
	go func(expires int) {
		for range time.Tick(time.Duration(expires) * time.Second) {
			if _, err := RefreshAccessToken(); err != nil {
				for range time.Tick(5 * time.Second) {
					if _, err = RefreshAccessToken(); err == nil {
						break
					}
				}
			}
		}
	}(expires - timingAdvance)
}

// 获取 AccessToken
// https://ding-doc.dingtalk.com/doc#/serverapi2/eev437
func RefreshAccessToken() (int, error) {
	apiUrl := baseUrl + "/gettoken"
	params := req.Param{
		"appkey":    conf.Config.DD.AppKey,
		"appsecret": conf.Config.DD.AppSecret,
	}
	res, err := req.Get(apiUrl, params)
	if err != nil {
		return 0, err
	}

	var data struct {
		ErrResult
		AccessToken string `json:"access_token"`
		Expires     int    `json:"expires_in"`
	}
	err = res.ToJSON(&data)
	if err == nil {
		if data.AccessToken == "" {
			return 0, fmt.Errorf("%s", res.String())
		}
		conf.Config.DD.AccessToken = data.AccessToken
		if data.Expires < timingAdvance+1 {
			data.Expires = accessTokenExpires
		}
	}

	return data.Expires, err
}

// 创建会话
// https://developers.dingtalk.com/document/app/create-group-session
func CreateChat(name, owner string, userIDList []string) (string, error) {
	apiUrl := baseUrl + "/chat/create?access_token=" + conf.Config.DD.AccessToken
	params := req.Param{
		"name":       name,
		"owner":      owner,
		"useridlist": userIDList,
	}
	res, err := req.Post(apiUrl, req.BodyJSON(&params), conf.ReqUserAgent)
	if err != nil {
		return "", err
	}

	var data struct {
		ErrResult
		ChatID string `json:"chatid"`
	}
	err = res.ToJSON(&data)
	if err == nil {
		if data.ErrCode == 0 {
			return data.ChatID, nil
		}
		return "", errors.New(fmt.Sprintf("[%d] %s", data.ErrCode, data.ErrMsg))
	}

	return "", err
}

// 发送工作通知消息
// https://developers.dingtalk.com/document/app/asynchronous-sending-of-enterprise-session-messages
func SendTopAPIMessage(userIDList, deptIDList string, toAll bool, msg map[string]interface{}) (int, error) {
	apiUrl := baseUrl + "/topapi/message/corpconversation/asyncsend_v2?access_token=" + conf.Config.DD.AccessToken
	params := req.Param{
		"agent_id": conf.Config.DD.AgentID,
		"msg":      msg,
	}
	if toAll {
		params["to_all_user"] = true
	} else {
		if userIDList != "" {
			params["userid_list"] = userIDList
		}
		if deptIDList != "" {
			params["dept_id_list"] = deptIDList
		}
	}

	if len(params) < 3 {
		return 0, errors.New("请求参数有误")
	}

	res, err := req.Post(apiUrl, req.BodyJSON(&params), conf.ReqUserAgent)
	if err != nil {
		return 0, err
	}

	var data struct {
		ErrResult
		TaskID int `json:"task_id"`
	}
	err = res.ToJSON(&data)
	if err == nil {
		if data.ErrCode == 0 {
			return data.TaskID, nil
		}
		return 0, errors.New(fmt.Sprintf("[%d] %s", data.ErrCode, data.ErrMsg))
	}

	return 0, err
}

// 发送消息到企业群
// https://developers.dingtalk.com/document/app/send-group-messages
func SendChatMessage(chatID string, msg map[string]interface{}) (string, error) {
	apiUrl := baseUrl + "/chat/send?access_token=" + conf.Config.DD.AccessToken
	params := req.Param{
		"access_token": conf.Config.DD.AccessToken,
		"chatid":       chatID,
		"msg":          msg,
	}
	res, err := req.Post(apiUrl, req.BodyJSON(&params), conf.ReqUserAgent)
	if err != nil {
		return "", err
	}

	var data struct {
		ErrResult
		MessageId string `json:"messageId"`
	}
	err = res.ToJSON(&data)
	if err == nil {
		if data.ErrCode == 0 {
			return data.MessageId, nil
		}
		return "", errors.New(fmt.Sprintf("[%d] %s", data.ErrCode, data.ErrMsg))
	}

	return "", err
}

// 上传图片, 语音, 文件
// https://developers.dingtalk.com/document/app/upload-media-files
func MediaUpload(mediaName, mediaType string, mediaFile multipart.File) (string, error) {
	apiUrl := baseUrl + "/media/upload?access_token=" + conf.Config.DD.AccessToken
	params := req.Param{
		"type": mediaType,
	}
	media := req.FileUpload{
		File:      mediaFile,
		FieldName: "media",
		FileName:  mediaName,
	}
	res, err := req.Post(apiUrl, params, media, conf.ReqUserAgent)
	if err != nil {
		return "", err
	}

	var data struct {
		ErrResult
		MediaID string `json:"media_id"`
	}
	err = res.ToJSON(&data)
	if err == nil {
		if data.ErrCode == 0 {
			return data.MediaID, nil
		}
		return "", errors.New(fmt.Sprintf("[%d] %s", data.ErrCode, data.ErrMsg))
	}

	return "", err
}
