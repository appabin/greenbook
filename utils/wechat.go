package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WeChatSessionResponse 微信会话响应结构
type WeChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetWeChatSession 通过code获取微信会话信息
func GetWeChatSession(appID, appSecret, code string) (*WeChatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sessionRes WeChatSessionResponse
	if err := json.Unmarshal(body, &sessionRes); err != nil {
		return nil, err
	}

	if sessionRes.ErrCode != 0 {
		return nil, errors.New(sessionRes.ErrMsg)
	}

	return &sessionRes, nil
}

// DecryptWeChatData 解密微信加密数据
func DecryptWeChatData(sessionKey, encryptedData, iv string) (map[string]interface{}, error) {
	// Base64解码
	key, _ := base64.StdEncoding.DecodeString(sessionKey)
	cipherText, _ := base64.StdEncoding.DecodeString(encryptedData)
	ivBytes, _ := base64.StdEncoding.DecodeString(iv)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherText, cipherText)

	// 去除填充
	pad := int(cipherText[len(cipherText)-1])
	decrypted := cipherText[:len(cipherText)-pad]

	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, err
	}

	return result, nil
}
