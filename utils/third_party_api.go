package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	log "web-server/alog"
	httpLib "web-server/http"
	redisLib "web-server/redis"
)

// 聚合数据二要素(手机号码绑定身份证号和姓名校验)验证AppKey, 文档地址: https://www.juhe.cn/docs/api/id/251
const (
	KJuheAppKey = "82a1d7f56767e0ce0a10b1ab9b92cd6c"
	KJuheURL    = "http://v.juhe.cn/telecom2/query"

	KCheckFailed = 0 // 校验失败
	KCheckTrue   = 1 // 二要素身份验证一致
	KCheckFalse  = 2 // 二要素身份验证不一致
)

// 聚合手机二元素校验
func JuheMobileUsernameCheck(username, mobile string) int {
	if username == "" || len(mobile) != 11 {
		log.Errorf("JuheMobileUsernameCheck username[%s] mobile[%s] format error", username, mobile)
		return KCheckFailed
	}

	httpUtil := httpLib.HttpUtils{}
	httpUtil.Init()
	httpUtil.SetTimeout(2)

	params := make(map[string]string, 0)
	params["key"] = KJuheAppKey
	params["realname"] = username
	params["mobile"] = mobile
	respByte, err := httpUtil.Get(KJuheURL, httpUtil.GetParam(params))
	log.Infof("JuheMobileUsernameCheck mobile[%s] username[%s] resp[%s] error[%v]", mobile, username, string(respByte), err)
	if err != nil {
		log.Errorf("JuheMobileUsernameCheck mobile[%s] username[%s] error[%v]", mobile, username, err)
		return KCheckFailed
	}

	// {"reason":"成功","result":{"realname":"夏笑声","mobile":"17620163567","res":1,"resmsg":"二要素身份验证一致"},"error_code":0}
	type Result struct {
		RealName string `json:"realname"`
		Mobile   string `json:"mobile"`
		ResCode  int    `json:"res"`
		ResMsg   string `json:"resmsg"`
	}
	type Resp struct {
		Reason    string `json:"reason"`
		Result    Result `json:"result"`
		ErrorCode int    `json:"error_code"`
	}
	resp := Resp{}

	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		log.Errorf("JuheMobileUsernameCheck mobile[%s] username[%s] resp[%s] error[%v]", mobile, username, string(respByte), err)
		return KCheckFailed
	}
	return resp.Result.ResCode
}

const (
	KWanshuURL    = "https://api.253.com/open/carriers/carriers-auth"
	KWanshuAppID  = "bLxTnZu9"
	KWanshuAppKey = "wx8NajUu"

	KWanshuMainKey  = "WS_THREE" // 万数三要素redis缓存key
	KWanshuCacheDay = 30         // 万数三要素缓存天数
)

