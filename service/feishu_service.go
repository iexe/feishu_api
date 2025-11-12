package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"oapi-sdk-go-demo/config"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuService struct {
	client *lark.Client
	config *config.Config
}

// 用户信息结构体
type UserInfo struct {
	UserID        string     `json:"user_id"`
	OpenID        string     `json:"open_id"`
	UnionID       string     `json:"union_id"`
	Mobile        string     `json:"mobile"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	DepartmentIds []string   `json:"department_ids"`
	Status        UserStatus `json:"status"`
}

// 用户状态
type UserStatus struct {
	IsFrozen    bool `json:"is_frozen"`
	IsResigned  bool `json:"is_resigned"`
	IsActivated bool `json:"is_activated"`
	IsExited    bool `json:"is_exited"`
	IsUnjoin    bool `json:"is_unjoin"`
}

func NewFeishuService(cfg *config.Config) *FeishuService {
	client := lark.NewClient(cfg.AppID, cfg.AppSecret)
	return &FeishuService{
		client: client,
		config: cfg,
	}
}

// BatchGetUserIds 根据手机号或邮箱批量获取用户ID信息
func (s *FeishuService) BatchGetUserIds(phoneNumbers []string, emails []string, includeResigned bool) (*larkcontact.BatchGetIdUserRespData, error) {
	// 使用飞书通讯录API批量获取用户ID
	req := larkcontact.NewBatchGetIdUserReqBuilder().
		UserIdType("user_id").
		Body(larkcontact.NewBatchGetIdUserReqBodyBuilder().
			Emails(emails).
			Mobiles(phoneNumbers).
			IncludeResigned(includeResigned).
			Build()).
		Build()

	resp, err := s.client.Contact.User.BatchGetId(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("batch get user ids failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// SearchUserByPhone 根据手机号搜索用户
func (s *FeishuService) SearchUserByPhone(phoneNumber string) ([]*UserInfo, error) {
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	data, err := s.BatchGetUserIds([]string{phoneNumber}, []string{}, false)
	if err != nil {
		return nil, err
	}

	var users []*UserInfo
	if data != nil && data.UserList != nil {
		for _, userContact := range data.UserList {
			user := &UserInfo{
				UserID: getStringValue(userContact.UserId),
				Mobile: getStringValue(userContact.Mobile),
				Email:  getStringValue(userContact.Email),
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// UploadImage 上传图片到飞书并返回image_key
func (s *FeishuService) UploadImage(image io.Reader) (*larkim.CreateImageRespData, error) {
	req := larkim.NewCreateImageReqBuilder().
		Body(larkim.NewCreateImageReqBodyBuilder().
			ImageType("message").
			Image(image).
			Build()).
		Build()

	resp, err := s.client.Im.Image.Create(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("upload image request failed: %v", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("upload image failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// UploadFile 上传文件到飞书并返回file_key
func (s *FeishuService) UploadFile(fileType, fileName string, file io.Reader, duration int) (*larkim.CreateFileRespData, error) {
	req := larkim.NewCreateFileReqBuilder().
		Body(larkim.NewCreateFileReqBodyBuilder().
			FileType(fileType).
			FileName(fileName).
			Duration(duration).
			File(file).
			Build()).
		Build()

	resp, err := s.client.Im.File.Create(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("upload file failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// SendTextMessage 发送文本消息
func (s *FeishuService) SendTextMessage(receiveIdType, receiveId, content string) (*larkim.CreateMessageRespData, error) {
	// 构建消息内容
	msgContent := map[string]interface{}{
		"text": content,
	}
	contentBytes, _ := json.Marshal(msgContent)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveId).
			MsgType("text").
			Content(string(contentBytes)).
			Build()).
		Build()

	resp, err := s.client.Im.Message.Create(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("send message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// SendImageMessage 发送图片消息
func (s *FeishuService) SendImageMessage(receiveIdType, receiveId, imageKey string) (*larkim.CreateMessageRespData, error) {
	msgContent := map[string]interface{}{
		"image_key": imageKey,
	}
	contentBytes, _ := json.Marshal(msgContent)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveId).
			MsgType("image").
			Content(string(contentBytes)).
			Build()).
		Build()

	resp, err := s.client.Im.Message.Create(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("send image message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// SendFileMessage 发送文件消息
func (s *FeishuService) SendFileMessage(receiveIdType, receiveId, fileKey string) (*larkim.CreateMessageRespData, error) {
	msgContent := map[string]interface{}{
		"file_key": fileKey,
	}
	contentBytes, _ := json.Marshal(msgContent)

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveId).
			MsgType("file").
			Content(string(contentBytes)).
			Build()).
		Build()

	resp, err := s.client.Im.Message.Create(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("send file message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// 辅助函数定义

// getStringValue 安全获取字符串指针的值
func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// getBoolValue 安全获取布尔指针的值
func getBoolValue(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

// getDepartmentIds 安全获取部门ID列表
func getDepartmentIds(ids []*string) []string {
	var result []string
	if ids != nil {
		for _, id := range ids {
			if id != nil {
				result = append(result, *id)
			}
		}
	}
	return result
}