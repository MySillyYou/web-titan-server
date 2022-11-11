package utils

import (
	log "web-server/alog"
	httpLib "web-server/http"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	KBaiduSDK_AK = "YKFXqfXb0igGhdo3emucyu3HfLgFx3gS" // 开发者账号: xiaowentian
)

type Location struct {
	Province string `json:"province"`
	City     string `json:"city"`
}

type Content struct {
	AddressDetail Location `json:"address_detail"`
}

type IPLocationResp struct {
	Address string  `json:"address"`
	Content Content `json:"content"`
}

// {"address":"CN|云南|丽江|None|CHINANET|0|0","content":{"address_detail":{"province":"云南省","city":"丽江市","district":"","street":"","street_number":"","city_code":114},"address":"云南省丽江市","point":{"y":"3088416.04","x":"11157632.6"}},"status":0}

/*
  Return: Location IP归属地省市 int64 查询耗时，单位 ms
*/
func GetBaiduSdkIPLocation(ip string) (Location, int64) {
	start := time.Now().UnixNano()
	runTime := func(s int64) int64 {
		return (time.Now().UnixNano() - s) / 1e6
	}

	httpUtil := &httpLib.HttpUtils{}
	httpUtil.Init()
	httpUtil.SetTimeout(2)

	loc := Location{Province: "", City: ""}

	reqUrl := fmt.Sprintf("http://api.map.baidu.com/location/ip?ip=%s&ak=%s&coor=", ip, KBaiduSDK_AK)
	respByte, err := httpUtil.Get(reqUrl, "")
	if err != nil {
		log.Errorf("GetBaiduSdkIPLocation ip[%s] request error[%s]", ip, err.Error())
		return loc, runTime(start)
	}

	resp := IPLocationResp{}

	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		log.Errorf("GetBaiduSdkIPLocation ip[%s] Unmarshal error[%s] content[%s]", ip, err.Error(), string(respByte))
		return loc, runTime(start)
	}
	resp.Content.AddressDetail.Province = strings.Replace(resp.Content.AddressDetail.Province, "省", "", -1)
	resp.Content.AddressDetail.Province = strings.Replace(resp.Content.AddressDetail.Province, "市", "", -1)
	resp.Content.AddressDetail.City = strings.Replace(resp.Content.AddressDetail.City, "市", "", -1)
	return resp.Content.AddressDetail, runTime(start)
}
