package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"wishes/middleware"
	"wishes/models"

	"gorm.io/gorm"
)

type WechatService struct {
	DB        *gorm.DB
	AppID     string
	AppSecret string
	JWTSecret []byte
}

func NewWechatService(db *gorm.DB, appId string, appSecret string, jwtSecret []byte) *WechatService {
	return &WechatService{
		DB:        db,
		AppID:     appId,
		AppSecret: appSecret,
		JWTSecret: jwtSecret,
	}
}

type WechatLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (s *WechatService) Login(code string) (string, *models.User, error) {
	loginResp, err := s.Code2Session(code)
	if err != nil {
		return "", nil, err
	}

	if loginResp.ErrCode != 0 {
		return "", nil, fmt.Errorf("微信登录失败: %s", loginResp.ErrMsg)
	}

	var user models.User
	result := s.DB.Where("wechat_openid = ?", loginResp.OpenID).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			WechatOpenID: loginResp.OpenID,
		}
		if loginResp.UnionID != "" {
			user.WechatUnionID = loginResp.UnionID
		}

		if err := s.DB.Create(&user).Error; err != nil {
			return "", nil, err
		}
	} else if result.Error != nil {
		return "", nil, result.Error
	}

	token, err := middleware.GenerateUserToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

func (s *WechatService) Code2Session(code string) (*WechatLoginResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.AppID,
		s.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loginResp WechatLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}

func (s *WechatService) UpdateUserInfo(userID uint, nickname, avatarURL string) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}

	return s.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"nickname":   nickname,
		"avatar_url": avatarURL,
	}).Error
}

func (s *WechatService) UpdateAdminInfo(adminID uint, nickname, avatarURL string, department string) error {
	if adminID == 0 {
		return errors.New("管理员ID不能为空")
	}

	return s.DB.Model(&models.Admin{}).Where("id = ?", adminID).Updates(map[string]interface{}{
		"nickname":   nickname,
		"avatar_url": avatarURL,
		"department": department,
	}).Error
}
