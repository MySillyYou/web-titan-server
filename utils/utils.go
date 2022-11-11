package utils

import (
	log "web-server/alog"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/context"
	"net/http"
	"runtime"
	"strings"
	"regexp"
)

var (
	errCusID = errors.New("invaild customer id")
)

// HandleSuccessFile ...
func HandleSuccessFile(
	w http.ResponseWriter,
	r *http.Request,
	data []byte,
	fileName string,
	path string,
) {
	w.Header().Add("Content-Type", "application/octet-stream;charset=utf-8")
	if path != "" {
		w.Header().Add("File-URL", path)
	}
	if fileName != "" {
		w.Header().Add("Content-disposition", "attachment;filename="+fileName)
	}

	w.Write(data)
	logRequestSuccess(r, 1)
}

// HandleSuccess ...
func HandleSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	// set Response Header
	customReturnHeader(w)

	// ret map
	retMap := map[string]interface{}{
		"code": 0,
		"msg":  "",
		"data": data,
	}

	retByte, err := json.Marshal(retMap)
	if err != nil {
		log.Error(err)
	}
	w.Write(retByte)
	logRequestSuccess(r, 1)
}

//handleSuccess9188 9188成功返回的数据
func HandleSuccess9188(w http.ResponseWriter, dataString string) {
	//	_, file, line, _ := runtime.Caller(1)
	dataStr := ""
	if "" == dataString {
		dataStr = `""`
	} else {
		dataStr = fmt.Sprintf(`%s`, dataString)
	}
	w.Write([]byte(fmt.Sprintf(`
		{
			"code":0,
			"msg":"",
			"data":%s
		}
		`, dataStr)))
}

func HandleCustomerError(w http.ResponseWriter, r *http.Request, customerError error) {
	// set Response Header
	customReturnHeader(w)

	retByte, err := json.Marshal(customerError)
	if err != nil {
		log.Error(err)
	}
	w.Write(retByte)
	logRequestFail(r, 1)
}

// HandleError ...
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	// set Response Header
	customReturnHeader(w)

	// ret map
	retMap := map[string]interface{}{
		"code": -1,
		"msg":  err.Error(),
	}

	retByte, err := json.Marshal(retMap)
	if err != nil {
		log.Error(err)
	}
	w.Write(retByte)
	logRequestFail(r, 1)
}

const (
	kPreChannel = "_channel_"
	kPreID      = "_id_"
	kPreIDType  = "_idType_"
	kPreAppend  = "_append_"
)

// GetCustomerID ...
func GetCustomerID(channel, id, idType, append string) string {
	return fmt.Sprintf("%s%s%s%s%s%s%s%s", kPreChannel, channel, kPreID, id, kPreIDType, idType, kPreAppend, append)
}

// ParseCustomerID ...
func ParseCustomerID(customerID string) (string, string, string, string, error) {
	//channel_postMan_id_afjiefajo3259584905_idType_1
	channel := ""
	id := ""
	idType := ""
	append := ""
	err := func() error {
		find := func(pre string) (string, error) {
			nBeg := strings.Index(customerID, pre)
			if nBeg >= 0 {
				strTmp := customerID[nBeg:]
				customerID = customerID[:nBeg]
				return strTmp[len(pre):], nil
			}
			return "", errCusID
		}

		// take care the order
		var err error
		append, err = find(kPreAppend)
		if err != nil {
			return err
		}

		idType, err = find(kPreIDType)
		if err != nil {
			return err
		}

		id, err = find(kPreID)
		if err != nil {
			return err
		}

		channel, err = find(kPreChannel)
		if err != nil {
			return err
		}

		return nil
	}()

	return channel, id, idType, append, err
}

func customReturnHeader(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, No-Cache, X-Requested-With, If-Modified-Since, Pragma, Last-Modified, Cache-Control, Expires, Content-Type, X-E4M-With, channel, version")
}

func logRequestSuccess(r *http.Request, dep int) {
	fnName := getFuncName(dep + 1)
	log.Info(GetChannelIDLogMsg(r), "request success on func:", fnName)
}

func logRequestFail(r *http.Request, dep int) {
	fnName := getFuncName(dep + 1)
	log.Info(GetChannelIDLogMsg(r), "request fail on func:", fnName)
}

func getFuncName(dep int) string {
	pc, _, _, _ := runtime.Caller(dep + 1)
	fn := runtime.FuncForPC(pc)
	return fn.Name()
}

