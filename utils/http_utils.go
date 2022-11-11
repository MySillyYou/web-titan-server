package utils

import (
	"bytes"
	log "web-server/alog"
	"compress/gzip"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const kDefaultTimeOut = 60

// HTTPUtil for customer Header
type HTTPUtil struct {
	Header   http.Header
	Client   *http.Client
	Resp     *http.Response
	LastPage string
}

// Get http get method
func (p *HTTPUtil) Get(reqURL string, values url.Values) error {
	if values != nil && len(values) > 0 {
		reqURL += "?" + values.Encode()
	}

	var err error
	p.Resp, err = p.DoHTTPRequest(reqURL, "GET", "")
	if err != nil {
		log.Error(err)
	}

	return err
}

// GetLastPageURL get last page url
func (p *HTTPUtil) GetLastPageURL() string {
	if p.Resp != nil {
		return p.Resp.Request.URL.String()
	}
	return ""
}

// PostForm post form use a lot in common case
func (p *HTTPUtil) PostForm(reqURL string, form url.Values) error {

	p.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return p.Post(reqURL, form.Encode())
}

// PostJSON post json use a lot in common case
func (p *HTTPUtil) PostJSON(reqURL string, json string) error {

	p.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return p.Post(reqURL, json)
}

// Post http post method
func (p *HTTPUtil) Post(reqURL, body string) error {
	var err error
	p.Resp, err = p.DoHTTPRequest(reqURL, "POST", body)
	if err != nil {
		log.Error(err)
	}

	return err
}

// DoHTTPRequest 发出http请求
func (p *HTTPUtil) DoHTTPRequest(reqURL, method, body string) (*http.Response, error) {
	pURL, err := url.Parse(reqURL)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req := &http.Request{
		URL:           pURL,
		Method:        strings.ToUpper(method),
		Body:          ioutil.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
		Header:        p.Header,
	}

	if p.Client.Jar == nil {
		p.Client.Jar, _ = cookiejar.New(nil)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resp, err
}

// SetHTTPS 设置https
func (p *HTTPUtil) SetHTTPS() {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*kDefaultTimeOut)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * kDefaultTimeOut))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * kDefaultTimeOut,
	}

	p.Client.Transport = tr
}

// Init 初始化HTTPUtil
func (p *HTTPUtil) Init() {
	p.Header = make(http.Header)
	p.Client = new(http.Client)
	p.Resp = nil
	p.LastPage = ""
}

// ReadBody 读取response body

// ReadBodyByte get body byte use a lot in common case

// ReadBodyString get body string use a lot in common case


func getProperRespBodyReader(resp *http.Response) (io.Reader, error) {
	// Content-Encoding
	ContentEncoding := resp.Header.Get("Content-Encoding")
	ContentEncoding = strings.ToLower(ContentEncoding)

	switch ContentEncoding {
	case "gzip":
		data, err := decodingGZIP(resp)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(string(data)), nil
	default:
		return resp.Body, nil
	}
}

func decodingGZIP(resp *http.Response) ([]byte, error) {
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer gzipReader.Close()

	dataDec, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		log.Error(dataDec)
		return nil, err
	}
	return dataDec, nil
}

func getPBCCRCCharset(header http.Header) string {
	ContentType := header["Content-Type"]
	for _, item := range ContentType {
		strTmp := strings.ToLower(item)

		switch {
		case strings.Contains(strTmp, `charset=utf-8`):
			return `utf-8`
		case strings.Contains(strTmp, `charset=gbk`):
			return `GBK`
		}
	}
	return ""
}

// GetParam ...
func (p *HTTPUtil) GetParam(param map[string]string) string {
	params := p.MergeParam(p.EscapeParam(param))

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

//MergeParam 合并参数
func (p *HTTPUtil) MergeParam(a ...map[string]string) map[string]string {
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

//EscapeParam url encode
func (p *HTTPUtil) EscapeParam(params map[string]string) map[string]string {
	for k, v := range params {
		params[k] = url.QueryEscape(v)
	}
	return params
}
