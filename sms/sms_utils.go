package sms

import (
	log "web-server/alog"
	"crypto/md5"
	sha2562 "crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	g_user_id         = "200521"
	g_password        = "176fb98a71333b1c9a3e0c2ad7763839"
	g_send_sms_url    = "http://112.74.139.4:8002/sms3_api/jsonapi/jsonrpc2.jsp"
	k_allow_phone_num = 100
	k_allow_ip_num    = 1000
)

// 希奥科技
const (
	KSIOOUrl  = "http://sms.10690221.com:9011/hy/"
	KSIOOUid  = "5069915"
	KSIOOAuth = "ddc1c6dd46de2b2aac8321c65337c96f" // MD5(企业代码+用户密码),32位加密小写
	//KSIOOUidRelease  = "90244"
	//KSIOOAuthRelease = "50e2c92e9291809ca26eb81edda70557"
	KSIOOUidRelease  = "902441"
	KSIOOAuthRelease = "0545fb88675228ddd04ba71e3b7c7873"
)

// 凌沃科技
const (
	KLingWoUrl          = "http://116.62.225.148:9001/sms.aspx?%20action=send"
	KLingWoUid          = "3646"
	KLingWoAccount      = "lwhxy"
	KLingWoPassword     = "intelysms666"
	KLingWoCheckContent = "0" // 是否检查内容是否合法，1: 检查  0:不检查
)

//请求状态
type submitPackag struct {
	Content string `json:"content"`
	Phone   string `json:"phone"`
}
type patamPackag struct {
	Userid   string         `json:"userid"`
	Passward string         `json:"password"`
	Submit   []submitPackag `json:"submit"`
}

type RequestPackag struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params patamPackag `json:"params"`
}

//返回结果状态
type returnType struct {
	TypeVar string `json:"return"`
}

type ResponePackag struct {
	Result []returnType `json"result"`
}

func SendSMSMessage(sms, phone string) error {
	//把结构体数据通过json解析成
	var request RequestPackag
	request.ID = 1
	request.Method = "send"
	request.Params.Userid = g_user_id
	request.Params.Passward = g_password

	request.Params.Submit = append(request.Params.Submit, submitPackag{sms, phone})

	data, err := json.Marshal(request)
	if err != nil {
		log.Error("SendSMSMessage json marshal error, content:", request)
		return err
	}

	//http发送验证码
	log.Debug("SendSMSMessage send a message now, content:", string(data))
	resq, err := GetHttpUtil().Post(g_send_sms_url, string(data))
	if err != nil {
		log.Error("SendSMSMessage request error:", err)
		return err
	}

	//解析发送短信之后的状态
	var responePackag ResponePackag
	responePackag.Result = make([]returnType, 1)
	err = json.Unmarshal(resq, &responePackag)
	if err != nil {
		log.Errorf("SendSMSMessage sms[%s] phone[%s] response[%s] error[%v]", sms, phone, string(resq), err)
		return err
	}

	//判断返回状态
	if responePackag.Result[0].TypeVar != "0" {
		log.Errorf("SendSMSMessage sms[%s] phone[%s] response[%s]", sms, phone, string(resq))
		return errors.New("unknow error")
	}
	return nil
}

//通过阿里大于发送短信验证码到指定手机号码
func UsingAlidayuSendSMS(appKey, phone, smsName, smsTemplateCode, code, secret string) error {
	smsParams := fmt.Sprintf("{customer:'%s'}", code)
	return SendSMSAlidayu(appKey, phone, smsName, smsTemplateCode, secret, smsParams)
}

