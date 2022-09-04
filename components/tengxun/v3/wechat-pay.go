package v3

import (
	"bytes"
	"catuan/components/http_client"
	"catuan/util"
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

type WechatPay struct {
	Appid               string
	MchId               string
	MchKey              string
	NotifyUrl           string
	CertKeyFile         string
	CertFile            string
	MchCertSerialNumber string
	Client              *core.Client
	rsaKey              *rsa.PrivateKey
}

type ReqJsapi struct {
	Appid       string      `json:"appid"`
	Mchid       string      `json:"mchid"`
	Description string      `json:"description"`
	OutTradeNo  string      `json:"out_trade_no"`
	TimeExpire  string      `json:"time_expire,omitempty"`
	Attach      string      `json:"attach,omitempty"`
	NotifyUrl   string      `json:"notify_url"`
	GoodsTag    string      `json:"goods_tag,omitempty"`
	Amount      *amountInfo `json:"amount"`
	Payer       *payerInfo  `json:"payer,omitempty"`
	SceneInfo   *sceneInfo  `json:"scene_info,omitempty"`
}

type amountInfo struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type payerInfo struct {
	Openid string `json:"openid"`
}

type sceneInfo struct {
	PayerClientIp string  `json:"payer_client_ip,omitempty"`
	DeviceId      string  `json:"device_id,omitempty"`
	H5Info        *h5Info `json:"h5_info,omitempty"`
}

type h5Info struct {
	Type string `json:"type"`
}

type JsApiPayer struct {
	Description string
	OutTradeNo  string
	Total       int64
	Openid      string
}

type JsApiCallSign struct {
	Appid     string `json:"appId"`
	Timestamp string `json:"timestamp"`
	NonceStr  string `json:"nonceStr"` //随机字符串
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

func (w *WechatPay) GetClient() (*core.Client, error) {
	if w.Client == nil {
		mchPrivateKey, err := w.GetRsaPrivateKey()
		if err != nil {
			return nil, err
		}
		opts := []core.ClientOption{
			option.WithWechatPayAutoAuthCipher(w.MchId, w.MchCertSerialNumber, mchPrivateKey, w.MchKey),
		}
		client, err := core.NewClient(context.Background(), opts...)
		if err != nil {
			return nil, errors.New("初始化微信支付客户端失败")
		}
		w.Client = client
	}
	return w.Client, nil
}

func (w *WechatPay) loadPrivateKey() (*rsa.PrivateKey, error) {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(w.CertKeyFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip": "加载证书Key文件失败",
		}).Error(err.Error())
		return nil, errors.New("加载证书Key文件失败")
	}
	w.rsaKey = mchPrivateKey
	return mchPrivateKey, nil
}

func (w *WechatPay) GetRsaPrivateKey() (*rsa.PrivateKey, error) {
	if w.rsaKey == nil {
		rsaKey, err := w.loadPrivateKey()
		if err != nil {
			return nil, err
		}
		w.rsaKey = rsaKey
	}
	return w.rsaKey, nil
}

