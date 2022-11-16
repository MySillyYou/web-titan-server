package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	log "web-server/alog"
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

// SetTimeout 设置超时时间(单位: 秒)
func (this *HttpUtils) SetTimeout(timeout int64) {
	this.Client.Timeout = time.Duration(timeout) * time.Second
}

func (this *HttpUtils) SetHttpsWithTimeout(seconds int64) {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*time.Duration(seconds))
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * time.Duration(seconds)))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * time.Duration(seconds),
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
		Proxy:                 http.ProxyURL(proxy),
	}

	this.Client.Transport = tr
	return nil
}

func (this *HttpUtils) SetProxyNos(proxyAddr string) error {
	proxy, err := url.Parse(proxyAddr)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	tr := &http.Transport{
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
		Proxy:                 http.ProxyURL(proxy),
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
		if strings.Contains(urlStr, "?") {
			urlStr = strings.TrimSuffix(urlStr, "&")
			urlStr += "&" + param
		} else {
			urlStr += "?" + param
		}
	}
	//log.Debug("GET", urlStr)

	this.Client.Jar = this.Cookiejar
	//log.Debug("this Cookiejar1", this.Cookiejar)
	log.Debug(urlStr)
	resp, err := this.Client.Get(urlStr)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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
		//	log.Debugf("%d redirect ", resp.StatusCode)
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
	//log.Debug("this Cookiejar1", this.Cookiejar)

	req, err := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for key, value := range this.CommonHeader {
		req.Header.Set(key, value)
	}

	resp, err := this.Client.Do(req)
	//log.Debug("GET", urlStr)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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
		//	log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Get1(location, "")
		if err != nil {
			return nil, err
		}
	}

	return data, err

}

func (this *HttpUtils) Post(urlStr string, param map[string]string) ([]byte, error) {
	//log.Debug("POST", urlStr)

	paramData := url.Values{}

	if param != nil {
		for key, value := range param {
			paramData[key] = []string{value}
		}
	}

	this.Client.Jar = this.Cookiejar
	//log.Debug("this Cookiejar1", this.Cookiejar)
	resp, err := this.Client.PostForm(urlStr, paramData)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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
		//	log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Post(location, nil)
		if err != nil {
			return nil, err
		}
	}

	return data, err
}

func (this *HttpUtils) Post1(url string, param string) ([]byte, error) {
	//log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
	//log.Debug("this Cookiejar1", this.Cookiejar)

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
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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
		log.Debugf("%d redirect ", resp.StatusCode)
		location := resp.Header["Location"][0]
		data, err = this.Post1(location, "")
		if err != nil {
			return nil, err
		}
	}

	return data, err
}

func (this *HttpUtils) Post2(url string, param string) (http.Header, []byte, error) {
	//log.Debug("POST", url)

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

func (this *HttpUtils) PostJsonWithDump(url string, param map[string]string) ([]byte, error) {
	//log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
	//log.Debug("this Cookiejar1", this.Cookiejar)

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

	dump, _ := httputil.DumpRequestOut(req, true)
	log.Debugf("request:%s", string(dump))

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	dump, _ = httputil.DumpResponse(resp, true)
	log.Debugf("resp:%s", string(dump))

	this.Cookies = resp.Cookies()
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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

func (this *HttpUtils) PostJson(url string, param map[string]string) ([]byte, error) {
	//log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
	//log.Debug("this Cookiejar1", this.Cookiejar)

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
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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

func (this *HttpUtils) PostJsonIf(url string, param interface{}) ([]byte, error) {
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

func (this *HttpUtils) PostJsonIfWithDump(url string, param interface{}) ([]byte, error) {
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

	dump, _ := httputil.DumpRequestOut(req, true)
	log.Debugf("request:%s", string(dump))

	resp, err := this.Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	this.Cookies = resp.Cookies()
	//	log.Debug("Cookies", this.Cookies)
	//	log.Debug("this Cookiejar2", this.Cookiejar)

	dump, _ = httputil.DumpResponse(resp, true)
	log.Debugf("resp:%s", string(dump))

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

func (this *HttpUtils) PostFormData(url string, params map[string]string) ([]byte, error) {
	this.Client.Jar = this.Cookiejar

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
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
	//log.Debug("Cookies", this.Cookies)
	//log.Debug("this Cookiejar2", this.Cookiejar)

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

func (this *HttpUtils) GetJson(url string, param interface{}) ([]byte, error) {
	log.Debug("GET", url)

	this.Client.Jar = this.Cookiejar
	log.Debug("this Cookiejar1", this.Cookiejar)

	bytesData, err := json.Marshal(param)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewReader(bytesData))
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
	log.Debug("Cookies", this.Cookies)
	log.Debug("this Cookiejar2", this.Cookiejar)

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

func (this *HttpUtils) PostXml(url string, param string) ([]byte, error) {
	log.Debug("POST", url)

	this.Client.Jar = this.Cookiejar
	log.Debug("this Cookiejar1", this.Cookiejar)

	req, err := http.NewRequest("POST", url, strings.NewReader(param))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")
	//	req.Header.Set("Accept", "application/xml;charset=UTF-8")
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
	log.Debug("Cookies", this.Cookies)
	log.Debug("this Cookiejar2", this.Cookiejar)

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

// MergeParam 合并参数
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

// EscapeParam url encode
func (this *HttpUtils) EscapeParam(params map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range params {
		newMap[k] = url.QueryEscape(v)
	}
	return newMap
}

func (this *HttpUtils) GetCookieStr() string {
	ret_str := ""
	for _, cookie := range this.Cookies {
		//		ret_str += cookie.String() + ";"
		ret_str += cookie.Name + "=" + cookie.Value + ";"
	}

	log.Debug("GetCookieStr Cookies", this.Cookies)
	log.Debug("GetCookieStr ret_str", ret_str)
	return ret_str
}

func GetRequestIP(req *http.Request) string {
	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	forward := req.Header.Get("X-Forwarded-For")
	ips := strings.Split(forward, ",")
	//log.Debugf("Forwarded for: %v", ips)
	if len(ips) != 0 {
		ip := strings.TrimSpace(ips[0])
		if isValid := net.ParseIP(ip); isValid != nil {
			return ip
		}
	}

	if realIP := GetRealIP(req); realIP != "" {
		return realIP
	}

	// get RemoteAddr
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	//log.Debugf("RemoteAddr IP: %s", ip)
	//log.Debugf("RemoteAddr Port: %s", port)
	if err != nil {
		log.Debugf("userip: %q is not IP:port", req.RemoteAddr)
		return ""
	} else {
		if isValid := net.ParseIP(ip); isValid != nil {
			return ip
		}
	}

	return ""
}

func GetRealIP(req *http.Request) string {
	return req.Header.Get("X-Real-IP")
}