// 阿里大于发送短信，可自定义参数
func SendSMSAlidayu(appKey, phone, smsName, smsTemplateCode, secret, smsParams string) error {
	url := `http://gw.api.taobao.com/router/rest`
	httpUtil := HttpUtils{}
	httpUtil.Init()
	httpUtil.SetHttps()

	param := make(map[string]string)
	param["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	param["v"] = "2.0"
	param["app_key"] = appKey
	param["method"] = "alibaba.aliqin.fc.sms.num.send"
	param["partner_id"] = "top-apitools"
	param["sign_method"] = "md5"
	param["format"] = "json"

	param["sms_type"] = "normal"
	param["rec_num"] = phone
	param["sms_free_sign_name"] = smsName
	param["sms_template_code"] = smsTemplateCode
	param["force_sensitive_param_fuzzy"] = "true"
	param["sms_param"] = smsParams

	param["sign"] = GetSignature(param, secret)

	resp, err := httpUtil.Post1(url, httpUtil.GetParam(param))
	if err != nil {
		log.Errorf("in UsingAlidayuSendSMS::Post error[%s]", err)
		return err
	}

	//出现"err_code":"0"判断为成功
	if strings.Contains(string(resp), `"err_code":"0"`) == false {
		log.Errorf("in UsingAlidayuSendSMS:: error[%s]", string(resp))
		return errors.New(string(resp))
	}
	return nil
}

//计算签名数据
func GetSignature(param map[string]string, secret string) string {
	//对key进行排序
	tmp := make([]string, 0)
	for key, _ := range param {
		tmp = append(tmp, key)
	}
	sort.Strings(tmp)

	signData := ""
	//拼凑签名源数据
	for _, value := range tmp {
		signData = signData + value + param[value]
	}
	//前后都增加secret字符串
	signData = secret + signData + secret
	//	log.Debug("in GetSignature::", signData)

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signData))
	return strings.ToUpper(hex.EncodeToString(md5Ctx.Sum(nil)))
}

// 奥希科技短信平台
func SendMessageSIOO(mobile, content string) error {
	reqStr := fmt.Sprintf("uid=%s&auth=%s&mobile=%s&msg=%s&expid=0&encode=utf-8", KSIOOUidRelease, KSIOOAuthRelease, mobile, content)
	resp, err := post(KSIOOUrl, reqStr)
	if err != nil {
		return err
	}
	log.Infof("SendMessageSIOO mobile[%s] content[%s] reqStr[%s] response[%s]", mobile, content, reqStr, resp)
	rege, _ := regexp.Compile(`^(-*?\d*)(,|)\d*$`)
	result := rege.ReplaceAllString(resp, `$1`)
	if result == "0" {
		return nil
	}
	return errors.New(fmt.Sprintf("ERROR: %s", resp))
}

// 凌沃科技短信平台
func SendMessageLingWo(mobile, content string) error {
	mobileList := strings.Split(mobile, ",")
	reqParse := url.Values{}
	reqParse.Add("userid", KLingWoUid)
	reqParse.Add("account", KLingWoAccount)
	reqParse.Add("password", KLingWoPassword)
	reqParse.Add("mobile", mobile)
	reqParse.Add("content", content)
	reqParse.Add("sendTime", "")
	reqParse.Add("action", "send")
	reqParse.Add("checkcontent", KLingWoCheckContent)
	reqParse.Add("countnumber", fmt.Sprintf("%d", len(mobileList)))
	reqParse.Add("mobilenumber", fmt.Sprintf("%d", len(mobileList)))
	reqParse.Add("telephonenumber", "0")
	reqStr := reqParse.Encode()
	resp, err := post(KLingWoUrl, reqStr)
	if err != nil {
		return err
	}
	log.Infof("SendMessageLingWo mobile[%s] content[%s] reqStr[%s] response[%s]", mobile, content, reqStr, resp)
	rege, _ := regexp.Compile(`<returnstatus>\s*(Success|Faild)\s*</returnstatus>`)
	if strings.Contains(rege.FindString(resp), "Success") {
		return nil
	}
	rege, _ = regexp.Compile(`<message>.*?</message>`)
	return errors.New(rege.FindString(resp))
}