// INFO : 2018/12/12 15:20:50 WanshuThreeKeyElements params[map[idNum:360681199401136130 mobile:17620163567 appId:bLxTnZu9 appKey:wx8NajUu name:夏笑声]] resp[{"chargeStatus":1,"message":"成功","data":{"orderNo":"6668677421942108","handleTime":"2018-12-12 15:20:50","type":"2","result":"01","gender":"1","age":"25","remark":"一致"},"code":"200000"}] error[<nil>]
// 万数三要素接口
func WanshuThreeKeyElements(username, mobile, idcard string, redis *redisLib.Util) bool {
	params := map[string]string{"appId": KWanshuAppID, "appKey": KWanshuAppKey, "name": username, "idNum": idcard, "mobile": mobile}

	if redis != nil {
		// redis value格式: 姓名_身份证号_验证通过时间戳
		personInfo, err := redis.GetMapKeyValue(KWanshuMainKey, mobile)
		if err == nil {
			if tmp := strings.Split(personInfo, "_"); len(tmp) >= 3 && tmp[0] == username && tmp[1] == idcard &&
				(time.Now().Unix()-StrToInt64(tmp[2])) < 24*3600*KWanshuCacheDay { // 30天有效期
				log.Infof("WanshuThreeKeyElements noquery params[%v] result is ok", params)
				return true
			}
		}

	}

	httpUtil := httpLib.HttpUtils{}
	httpUtil.Init()
	httpUtil.SetHttps()
	httpUtil.SetTimeout(3)

	respByte, err := httpUtil.Post(KWanshuURL, params)
	log.Infof("WanshuThreeKeyElements newquery params[%v] resp[%s] error[%v]", params, string(respByte), err)
	if err != nil {
		log.Errorf("WanshuThreeKeyElements error[%v] param[%v]", err, params)
		return false
	}

	type Resp struct {
		Code         string            `json:"code"`
		ChargeStatus int               `json:"chargeStatus"`
		Message      string            `json:"message"`
		Data         map[string]string `json:"data"`
	}
	resp := Resp{}
	if err = json.Unmarshal(respByte, &resp); err != nil {
		log.Errorf("WanshuThreeKeyElements error[%v] param[%v]", err, params)
		return false
	}

	if resp.Code == "200000" && resp.Data != nil && resp.Data["result"] == "01" {
		redis.SetMapKeyValue(KWanshuMainKey, mobile, fmt.Sprintf("%s_%s_%d", username, idcard, time.Now().Unix()))
		return true
	}
	return false
}

// 万数三要素接口, 增加可传入通路的配置
// params 可传入通路等参数便于进行日志统计
func WanshuTriElementCheck(username, mobile, idcard string, redis *redisLib.Util, args map[string]string) bool {
	params := map[string]string{"appId": KWanshuAppID, "appKey": KWanshuAppKey, "name": username, "idNum": idcard, "mobile": mobile}

	if redis != nil {
		// redis value格式: 姓名_身份证号_验证通过时间戳
		personInfo, err := redis.GetMapKeyValue(KWanshuMainKey, mobile)
		if err == nil {
			if tmp := strings.Split(personInfo, "_"); len(tmp) >= 3 && tmp[0] == username && tmp[1] == idcard &&
				(time.Now().Unix()-StrToInt64(tmp[2])) < 24*3600*KWanshuCacheDay { // 30天有效期
				log.Infof("WanshuThreeKeyElements noquery params[%v] result is ok", params)
				log.Debugf("loan-mark-weshare|tri-element-cache-pass|source:%s", args["source"])
				return true
			}
		}

	}

	httpUtil := httpLib.HttpUtils{}
	httpUtil.Init()
	httpUtil.SetHttps()
	httpUtil.SetTimeout(3)

	log.Debugf("loan-mark-weshare|tri-element-call|source:%s|loan_id:%s|mobile:%s", args["source"], args["loan_id"], args["mobile"])
	respByte, err := httpUtil.Post(KWanshuURL, params)
	log.Infof("WanshuThreeKeyElements newquery params[%v] resp[%s] error[%v]", params, string(respByte), err)
	if err != nil {
		log.Errorf("WanshuThreeKeyElements error[%v] param[%v]", err, params)
		return false
	}

	type Resp struct {
		Code         string            `json:"code"`
		ChargeStatus int               `json:"chargeStatus"`
		Message      string            `json:"message"`
		Data         map[string]string `json:"data"`
	}
	resp := Resp{}
	if err = json.Unmarshal(respByte, &resp); err != nil {
		log.Errorf("WanshuThreeKeyElements error[%v] param[%v]", err, params)
		return false
	}

	if resp.Code == "200000" && resp.Data != nil && resp.Data["result"] == "01" {
		log.Debugf("loan-mark-weshare|tri-element-postok|source:%s|loan_id:%s|mobile:%s", args["source"], args["loan_id"], args["mobile"])
		redis.SetMapKeyValue(KWanshuMainKey, mobile, fmt.Sprintf("%s_%s_%d", username, idcard, time.Now().Unix()))
		return true
	}
	return false
}
