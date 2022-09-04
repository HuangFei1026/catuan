package tengxun

import (
	"bytes"
	"catuan/components/http_client"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
)

type MiniApp struct {
	appid  string
	secret string
}

func NewMiniApp(appid, secret string) *MiniApp {
	return &MiniApp{
		appid:  appid,
		secret: secret,
	}
}

type RespCode2SessionInfo struct {
	Openid     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

func (m *MiniApp) Code2Session(code string, result *RespCode2SessionInfo) error {
	reqUrl := "https://api.weixin.qq.com/sns/jscode2session?appid=" + m.appid + "&secret=" + m.secret + "&js_code=" + code + "&grant_type=authorization_code"
	resp, err := http_client.HttpClient.Get(reqUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyData, result)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip":      "解析微信返回结果异常",
			"bodyData": string(bodyData),
		}).Error(err.Error())
		return err
	}
	return nil
}

type RespAccessTokenInfo struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}

func (m *MiniApp) GetAccessToken(result *RespAccessTokenInfo) error {
	reqUrl := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + m.appid + "&secret=" + m.secret
	resp, err := http_client.HttpClient.Get(reqUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyData, result)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip":      "access_token获取异常",
			"bodyData": string(bodyData),
		}).Error(err.Error())
		return err
	}
	return nil
}

type RespUserPhoneInfo struct {
	Errcode   int       `json:"errcode"`
	Errmsg    string    `json:"errmsg"`
	PhoneInfo PhoneInfo `json:"phone_info"`
}

type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

// GetPhone 获取电话号码
func (m *MiniApp) GetPhone(accessToken string, code string, result *RespUserPhoneInfo) error {
	reqUrl := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=" + accessToken
	codeData := `{"code":"` + code + `"}`
	resp, err := http_client.HttpClient.Post(reqUrl, "application/json", bytes.NewReader([]byte(codeData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyData, result)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip":      "解析微信返回结果异常",
			"bodyData": string(bodyData),
		}).Error(err.Error())
		return err
	}
	return nil
}

type QRCodeOptionInfo struct {
	Page      string    `json:"page"`
	Scene     string    `json:"scene"`
	Width     int       `json:"width"`
	AutoColor bool      `json:"auto_color,omitempty"`
	LineColor LineColor `json:"line_color,omitempty"`
	IsHyaline bool      `json:"is_hyaline,omitempty"`
}

type LineColor struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

func (m *MiniApp) GetQRCode(accessToken string, option QRCodeOptionInfo) ([]byte, error) {
	reqUrl := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + accessToken
	optionData, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}
	resp, err := http_client.HttpClient.Post(reqUrl, "application/json", bytes.NewReader(optionData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	imgBuff, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return imgBuff, nil
}