//通过腾讯云发送短信验证码到指定手机号码
func SendTencentSMSCode(appID, appKey, smsName, smsTpl, phone, code string) error {
	rand.Seed(time.Now().UnixNano())
	random := fmt.Sprintf("%06d", rand.Intn(1000000))
	url := fmt.Sprintf(`https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%s&random=%s`, appID, random)
	httpUtil := HttpUtils{}
	httpUtil.Init()
	httpUtil.SetHttps()

	nowTime := time.Now().Unix()

	param := make(map[string]interface{})
	param["time"] = nowTime
	param["ext"] = ""
	param["extend"] = ""
	tplID, err := strconv.Atoi(smsTpl)
	if err != nil {
		log.Error(err.Error())
		param["tpl_id"] = smsTpl
	} else {
		param["tpl_id"] = tplID
	}
	param["sign"] = smsName

	telMap := make(map[string]string)
	telMap["mobile"] = phone
	telMap["nationcode"] = "86"
	param["tel"] = telMap

	tplParam := make([]string, 0)
	if code != "" {
		tplParam = append(tplParam, code)
	}
	param["params"] = tplParam

	getSign := func() string {
		unsigned := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s", appKey, random, nowTime, phone)
		sha256 := sha2562.New()
		sha256.Write([]byte(unsigned))
		return fmt.Sprintf("%x", sha256.Sum(nil))
	}

	param["sig"] = getSign()
	log.Debug(param["sign"])
	log.Debug("SendTencentSMSCode param:", param)

	resp, err := httpUtil.PostJson2(url, param)
	if err != nil {
		log.Errorf("in UsingTencentSendSMS::Post error[%s]", err)
		return err
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(resp, &resultMap)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	resCodeIntf, ok := resultMap["result"]
	if !ok {
		log.Errorf("in UsingTencentSendSMS:: error[%s]", string(resp))
		return errors.New("resp back code format error")
	}

	resCode := fmt.Sprint(resCodeIntf)
	msg, ok := resultMap["errmsg"].(string)
	if !ok {
		log.Errorf("in UsingTencentSendSMS:: error[%s]", string(resp))
		return errors.New("resp back msg format error")
	}

	//出现"result":"0"判断为成功
	if resCode != "0" {
		log.Error(msg)
		return errors.New(msg)
	}

	return nil
}

//通过腾讯云发送短信验证码到指定手机号码
func SendTencentSMSCodeMultiArgs(appID, appKey, smsName, smsTpl, phone string, args ...string) error {
	rand.Seed(time.Now().UnixNano())
	random := fmt.Sprintf("%06d", rand.Intn(1000000))
	url := fmt.Sprintf(`https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%s&random=%s`, appID, random)
	httpUtil := HttpUtils{}
	httpUtil.Init()
	httpUtil.SetHttps()

	nowTime := time.Now().Unix()

	param := make(map[string]interface{})
	param["time"] = nowTime
	param["ext"] = ""
	param["extend"] = ""
	tplID, err := strconv.Atoi(smsTpl)
	if err != nil {
		log.Error(err.Error())
		param["tpl_id"] = smsTpl
	} else {
		param["tpl_id"] = tplID
	}
	param["sign"] = smsName

	telMap := make(map[string]string)
	telMap["mobile"] = phone
	telMap["nationcode"] = "86"
	param["tel"] = telMap

	if len(args) > 0 {
		param["params"] = args
	} else {
		x := make([]string, 0, 0)
		param["params"] = x
	}

	getSign := func() string {
		unsigned := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s", appKey, random, nowTime, phone)
		sha256 := sha2562.New()
		sha256.Write([]byte(unsigned))
		return fmt.Sprintf("%x", sha256.Sum(nil))
	}

	param["sig"] = getSign()
	log.Debug(param["sign"])

	resp, err := httpUtil.PostJson2(url, param)
	if err != nil {
		log.Errorf("in UsingTencentSendSMS::Post error[%s]", err)
		return err
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(resp, &resultMap)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	resCodeIntf, ok := resultMap["result"]
	if !ok {
		log.Errorf("in UsingTencentSendSMS:: error[%s]", string(resp))
		return errors.New("resp back code format error")
	}

	resCode := fmt.Sprint(resCodeIntf)
	msg, ok := resultMap["errmsg"].(string)
	if !ok {
		log.Errorf("in UsingTencentSendSMS:: error[%s]", string(resp))
		return errors.New("resp back msg format error")
	}

	//出现"result":"0"判断为成功
	if resCode != "0" {
		log.Error(msg)
		return errors.New(msg)
	}

	return nil
}