func GetChannelIDLogMsg(r *http.Request) string {

	cntLogPrefix := context.Get(r, "logPrefix")
	logPrefix := fmt.Sprint(cntLogPrefix)
	switch {
	case cntLogPrefix == nil, logPrefix == "":
		id := r.FormValue("id")
		if id == "" {
			id = fmt.Sprint(context.Get(r, "id"))
		}
		id_type := r.FormValue("id_type")
		if id_type == "" {
			id_type = fmt.Sprint(context.Get(r, "id_type"))
		}
		channel := r.Header.Get("channel")
		if channel == "" {
			channel = fmt.Sprint(context.Get(r, "channel"))
		}

		funName := context.Get(r, "funName")
		strLog := fmt.Sprintf("%s channel:%s,id:%s,id_type:%s", funName, channel, id, id_type)

		switch id_type {
		case "3":
			accu_city := r.FormValue("accu_city")
			if accu_city == "" {
				accu_city = fmt.Sprint(context.Get(r, "accu_city"))
			}
			strLog += ",accu_city:" + accu_city
		case "7":
			app_name := r.FormValue("app_name")
			if app_name == "" {
				app_name = fmt.Sprint(context.Get(r, "app_name"))
			}
			strLog += ",app_name:" + app_name
		case "9":
			wx_appid := r.FormValue("wx_appid")
			if wx_appid == "" {
				wx_appid = fmt.Sprint(context.Get(r, "wx_appid"))
			}
			strLog += ",wx_appid:" + wx_appid
		}

		logPrefix = strLog + ";"
		context.Set(r, "logPrefix", strLog)
	}

	return logPrefix
}

/*
* 对sql语句转义，防止SQL注入攻击
 */
func EscapeStringBackslash(v string) string {
	pos := 0
	buf := make([]byte, len(v)*2)

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos += 1
		}
	}

	return string(buf[:pos])
}

/*
 * 生成sql语句，已包含字符串转义，防止SQL注入攻击
 * 使用方法同 fmt.Sprintf(format, a...)
 */
func GetEscapeSqlClause(format string, a ...interface{}) string {
	if len(a) <= 0 {
		return fmt.Sprintf(format, a...)
	}

	args := make([]interface{}, 0, len(a))
	for _, arg := range a {
		switch arg.(type) {
		case string:
			newArg := EscapeStringBackslash(arg.(string))
			args = append(args, newArg)
		default:
			args = append(args, arg)
		}
	}

	return fmt.Sprintf(format, args...)
}


/*
 * 纠正地址  level为精度
 * level =3 精确到路
 * level =2 精确到区县
 * level =1 精确到省市
 */
func CorrectAddress(address string , level int) string {
	var caddress string
	var index int = -1
	var kword int
	Kadd3 :=[]string{"路","街","巷","旗","胡同","道"}
	Kadd2 :=[]string{"市","区","县","自治州","自治县"}
	Kadd1 :=[]string{"省","北京市","天津市","上海市","重庆市","自治区"}
	//去非法字段
	var illegalChar = regexp.MustCompile("[a-zA-Z0-9/-]")
	loc := illegalChar.FindIndex([]byte(address))
	if loc != nil {
		caddress = string([]byte(address)[0:loc[0]])
	}else {
		caddress = address
	}
	//3
	if level == 3 {
		for key, value := range Kadd3 {
			temp := strings.LastIndex(caddress,value)
			if temp > index{
				index = temp
				kword = key
			}
		}
		if index != -1 {
			caddress = caddress[0:index]+ Kadd3[kword]
			return caddress
		}else {
			level = 2
		}
	}
	//2
	if level == 2 {
		index = -1
		kword = 0
		for key, value := range Kadd2 {
			temp := strings.LastIndex(caddress,value)
			if temp > index{
				index = temp
				kword = key
			}
		}
		for _, value := range Kadd1 {
			if strings.Index(caddress,value) != strings.LastIndex(caddress,value) {
				index = -1
			}
		}
		if index != -1 {
			caddress = caddress[0:index]+ Kadd2[kword]
			return caddress
		}else {
			level = 1
		}
	}

	//1
	if level == 1 {
		for key, value := range Kadd1 {
			temp := strings.Index(caddress,value)
			if temp != -1 {
				caddress = caddress[0:temp]+ Kadd1[key]
				return caddress
			}
		}
	}

	return caddress
}
