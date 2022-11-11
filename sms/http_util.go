package sms

import (
	log "web-server/alog"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type HttpUtil struct {
}

var util *HttpUtil

func GetHttpUtil() *HttpUtil {
	if util == nil {
		util = &HttpUtil{}
	}

	return util
}

func (this *HttpUtil) Get(url string, param string) ([]byte, error) {
	client := &http.Client{}
	urlStr := url
	if "" != param {
		urlStr += "?" + param
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	//	req.Header.Add("apikey", "984c96922664fec046e7d9b8ea4c6834")
	resp, err := client.Do(req)
	log.Debug(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, err

}

func (this *HttpUtil) Post(url string, param string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(param))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (this *HttpUtil) GetParam(param map[string]string) string {
	params := this.MergeParam(this.EscapeParam(param))

	var buf bytes.Buffer
	var sep = ""
	for key, value := range params {
		buf.WriteString(sep)
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(value)
		sep = "&"
	}

	return buf.String()
}

//合并参数
func (this *HttpUtil) MergeParam(a ...map[string]string) map[string]string {
	b := make(map[string]string)
	for _, m := range a {
		if m == nil {
			continue
		}
		for k, v := range m {
			b[k] = v
		}
	}
	return b
}

//url encode
func (this *HttpUtil) EscapeParam(params map[string]string) map[string]string {
	for k, v := range params {
		params[k] = url.QueryEscape(v)
	}
	return params
}
