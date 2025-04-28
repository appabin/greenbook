package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/appabin/greenbook/config"
)

// Code2Session 微信登录凭证校验
// WechatCode2SessionResponse 微信登录凭证校验返回结果
type WechatCode2SessionResponse struct {
	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 用户在开放平台的唯一标识符
	ErrCode    int    `json:"errcode"`     // 错误码
	ErrMsg     string `json:"errmsg"`      // 错误信息
}

func Code2Session(code string) (*WechatCode2SessionResponse, error) {
	appID := config.AppConfig.Wechat.AppID
	appSecret := config.AppConfig.Wechat.AppSecret

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wxResp WechatCode2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		return nil, err
	}

	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误：%s", wxResp.ErrMsg)
	}

	return &wxResp, nil
}
