package utils

import (
	log "web-server/alog"
	httpLib "web-server/http"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func IsStringIn(str string, strList []string) bool {
	for _, item := range strList {
		if item == str {
			return true
		}
	}
	return false
}

func StrToInt64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return int64(0)
	}
	return num
}

func StrToInt(str string) int {
	return int(StrToInt64(str))
}

/*
	通过sqlClause来获取不定字段的数据
	@sqlClause sql语句
	@return 返回map类型的[]，一行记录就是一个map，map里的key即为字段名。
		注意，字段名区分大小写，这里的字段名全部小写；如果失败返回nil, error
*/
func GetQueryDataList(db *sql.DB, sqlClause string) ([]map[string]string, error) {
	rows, err := db.Query(sqlClause)
	if err != nil {
		log.Error("query error: ", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	dataList := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		data := make(map[string]string)
		for i, col := range values {
			key := columns[i]
			key = strings.ToLower(key)
			data[key] = string(col)

		}
		dataList = append(dataList, data)
	}

	return dataList, nil
}

// 计算md5 32位小写
func GetMd5Value(cipherText string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(cipherText))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func IntoDataByMap(db *sql.DB, operator, tableName string, data map[string]interface{}) error {
	mpLen := len(data)
	valueList := make([]interface{}, mpLen)
	fieldListStr := " "
	tmpStr := " "
	i := 0
	for key, value := range data {
		valueList[i] = value
		if !strings.Contains(key, "`") {
			key = fmt.Sprintf("`%s`", key)
		}
		fieldListStr += key
		tmpStr = tmpStr + "?"
		if i < mpLen-1 {
			fieldListStr = fieldListStr + ","
			tmpStr = tmpStr + ","
		}
		i++
	}
	sqlClause := fmt.Sprintf("%s into %s (%s) values(%s)", operator, tableName, fieldListStr, tmpStr)
	_, err := db.Exec(sqlClause, valueList...)
	return err
}

/*
 * 通过map插入不定字段
 */
func InsertDataByMap(db *sql.DB, tableName string, insertMap map[string]interface{}) error {
	return IntoDataByMap(db, "insert", tableName, insertMap)
}

func InsertDataMap(db *sql.DB, tableName string, data map[string]string) error {
	tmp := make(map[string]interface{}, 0)
	for key, value := range data {
		tmp[key] = value
	}
	return IntoDataByMap(db, "insert", tableName, tmp)
}

func ReplaceDataByMap(db *sql.DB, tableName string, data map[string]interface{}) error {
	return IntoDataByMap(db, "replace", tableName, data)
}

// Replace : 非replace into(on duplicate key update) 实现
func ReplaceDataMap(db *sql.DB, tableName string, data map[string]string) (int64, error) {
	insertFieldsName := ""
	valuesList := ""
	updateList := ""

	for key, value := range data {
		if !strings.Contains(key, "`") {
			key = fmt.Sprintf("`%s`", key)
		}
		insertFieldsName += "," + key
		valuesList += fmt.Sprintf(",'%s'", value)
		updateList += fmt.Sprintf(",%s='%s'", key, value)
	}
	insertFieldsName = strings.TrimLeft(insertFieldsName, ",")
	valuesList = strings.TrimLeft(valuesList, ",")
	updateList = strings.TrimLeft(updateList, ",")

	sqlClause := fmt.Sprintf("insert into %s(%s) values(%s) on duplicate key update %s", tableName, insertFieldsName, valuesList, updateList)
	rs, err := db.Exec(sqlClause)
	if err != nil {
		log.Errorf("ReplaceDataMap error[%v] sqlClause[%s]", err, sqlClause)
		return 0, err
	}
	affectRows, _ := rs.RowsAffected()
	return affectRows, err
}

func UpdateDataByMap(db *sql.DB, tableName string,
	dataMap map[string]interface{}, sqlCondition string) error {
	sqlCluse := `update ` + tableName + ` set `
	fieldLen := len(dataMap)
	valueList := make([]interface{}, fieldLen)
	t := 0
	for key, value := range dataMap {
		valueList[t] = value
		if !strings.Contains(key, "`") {
			key = fmt.Sprintf("`%s`", key)
		}
		sqlCluse = sqlCluse + key + "=?"
		if t < fieldLen-1 {
			sqlCluse = sqlCluse + ","
		}
		t++
	}
	sqlCluse = sqlCluse + sqlCondition
	_, err := db.Exec(sqlCluse, valueList...)
	if err != nil {
		log.Error("SQL Error, SQL:", sqlCluse, err)
		return err
	}
	return nil
}

func UpdateDataMap(db *sql.DB, tableName, condition string, data map[string]string) error {
	tmp := make(map[string]interface{}, 0)
	for key, value := range data {
		tmp[key] = value
	}
	return UpdateDataByMap(db, tableName, tmp, condition)
}

type TaobaoIPResp struct {
	Code int
	Data interface{}
}

func GetProvinceAndCityFromIP(ip string) (string, string) {
	resp, err := http.Get(fmt.Sprintf("http://ip.taobao.com/service/getIpInfo.php?ip=%s", ip))
	if err != nil {
		return "", ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ""
	}

	taobaoResp := &TaobaoIPResp{}
	if err = json.Unmarshal(body, taobaoResp); err != nil {
		return "", ""
	}

	if taobaoResp.Code != 0 {
		errData := ""
		if errInfo, ok := taobaoResp.Data.(string); ok {
			errData = errInfo
		}
		log.Errorf("get ip error: ip = %s, code = %d, info = %s", ip, taobaoResp.Code, errData)
	} else {
		if data, ok := taobaoResp.Data.(map[string]interface{}); ok {
			//log.Debugf("country: %s", data["country"])
			//log.Debugf("region: %s", data["region"])
			//log.Debugf("city: %s", data["city"])
			return data["region"].(string), data["city"].(string)
		} else {
			log.Error("no data-mapping.")
		}
	}

	return "", ""
}

func GetCityFromIP(ip string) string {
	_, city := GetProvinceAndCityFromIP(ip)
	return city
}

func GetCityFromIPBaidu(ip string) string {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(1 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*1)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	resp, err := client.Get(fmt.Sprintf(`https://sp0.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php?query=%s&co=&resource_id=6006&t=1501146980039&ie=utf8&oe=utf8&cb=op_aladdin_callback&format=json&tn=baidu&cb=jQuery1102016651985494408095_1501146353506&_=1501146353516`, ip))
	if err != nil {
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	rege, _ := regexp.Compile(`^.*?"location":"(.*?)", ".*$`)
	out := rege.ReplaceAll(body, []byte("$1"))
	return string(out)
}

var TryTime = 2 // 重试次数

type Mobile struct {
	Province string `json:"prov"`    // 省份
	City     string `json:"city"`    // 城市
	Type     string `json:"type"`    // 运营商
	Mobile   string `json:"phoneno"` // 手机号码
}

type baiduResp struct {
	Status string   `json:"status"` // 状态，为0时为成功
	Data   []Mobile `json:"data"`   //
}

// 从因特利公共服务接口获取号码归属地
func MobileInfoIntelyService(mobile string) (Mobile, error, int64) {
	start := time.Now().UnixNano()
	runTime := func() int64 {
		return (time.Now().UnixNano() - start) / 1e6
	}

	httpUtil := httpLib.HttpUtils{}
	httpUtil.Init()
	httpUtil.SetTimeout(2) // 设置2秒超时

	mInfo := Mobile{}
	resp, err := httpUtil.Get("http://service.intely.cn/mobile_info?mobile="+mobile, "")
	if err != nil {
		return mInfo, err, runTime()
	}

	type Resp struct {
		Code       int               `json:"code"`
		Msg        string            `json:"msg"`
		MobileInfo map[string]string `json:"mobile_info"`
	}

	rs := Resp{}
	err = json.Unmarshal(resp, &rs)
	if err != nil {
		return mInfo, err, runTime()
	}
	return Mobile{Province: rs.MobileInfo["province"], City: rs.MobileInfo["city"], Mobile: rs.MobileInfo["mobile"], Type: rs.MobileInfo["company"]}, nil, runTime()
}


/*
 * 编码转换工具
 */



type BaiDuResp struct {
	Status string        `json:"status"` // 状态，为0时为成功
	Data   []interface{} `json:"data"`   //包含各种数据
}

func getDataFromBaiDu(query, resID string) ([]interface{}, error) {

	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(3 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*3)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	urlStr := fmt.Sprintf("http://opendata.baidu.com/api.php?query=%s&resource_id=%s&format=json&ie=utf8&oe=utf8", query, resID)
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var baiDuResp BaiDuResp
	err = json.Unmarshal(respByte, &baiDuResp)
	if err != nil {
		return nil, err
	}

	if baiDuResp.Status != "0" {
		return nil, errors.New("return no expected error")
	}

	return baiDuResp.Data, nil
}

/*
	返回IP地址归属地
	输入：
		ip	string	IP地址
	输出：
		string	省份+城市，格式为用下划线分割，如：广东_广州，
		失败时返回：_，调用者可直接根据strings.Split(loc,"_")函数进行获取省份和城市
*/
func GetIPLocationFromBaiDu(ip string) string {
	dataList, err := getDataFromBaiDu(ip, "6006")
	if err != nil {
		log.Error(err.Error())
		return "_"
	}

	if len(dataList) <= 0 {
		return "_"
	}

	data, ok := dataList[0].(map[string]interface{})
	if !ok {
		log.Error("getDataFromBaiDu>resp>data format error")
		return "_"
	}

	location, ok := data["location"].(string)
	if !ok {
		log.Error("getDataFromBaiDu>resp>data>location format error")
		return "_"
	}
	log.Debug(location)

	reg := regexp.MustCompile(`(.+省|.+自治区)?(.+市|.+自治州).*`)
	temp := reg.FindStringSubmatch(location)

	if len(temp) < 2 {
		return "_"
	}
	prov, city := "", ""
	if len(temp) == 2 {
		city = temp[1]
	}
	if len(temp) == 3 {
		prov = temp[1]
		city = temp[2]
	}

	prov = strings.Replace(prov, "省", "", -1)
	city = strings.Replace(city, "市", "", -1)

	if IsStringIn(city, []string{"北京", "上海", "天津", "重庆"}) {
		prov = city
	}

	newLoc := prov + "_" + city

	return newLoc
}