func (w *WechatPay) Jsapi(payInfo JsApiPayer) (*JsApiCallSign, error) {
	reqInfo := ReqJsapi{
		Appid:       w.Appid,
		Mchid:       w.MchId,
		Description: payInfo.Description,
		OutTradeNo:  payInfo.OutTradeNo,
		NotifyUrl:   w.NotifyUrl,
		Amount: &amountInfo{
			Total:    int(payInfo.Total),
			Currency: "CNY",
		},
		Payer: &payerInfo{
			Openid: payInfo.Openid,
		},
	}
	client, err := w.GetClient()
	if err != nil {
		return nil, err
	}
	reqUrl := "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"
	resp, _ := client.Post(context.TODO(), reqUrl, reqInfo)
	if resp.Response == nil {
		return nil, errors.New("微信JSAPI支付请求失败")
	}
	bodyData, err := io.ReadAll(resp.Response.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Response.Body.Close()
	prepayInfo := struct {
		PrepayId string `json:"prepay_id"`
		Code     string `json:"code"`
		Message  string `json:"message"`
	}{}
	err = json.Unmarshal(bodyData, &prepayInfo)
	if err != nil {
		return nil, err
	}
	if prepayInfo.PrepayId == "" {
		return nil, errors.New(prepayInfo.Message)
	}
	signInfo, err := w.jsApiCallSign(prepayInfo.PrepayId)
	if err != nil {
		return nil, err
	}
	return signInfo, nil
}

func (w *WechatPay) jsApiCallSign(prepayId string) (*JsApiCallSign, error) {
	jsApiCallInfo := JsApiCallSign{
		Appid:     w.Appid,
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		NonceStr:  util.RandStr(24, false),
		Package:   "prepay_id=" + prepayId,
		SignType:  "RSA",
	}
	//计算签名
	signStr := w.Appid + "\n" + jsApiCallInfo.Timestamp + "\n" + jsApiCallInfo.NonceStr + "\n" + jsApiCallInfo.Package + "\n"
	rsaKey, err := w.GetRsaPrivateKey()
	if err != nil {
		return nil, err
	}
	s, err := utils.SignSHA256WithRSA(signStr, rsaKey)
	if err != nil {
		return nil, err
	}
	jsApiCallInfo.PaySign = s
	return &jsApiCallInfo, nil
}

type ReqRefundInfo struct {
	TransactionId string            `json:"transaction_id,omitempty"`
	OutTradeNo    string            `json:"out_trade_no,omitempty"`
	OutRefundNo   string            `json:"out_refund_no"`
	Reason        string            `json:"reason"`
	NotifyUrl     string            `json:"notify_url,omitempty"`
	Amount        *refundAmountInfo `json:"amount"`
}

type refundAmountInfo struct {
	Refund   int    `json:"refund"`
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type RefundInfo struct {
	TransactionId string
	OutRefundNo   string
	Reason        string
	RefundAmount  int
	TotalAmount   int
}

// Refund 申请退款
func (w *WechatPay) Refund(refundInfo *RefundInfo) error {
	reqInfo := ReqRefundInfo{
		TransactionId: refundInfo.TransactionId,
		OutRefundNo:   refundInfo.OutRefundNo,
		Reason:        refundInfo.Reason,
		Amount: &refundAmountInfo{
			Refund: refundInfo.RefundAmount,
			Total:  refundInfo.TotalAmount,
		},
	}
	client, err := w.GetClient()
	if err != nil {
		return err
	}
	reqUrl := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	resp, _ := client.Post(context.TODO(), reqUrl, reqInfo)
	if resp.Response == nil {
		return errors.New("微信退款请求失败")
	}
	bodyData, err := io.ReadAll(resp.Response.Body)
	if err != nil {
		return err
	}
	defer resp.Response.Body.Close()
	respInfo := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}{}
	err = json.Unmarshal(bodyData, &respInfo)
	if err != nil {
		return err
	}
	if respInfo.Code != "" {
		return errors.New(respInfo.Message)
	}
	return nil
}

// ReqTransferToBalance 企业付款到零钱请求参数
type ReqTransferToBalance struct {
	Appid              string               `json:"appid"`
	OutBatchNo         string               `json:"out_batch_no"`
	BatchName          string               `json:"batch_name"`
	BatchRemark        string               `json:"batch_remark"`
	TotalAmount        int64                `json:"total_amount"`
	TotalNum           int64                `json:"total_num"`
	TransferDetailList []transferDetailInfo `json:"transfer_detail_list"`
}

type transferDetailInfo struct {
	OutDetailNo    string `json:"out_detail_no"`
	TransferAmount int    `json:"transfer_amount"`
	TransferRemark string `json:"transfer_remark"`
	Openid         string `json:"openid"`
	UserName       string `json:"user_name,omitempty"`
}

type TransferUserInfo struct {
	Openid   string `json:"openid"`
	UserName string `json:"user_name,omitempty"`
	Amount   int    `json:"amount"`
}

type RespTransferToBalance struct {
	OutBatchNo string `json:"out_batch_no"`
	BatchId    string `json:"batch_id"`
	CreateTime string `json:"create_time"`
}

// ToBalance 商家付款到零钱
func (w *WechatPay) ToBalance(title string, cashUserList []TransferUserInfo, result *RespTransferToBalance) error {
	totalNum := len(cashUserList)
	totalAmount := 0
	transferDetailList := make([]transferDetailInfo, 0, len(cashUserList))
	for _, item := range cashUserList {
		totalAmount = totalAmount + item.Amount
		transferDetailList = append(transferDetailList, transferDetailInfo{
			OutDetailNo:    w.createOutBatchNo("i"),
			TransferAmount: item.Amount,
			TransferRemark: title,
			Openid:         item.Openid,
			UserName:       item.UserName,
		})
	}

	reqInfo := ReqTransferToBalance{}
	reqInfo.Appid = w.Appid
	reqInfo.OutBatchNo = w.createOutBatchNo("o")
	reqInfo.BatchName = title
	reqInfo.BatchRemark = title
	reqInfo.TotalAmount = int64(totalAmount)
	reqInfo.TotalNum = int64(totalNum)
	reqInfo.TransferDetailList = transferDetailList

	reqUrl := "https://api.mch.weixin.qq.com/v3/transfer/batches"

	//reqData, _ := json.Marshal(reqInfo)
	client, err := w.GetClient()
	if err != nil {
		return err
	}

	respApi, err := client.Post(context.TODO(), reqUrl, reqInfo)

	if respApi.Response == nil {
		return errors.New("微信企业付款到零钱请求失败" + err.Error())
	}

	resp := respApi.Response
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.Unmarshal(bodyData, result)
	return err
}

func (w *WechatPay) createOutBatchNo(prefix string) string {
	dateStr := time.Now().Format("20060102150405")
	bNo := prefix + dateStr + util.RandStr(6, true)
	return bNo
}

type H5PayerInfo struct {
	OutTradeNo  string `json:"out_trade_no"`
	Description string `json:"description"`
	Total       int    `json:"total"`
	ClientIp    string `json:"client_ip"`
}

// H5Pay h5支付
func (w *WechatPay) H5Pay(payInfo H5PayerInfo) (string, error) {
	reqInfo := ReqJsapi{
		Appid:       w.Appid,
		Mchid:       w.MchId,
		OutTradeNo:  payInfo.OutTradeNo,
		Description: payInfo.Description,
		NotifyUrl:   w.NotifyUrl,
		Amount: &amountInfo{
			Total:    payInfo.Total,
			Currency: "CNY",
		},
		SceneInfo: &sceneInfo{
			H5Info: &h5Info{
				Type: "Wap",
			},
			PayerClientIp: payInfo.ClientIp,
		},
	}
	client, err := w.GetClient()
	if err != nil {
		return "", err
	}
	reqUrl := "https://api.mch.weixin.qq.com/v3/pay/transactions/h5"
	respApi, err := client.Post(context.TODO(), reqUrl, reqInfo)
	if respApi.Response == nil {
		return "", errors.New("微信H5支付请求失败" + err.Error())
	}
	resp := respApi.Response
	bodyData, err := io.ReadAll(resp.Body)
	logrus.Info(string(bodyData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respInfo := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		H5Url   string `json:"h5_url"`
	}{}
	err = json.Unmarshal(bodyData, &respInfo)
	if err != nil {
		return "", err
	}
	if respInfo.H5Url != "" {
		return respInfo.H5Url, nil
	}
	return "", errors.New(respInfo.Message)
}

type NativePayerInfo struct {
	Description string `json:"description"`
	OutTradeNo  string `json:"out_trade_no"`
	Total       int    `json:"total"`
}

// NativePay Native支付
func (w *WechatPay) NativePay(payInfo NativePayerInfo) (string, error) {
	reqInfo := ReqJsapi{
		Appid:       w.Appid,
		Mchid:       w.MchId,
		OutTradeNo:  payInfo.OutTradeNo,
		NotifyUrl:   w.NotifyUrl,
		Description: payInfo.Description,
		Amount: &amountInfo{
			Total:    payInfo.Total,
			Currency: "CNY",
		},
	}
	client, err := w.GetClient()
	if err != nil {
		return "", err
	}
	reqUrl := "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
	respApi, err := client.Post(context.TODO(), reqUrl, reqInfo)
	if respApi.Response == nil {
		return "", errors.New("微信Native支付请求失败" + err.Error())
	}
	resp := respApi.Response
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respInfo := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		CodeUrl string `json:"code_url"`
	}{}
	err = json.Unmarshal(bodyData, &respInfo)
	if err != nil {
		return "", err
	}
	if respInfo.CodeUrl != "" {
		return respInfo.CodeUrl, nil
	}
	return "", errors.New(respInfo.Message)
}

// ReqScanPayXml v2 扫码支付
type ReqScanPayXml struct {
	XMLName        xml.Name `xml:"xml" json:"-"`
	Appid          string   `xml:"appid" json:"appid"`
	MchId          string   `xml:"mch_id" json:"mch_id"`
	NonceStr       string   `xml:"nonce_str" json:"nonce_str"`
	Sign           string   `xml:"sign" json:"sign"`
	Body           string   `xml:"body" json:"body"`
	OutTradeNo     string   `xml:"out_trade_no" json:"out_trade_no"`
	TotalFee       int      `xml:"total_fee" json:"total_fee"`
	SpbillCreateIp string   `xml:"spbill_create_ip" json:"spbill_create_ip"`
	AuthCode       string   `xml:"auth_code" json:"auth_code"`
}

func (req *ReqScanPayXml) BuildSign(mchKey string) {
	dataMap := make(map[string]string)
	dataMap["appid"] = req.Appid
	dataMap["mch_id"] = req.MchId
	dataMap["nonce_str"] = req.NonceStr
	dataMap["body"] = req.Body
	dataMap["out_trade_no"] = req.OutTradeNo
	dataMap["total_fee"] = strconv.Itoa(req.TotalFee)
	dataMap["spbill_create_ip"] = req.SpbillCreateIp
	dataMap["auth_code"] = req.AuthCode

	//key排序
	var keys []string
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//拼接
	var signStr string
	for _, k := range keys {
		signStr += k + "=" + dataMap[k] + "&"
	}
	signStr += "key=" + mchKey
	req.Sign = strings.ToUpper(util.MD5(signStr))
}

type ScanPayerInfo struct {
	Total       int
	Description string
	OutTradeNo  string
	ClientIp    string
	AuthCode    string
}

type RespScancodePayV2 struct {
	XMLName       xml.Name `xml:"xml" json:"-"`
	ReturnCode    string   `xml:"return_code" json:"return_code"`
	ReturnMsg     string   `xml:"return_msg" json:"return_msg"`
	Appid         string   `xml:"appid" json:"appid"`
	MchId         string   `xml:"mch_id" json:"mch_id"`
	DeviceInfo    string   `xml:"device_info" json:"device_info"`
	NonceStr      string   `xml:"nonce_str" json:"nonce_str"`
	Sign          string   `xml:"sign" json:"sign"`
	ResultCode    string   `xml:"result_code" json:"result_code"`
	ErrCode       string   `xml:"err_code" json:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des" json:"err_code_des"`
	Openid        string   `xml:"openid" json:"openid"`
	IsSubscribe   string   `xml:"is_subscribe" json:"is_subscribe"`
	TradeType     string   `xml:"trade_type" json:"trade_type"`
	BankType      string   `xml:"bank_type" json:"bank_type"`
	FeeType       string   `xml:"fee_type" json:"fee_type"`
	TotalFee      int      `xml:"total_fee" json:"total_fee"`
	CashFee       int      `xml:"cash_fee" json:"cash_fee"`
	CashFeeType   string   `xml:"cash_fee_type" json:"cash_fee_type"`
	TransactionId string   `xml:"transaction_id" json:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no" json:"out_trade_no"`
	Attach        string   `xml:"attach" json:"attach"`
	TimeEnd       string   `xml:"time_end" json:"time_end"`
}

func (w *WechatPay) ScancodePayV2(payInfo ScanPayerInfo) (*RespScancodePayV2, error) {
	req := &ReqScanPayXml{
		Appid:          w.Appid,
		MchId:          w.MchId,
		NonceStr:       util.RandStr(24, false),
		Body:           payInfo.Description,
		OutTradeNo:     payInfo.OutTradeNo,
		TotalFee:       payInfo.Total,
		SpbillCreateIp: payInfo.ClientIp,
		AuthCode:       payInfo.AuthCode,
	}
	req.BuildSign(w.MchKey)

	reqBytes, err := xml.Marshal(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip": "支付生成xml失败",
		}).Error(err.Error())
		return nil, err
	}
	reqUrl := "https://api.mch.weixin.qq.com/pay/micropay"
	resp, err := http_client.HttpClient.Post(reqUrl, "application/xml", bytes.NewReader(reqBytes))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip": "扫码请求支付请求失败",
			"url": reqUrl,
		}).Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &RespScancodePayV2{}
	err = xml.Unmarshal(bodyData, result)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tip":  "扫码请求支付返回xml解析失败",
			"body": string(bodyData),
		}).Error(err.Error())
		return nil, err
	}
	if result.ReturnCode != "SUCCESS" {
		return nil, errors.New(result.ReturnMsg)
	}
	if result.ResultCode != "SUCCESS" {
		return nil, errors.New(result.ErrCodeDes)
	}
	return result, nil
}
