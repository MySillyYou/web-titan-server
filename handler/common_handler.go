package handler

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"net/http"
	"net/rpc"
	"reflect"
	"strconv"
	"strings"
	"time"
	log "web-server/alog"
)

// HTTP返回错误
const (
	KRoleAdmin   = "1" //管理员
	KRoleInside  = "2" //内部人员
	KRoleOutside = "3" //外部人员
)
const (
	KAddNew = "1" //新建
	KModify = "2" //修改
)

var Conn *rpc.Client
var RpcAddr string

// HTTP返回错误
const (
	KErrorSucc          = 0
	KErrorNoSource      = 401
	KErrorArgs          = 402
	KErrorNoArg         = 403
	KErrorProductRepeat = 404
	KErrorExistSrc      = 405
	KErrorUserNotExist  = 406
	KErrorSectionRepeat = 407
	KErrorServer        = 500
)

var KErrorMsg = map[int]string{
	KErrorSucc:          "",
	KErrorServer:        "服务器异常",
	KErrorNoSource:      "通路不存在",
	KErrorArgs:          "参数错误",
	KErrorNoArg:         "缺少参数",
	KErrorProductRepeat: "小贷名称不能重复",
	KErrorExistSrc:      "通路已存在",
	KErrorUserNotExist:  "用户不存在",
	KErrorSectionRepeat: "扣量区间重合",
}

const (
	// 年月
	TimeFormatYM = "2006-01"

	//年月日
	TimeFormatYMD = "2006-01-02"
	TimeFormatMD  = "01-02"
	TimeFormatHM  = "15:04"
	TimeFormatM   = "04"

	//年月日时分秒
	TimeFormatYMDHMS = "2006-01-02 15:04:05"
)

func HandleError(w http.ResponseWriter, msg string) {
	HandleBack(w, -1, msg, "")
}

func HandleSuccess(w http.ResponseWriter, dataString interface{}) {
	HandleBack(w, 0, "", dataString)
}

func HandleAuthError(w http.ResponseWriter) {
	HandleBack(w, -2, "没有访问权限", nil)
}

func HandleCodeMsg(w http.ResponseWriter, code int, msg string) {
	HandleBack(w, code, msg, nil)
}

func HandleBack(w http.ResponseWriter, code int, msg string, dataString interface{}) {
	ret := make(map[string]interface{})
	ret["code"] = code
	ret["msg"] = msg
	if dataString != nil {
		ret["data"] = dataString
	}

	byteJson, err := json.Marshal(ret)
	if err != nil {
		log.Error(err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	//	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, No-Cache, X-Requested-With, If-Modified-Since, Pragma, Last-Modified, Cache-Control, Expires, Content-Type, X-E4M-With, channel, version")

	w.Write(byteJson)
}

func getAuthChannelsCond(r *http.Request) string {
	if r.FormValue("role") == KRoleAdmin {
		return "true"
	}
	opChannels := r.FormValue("channels")
	if opChannels == "" {
		return "false"
	}

	tempList := strings.Split(opChannels, ",")
	ret := fmt.Sprintf("channel in ('%s')", strings.Join(tempList, "','"))
	return ret
}

func MakeXslxFileWithFieldNamesFromMapList(fieldNames []string, fields []string, contents []map[string]string) (*xlsx.File, error) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("sheet1")
	if err != nil {
		return nil, err
	}

	style := xlsx.NewStyle()
	style.Font.Bold = true

	headRow := sheet.AddRow()
	for _, itemName := range fieldNames {

		cell := headRow.AddCell()
		cell.Value = itemName
		cell.SetStyle(style)
	}

	for _, lineContent := range contents {
		dataRow := sheet.AddRow()
		for _, key := range fields {
			cell := dataRow.AddCell()
			cell.Value = lineContent[key]
		}
	}
	return file, nil
}

func Str2Int(s string) int {
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return ret
}

func Str2Int64(s string) int64 {
	ret, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error(err.Error())
		return int64(0)
	}
	return ret
}

func Str2Float64(s string) float64 {
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error(err.Error())
		return 0.00
	}
	return ret
}

func StructConvertMapByTag(obj interface{}, tagName string) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get(tagName)
		fmt.Println(tagName)
		if tagName != "" && tagName != "-" {
			data[tagName] = v.Field(i).Interface()
		}
	}
	return data
}

func IsInArray(list []string, val string) bool {
	for _, item := range list {
		if item == val {
			return true
		}
	}

	return false
}

/*
 * 比较日期
 * first > second return 1
 * first < second return -1
 * first = second return 0
 */
func compareDate(firstDate, secondDate string) int {
	t1, err := time.Parse("2006-01-02", firstDate)
	if err != nil {
		return -2
	}
	t2, err := time.Parse("2006-01-02", secondDate)
	if err != nil {
		return -2
	}

	if t1.After(t2) {
		return 1
	}
	if t1.Before(t2) {
		return -1
	}
	return 0
}

// int64 2 string
func Int64ToString(arg int64) string {
	return strconv.FormatInt(arg, 10)
}

func Str2int(val string) int {
	ret, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return ret
}

func str2Map(jsonData string) (result map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(jsonData), &result)
	return result, err
}

func map2Str(mapData map[string]interface{}) (result string, err error) {
	resultByte, errError := json.Marshal(mapData)
	result = string(resultByte)
	err = errError
	return result, err
}

func StrToFloat(str string) float64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return float64(0)
	}
	return v
}
