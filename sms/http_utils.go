package sms

import (
	log "web-server/alog"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type HttpUtils struct {
	Cookies      []*http.Cookie
	Client       *http.Client
	Cookiejar    *cookiejar.Jar
	CommonHeader map[string]string
}

const TIMEOUT = 10

var HttpsUtil *HttpUtils

func init() {
	if HttpsUtil == nil {
		HttpsUtil = &HttpUtils{}
		HttpsUtil.Init()
	}
}

func (this *HttpUtils) Init() {
	this.Client = &http.Client{}
	this.Cookiejar, _ = cookiejar.New(nil)
	this.CommonHeader = make(map[string]string, 0)
}

func (this *HttpUtils) SetHttps() {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*TIMEOUT)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * TIMEOUT))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * TIMEOUT,
	}

	this.Client.Transport = tr
}

func (this *HttpUtils) SetProxy(proxyAddr string) error {
	proxy, err := url.Parse(proxyAddr)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*TIMEOUT)
			if err != nil {
				log.Error(err.Error())
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * TIMEOUT))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * TIMEOUT,
		Proxy: http.ProxyURL(proxy),
	}

	this.Client.Transport = tr
	return nil
}

func (this *HttpUtils) SetTimeOutDefault() {
	this.Client.Transport = nil
}

func (this *HttpUtils) Get(url string, param string) ([]byte, error) {
	urlStr := url
	if "" != param {
		urlStr += "?" + param
	}
//	log.Debug("GET", urlStr)

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)
	resp, err := this.Client.Get(urlStr)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
//		log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Get(location, "")
		if err != nil {
			return nil, err
		}
	}

	return data, err

}

func (this *HttpUtils) Get1(url string, param string) ([]byte, error) {

	urlStr := url
	if "" != param {
		urlStr += "?" + param
	}

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)

	req, err := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
//	log.Debug("GET", urlStr)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
//		log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Get1(location, "")
		if err != nil {
			return nil, err
		}
	}

	return data, err

}

func (this *HttpUtils) Post(urlStr string, param map[string]string) ([]byte, error) {
//	log.Debug("POST", urlStr)

	paramData := url.Values{}

	if param != nil {
		for key, value := range param {
			paramData[key] = []string{value}
		}
	}

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)
	resp, err := this.Client.PostForm(urlStr, paramData)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
//		log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Post(location, nil)
		if err != nil {
			return nil, err
		}
	}

	return data, err
}

func (this *HttpUtils) Post1(url string, param string) ([]byte, error) {
//	log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)

	req, err := http.NewRequest("POST", url, strings.NewReader(param))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
//		log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Post1(location, "")
		if err != nil {
			return nil, err
		}
	}

	return data, err
}

func (this *HttpUtils) Post2(url string, param string) (http.Header, []byte, error) {
//	log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
	//	log.Debug("this Cookiejar1", this.Cookiejar)

	req, err := http.NewRequest("POST", url, strings.NewReader(param))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
	//	log.Debug(this.Cookies)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, nil, err
	}

	return resp.Header, data, err
}

func (this *HttpUtils) PostJson(url string, param map[string]string) ([]byte, error) {
//	log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)

	bytesData, err := json.Marshal(param)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bytesData))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	//	if resp.StatusCode == 301 || resp.StatusCode == 302 {
	//		log.Debugf("%d redirect ", resp.StatusCode)
	//		location := resp.Header["Location"][0]
	//		data, err = this.Post1(location, "")
	//		if err != nil {
	//			return nil, err
	//		}
	//	}

	return data, err
}

func (this *HttpUtils) PostJson2(url string, param interface{}) ([]byte, error) {
//	log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
//	log.Debug("this Cookiejar1", this.Cookiejar)

	bytesData, err := json.Marshal(param)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bytesData))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
//	log.Debug("Cookies", this.Cookies)
//	log.Debug("this Cookiejar2", this.Cookiejar)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" { // unzip gzip data
		b := bytes.NewReader(data)
		r, err := gzip.NewReader(b)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	return data, err
}

func (this *HttpUtils) GetHttps(url string, param string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*TIMEOUT)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * TIMEOUT))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * TIMEOUT,
	}

	this.Client.Transport = tr
	return this.Get(url, param)
}

func (this *HttpUtils) PostHttps(urlStr string, param map[string]string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*TIMEOUT)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * TIMEOUT))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * TIMEOUT,
	}

	this.Client.Transport = tr
	return this.Post(urlStr, param)
}

func (this *HttpUtils) GetParam(param map[string]string) string {
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
func (this *HttpUtils) MergeParam(a ...map[string]string) map[string]string {
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
func (this *HttpUtils) EscapeParam(params map[string]string) map[string]string {
	for k, v := range params {
		params[k] = url.QueryEscape(v)
	}
	return params
}

func (this *HttpUtils) GetCookieStr() string {
	ret_str := ""
	for _, cookie := range this.Cookies {
		ret_str += cookie.Name + "=" + cookie.Value + ";"
	}

//	log.Debug("GetCookieStr Cookies", this.Cookies)
//	log.Debug("GetCookieStr ret_str", ret_str)
	return ret_str
}

func post(url, content string) (string, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	client := http.Client{ // 设置5秒超时
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(5 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
