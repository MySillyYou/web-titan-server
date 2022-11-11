package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
	log "web-server/alog"
	com "web-server/mysql"
)

const (
	NodeUnknown = iota
	NodeEdge
	NodeCandidate
	NodeScheduler
)

func Registers(w http.ResponseWriter, r *http.Request) {
	beginTime := r.FormValue("begin_time")
	endTime := r.FormValue("end_time")
	action := r.FormValue("action")
	pageSize := r.FormValue("page_row")
	page := r.FormValue("page")
	sortType := r.FormValue("sort_type")
	sortField := r.FormValue("sort_field")
	channel := r.FormValue("channel")
	source := r.FormValue("source")
	referer := r.FormValue("event_module")
	platform := r.FormValue("platform")
	if beginTime == "" || endTime == "" {
		HandleError(w, "参数错误")
		return
	}
	authCond := getAuthChannelsCond(r)
	t, _ := time.Parse("2006-01-02", endTime)
	endTime = t.Add(time.Hour*24 - time.Second).Format("2006-01-02 15:04:05")
	u, _ := time.Parse("2006-01-02", beginTime)
	beginTime = u.Add(-time.Second).Format("2006-01-02 15:04:05")
	//查询数量
	time_format := "'%Y-%m-%d'"
	sql_count := fmt.Sprintf("select count(distinct platform,source,channel,DATE_FORMAT(update_time,%s)) as num from act_flow where %s", time_format, authCond)
	//查询数据
	sqlClause := ""
	if authCond == "false" || authCond == "true" {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,referer,count(event_id='details_page' or null) as num_details,count(event_id='click_pdl_apply' or null) as num_apply from act_flow a,channel_info b,customer_data c where c.channel=b.channel and a.user_id=c.id and %s", time_format, authCond)
	} else {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,referer,count(event_id='details_page' or null) as num_details,count(event_id='click_pdl_apply' or null) as num_apply from act_flow a,channel_info b,customer_data c where c.channel=b.channel and a.user_id=c.id and a.%s", time_format, authCond)
	}
	if endTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time<='%s'", endTime)
		sql_count += fmt.Sprintf(" and update_time<='%s'", endTime)
	}
	if beginTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time>='%s'", beginTime)
		sql_count += fmt.Sprintf(" and update_time>='%s'", beginTime)
	}
	if channel != "" {
		sqlClause += fmt.Sprintf(" and c.channel='%s'", channel)
		sql_count += fmt.Sprintf(" and channel='%s'", channel)
	}
	if source != "" {
		sqlClause += fmt.Sprintf(" and c.source like '%%%s%%'", source)
		sql_count += fmt.Sprintf(" and source like '%%%s%%'", source)
	}
	if referer != "" {
		sqlClause += fmt.Sprintf(" and referer like '%%%s%%'", referer)
		sql_count += fmt.Sprintf(" and referer like '%%%s%%'", referer)
	}
	if platform != "" {
		sqlClause += fmt.Sprintf(" and c.platform = '%s'", platform)
		sql_count += fmt.Sprintf(" and platform = '%s'", platform)
	}
	sqlClause += " group by DATE_FORMAT(a.update_time,'%Y-%m-%d'),c.source,c.channel,referer "
	log.Debug(sqlClause)
	cout_num, err := com.GetSQLHelper().GetQueryDataList(sql_count)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "")
		return
	}
	outMap := make(map[string]interface{})
	fields := []string{"datetime", "channel_name", "source", "platform", "referer", "num_details", "num_apply"}
	fieldNames := []string{"日期", "代理商渠道", "渠道通路", "设备", "app功能模块", "点击详情页总数", "点击小贷产品总数"}
	outMap["list"] = ""
	outMap["total_num"] = "0"
	outMap["fields"] = fields
	outMap["fields_name"] = fieldNames
	orderBy := ""
	if sortField != "" {
		if sortType != "" {
			orderBy = fmt.Sprintf(" order by %s %s", sortField, sortType)
		} else {
			orderBy = fmt.Sprintf(" order by %s asc", sortField)
		}
	}
	sqlClause += orderBy
	if action != "EXPORT" {
		beginLimit := (Str2Int(page) - 1) * Str2Int(pageSize)
		endLimit := Str2Int(pageSize)
		sqlClause += fmt.Sprintf(" limit %d,%d", beginLimit, endLimit)
	}
	dataList, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(cout_num) != 0 {
		outMap["total_num"] = Str2Int(cout_num[0]["num"])
	} else {
		outMap["total_num"] = Str2Int("0")
	}
	outMap["list"] = dataList
	if action == "EXPORT" {
		contents := outMap["list"].([]map[string]string)
		fields := outMap["fields"].([]string)
		fieldNames := outMap["fields_name"].([]string)

		xlsxFile, err := MakeXslxFileWithFieldNamesFromMapList(fieldNames, fields, contents)
		if err != nil {
			log.Error(err.Error())
			HandleError(w, err.Error())
			return
		}

		userAgent := strings.ToLower(r.UserAgent())
		fileName := fmt.Sprintf("card_data_%s.xlsx", time.Now().Format("2006_01_02_15_04_05"))
		if strings.Index(userAgent, "msie") != -1 { //fileName用urlencode编码后正常
			w.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		} else if strings.Index(userAgent, "firefox") != -1 { //正常
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else if strings.Index(userAgent, "chrome") != -1 { //不能显示文件名,文件名只显示"file",而且没有扩展名
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		xlsxFile.Write(w)
		return
	}
	HandleSuccess(w, outMap)
}

func Register(w http.ResponseWriter, r *http.Request) {
	// todo 验证钱包Id
	var UserinfoOld UserInfo
	var Userinfo UserInfo
	Userinfo.WalletId = r.FormValue("wallet_id")
	Userinfo.Name = r.FormValue("name")
	Userinfo.UserId = r.FormValue("user_id")
	Userinfo.IdCard = r.FormValue("id_card")
	if Userinfo.WalletId == "" || Userinfo.Name == "" {
		HandleError(w, "参数错误")
		return
	}

	result := GMysqlDb.Where("wallet_id != ?", Userinfo.WalletId).Where("name = ?", Userinfo.Name).First(&UserinfoOld)
	if result.RowsAffected > 0 {
		HandleError(w, "参数错误")
		return
	}
	result = GMysqlDb.Where("wallet_id = ?", Userinfo.WalletId).First(&UserinfoOld)
	if result.RowsAffected <= 0 {
		strID := RandAllString(10)
		Userinfo.UserId = strID
		err := GMysqlDb.Create(&Userinfo).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}
		HandleSuccess(w, "注册成功")
	} else {
		Userinfo.ID = UserinfoOld.ID
		Userinfo.UserId = UserinfoOld.UserId
		err := GMysqlDb.Save(&Userinfo).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}
		HandleSuccess(w, "收拾收拾")
	}
}

func DeviceCreate(w http.ResponseWriter, r *http.Request) {
	nodeType := r.FormValue("node_type")
	nodeTypeInt := Str2Int(nodeType)
	deviceID, err := newDeviceID(nodeTypeInt)
	if err != nil {
		return
	}
	res := make(map[string]interface{})
	secret := newSecret(deviceID)
	var incomeDaily DevicesInfo
	incomeDaily.DeviceId = deviceID
	incomeDaily.Secret = secret
	incomeDaily.NodeType = nodeTypeInt
	res["device_id"] = deviceID
	res["secret"] = secret
	result := GMysqlDb.Where("device_id = ?", incomeDaily.DeviceId).First(&incomeDaily)
	if result.RowsAffected <= 0 {
		err := GMysqlDb.Create(&incomeDaily).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}

		HandleSuccess(w, res)
	} else {
		HandleSuccess(w, "创建失败")
	}
	//
	go GDevice.getDeviceIds()
}
func DeviceBiding(w http.ResponseWriter, r *http.Request) {
	var incomeDaily DevicesInfo
	incomeDaily.DeviceId = r.FormValue("device_id")
	incomeDaily.UserId = r.FormValue("userId")
	var incomeDailyOld DevicesInfo

	result := GMysqlDb.Where("device_id = ?", incomeDaily.DeviceId).First(&incomeDailyOld)
	if result.RowsAffected <= 0 {
		err := GMysqlDb.Create(&incomeDaily).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}
		HandleSuccess(w, "绑定成功")
	} else {
		result := GMysqlDb.Where("device_id = ?", incomeDaily.DeviceId).Where("user_id = ?", incomeDaily.UserId).First(&incomeDailyOld)
		if result.RowsAffected <= 0 {
			incomeDailyOld.UserId = incomeDaily.UserId
			err := GMysqlDb.Save(&incomeDailyOld).Error
			if err != nil {
				log.Error(err.Error())
				HandleError(w, "参数错误")
				return
			}
			HandleSuccess(w, "绑定成功")
		}
	}
}

func newDeviceID(nodeType int) (string, error) {
	u2, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	s := strings.Replace(u2.String(), "-", "", -1)
	switch nodeType {
	case NodeEdge:
		s = fmt.Sprintf("e_%s", s)
		return s, nil
	case NodeCandidate:
		s = fmt.Sprintf("c_%s", s)
		return s, nil
	}

	return "", xerrors.Errorf("nodetype err:%v", nodeType)
}

func newSecret(input string) string {
	c := sha1.New()
	c.Write([]byte(input))
	bytes := c.Sum(nil)
	return hex.EncodeToString(bytes)
}

var CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

/*
RandAllString

	lenNum
*/
func RandAllString(lenNum int) string {
	str := strings.Builder{}
	length := len(CHARS)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < lenNum; i++ {
		l := CHARS[rand.Intn(length)]
		str.WriteString(l)
	}
	return str.String()
}

func GetDetailData(w http.ResponseWriter, r *http.Request) {
	beginTime := r.FormValue("begin_time")
	endTime := r.FormValue("end_time")
	action := r.FormValue("action")
	pageSize := r.FormValue("page_row")
	page := r.FormValue("page")
	sortType := r.FormValue("sort_type")
	sortField := r.FormValue("sort_field")
	channel := r.FormValue("channel")
	source := r.FormValue("source")
	referer := r.FormValue("event_module")
	platform := r.FormValue("platform")
	if beginTime == "" || endTime == "" {
		HandleError(w, "参数错误")
		return
	}
	authCond := getAuthChannelsCond(r)
	t, _ := time.Parse("2006-01-02", endTime)
	endTime = t.Add(time.Hour*24 - time.Second).Format("2006-01-02 15:04:05")
	u, _ := time.Parse("2006-01-02", beginTime)
	beginTime = u.Add(-time.Second).Format("2006-01-02 15:04:05")
	//查询数量
	time_format := "'%Y-%m-%d'"
	sql_count := fmt.Sprintf("select count(distinct platform,source,channel,DATE_FORMAT(update_time,%s)) as num from act_flow where %s ", time_format, authCond)
	//查询数据
	sqlClause := ""
	if authCond == "true" || authCond == "false" {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,product_name,count(event_id='details_page' or null) as num_details,count(event_id='click_pdl_apply' or null) as num_apply from act_flow a,channel_info b,customer_data c,product_info d where c.channel=b.channel and a.user_id=c.id and  a.pro_id=d.product_id and %s ", time_format, authCond)
	} else {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,product_name,count(event_id='details_page' or null) as num_details,count(event_id='click_pdl_apply' or null) as num_apply from act_flow a,channel_info b,customer_data c,product_info d where c.channel=b.channel and a.user_id=c.id and  a.pro_id=d.product_id and a.%s", time_format, authCond)
	}

	if endTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time<='%s'", endTime)
		sql_count += fmt.Sprintf(" and update_time<='%s'", endTime)
	}
	if beginTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time>='%s'", beginTime)
		sql_count += fmt.Sprintf(" and update_time>='%s'", beginTime)
	}
	if channel != "" {
		sqlClause += fmt.Sprintf(" and c.channel='%s'", channel)
		sql_count += fmt.Sprintf(" and channel='%s'", channel)
	}
	if source != "" {
		sqlClause += fmt.Sprintf(" and c.source like '%%%s%%'", source)
		sql_count += fmt.Sprintf(" and source like '%%%s%%'", source)
	}
	if referer != "" {
		sqlClause += fmt.Sprintf(" and referer like '%%%s%%'", referer)
		sql_count += fmt.Sprintf(" and referer like '%%%s%%'", referer)
	}
	if platform != "" {
		sqlClause += fmt.Sprintf(" and c.platform = '%s'", platform)
		sql_count += fmt.Sprintf(" and platform = '%s'", platform)
	}
	sqlClause += " group by DATE_FORMAT(a.update_time,'%Y-%m-%d'),c.source,a.channel,pro_id "
	log.Debug(sqlClause)
	cout_num, err := com.GetSQLHelper().GetQueryDataList(sql_count)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "")
		return
	}
	outMap := make(map[string]interface{})
	fields := []string{"datetime", "channel_name", "source", "platform", "product_name", "num_details", "num_apply"}
	fieldNames := []string{"日期", "代理商渠道", "渠道通路", "设备", "小贷", "点击详情页总数", "点击小贷产品总数"}
	outMap["list"] = ""
	outMap["total_num"] = "0"
	outMap["fields"] = fields
	outMap["fields_name"] = fieldNames
	orderBy := ""
	if sortField != "" {
		if sortType != "" {
			orderBy = fmt.Sprintf(" order by %s %s", sortField, sortType)
		} else {
			orderBy = fmt.Sprintf(" order by %s asc", sortField)
		}
	}
	sqlClause += orderBy
	if action != "EXPORT" {
		beginLimit := (Str2Int(page) - 1) * Str2Int(pageSize)
		endLimit := Str2Int(pageSize)
		sqlClause += fmt.Sprintf(" limit %d,%d", beginLimit, endLimit)
	}
	dataList, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(cout_num) != 0 {
		outMap["total_num"] = Str2Int(cout_num[0]["num"])
	} else {
		outMap["total_num"] = Str2Int("0")
	}
	outMap["list"] = dataList
	if action == "EXPORT" {
		contents := outMap["list"].([]map[string]string)
		fields := outMap["fields"].([]string)
		fieldNames := outMap["fields_name"].([]string)

		xlsxFile, err := MakeXslxFileWithFieldNamesFromMapList(fieldNames, fields, contents)
		if err != nil {
			log.Error(err.Error())
			HandleError(w, err.Error())
			return
		}

		userAgent := strings.ToLower(r.UserAgent())
		fileName := fmt.Sprintf("card_data_%s.xlsx", time.Now().Format("2006_01_02_15_04_05"))
		if strings.Index(userAgent, "msie") != -1 { //fileName用urlencode编码后正常
			w.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		} else if strings.Index(userAgent, "firefox") != -1 { //正常
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else if strings.Index(userAgent, "chrome") != -1 { //不能显示文件名,文件名只显示"file",而且没有扩展名
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		xlsxFile.Write(w)
		return
	}
	HandleSuccess(w, outMap)
}

func GetInnerRegistData(w http.ResponseWriter, r *http.Request) {
	beginTime := r.FormValue("begin_time")
	endTime := r.FormValue("end_time")
	action := r.FormValue("action")
	pageSize := r.FormValue("page_row")
	page := r.FormValue("page")
	sortType := r.FormValue("sort_type")
	sortField := r.FormValue("sort_field")
	channel := r.FormValue("channel")
	source := r.FormValue("source")
	referer := r.FormValue("referer")
	platform := r.FormValue("platform")
	if beginTime == "" || endTime == "" {
		HandleError(w, "参数错误")
		return
	}
	authCond := getAuthChannelsCond(r)
	t, _ := time.Parse("2006-01-02", endTime)
	endTime = t.Add(time.Hour*24 - time.Second).Format("2006-01-02 15:04:05")
	u, _ := time.Parse("2006-01-02", beginTime)
	beginTime = u.Add(-time.Second).Format("2006-01-02 15:04:05")
	//查询数量
	time_format := "'%Y-%m-%d'"
	sql_count := fmt.Sprintf("select count(distinct platform,source,channel,DATE_FORMAT(update_time,%s)) as num from act_flow where %s ", time_format, authCond)
	//查询数据
	sqlClause := ""
	if authCond == "false" || authCond == "true" {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,pro_id,count(a.event_id='get_cus_regist' or null) as num_regist,count(event_id='get_cus_regist' and event_times='1' or null) as num_apply,count(event_id='get_cus_upload' or null) as num_upload from act_flow  a,channel_info  b,customer_data c where c.channel=b.channel and a.user_id=c.id and %s ", time_format, authCond)
	} else {
		sqlClause = fmt.Sprintf("select DATE_FORMAT(a.update_time,%s) AS dateTime,c.source,b.channel_name as channel_name,c.platform,pro_id,count(a.event_id='get_cus_regist' or null) as num_regist,count(event_id='get_cus_regist' and event_times='1' or null) as num_apply,count(event_id='get_cus_upload' or null) as num_upload from act_flow  a,channel_info  b,customer_data c where c.channel=b.channel and a.user_id=c.id and a.%s ", time_format, authCond)
	}

	if endTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time<='%s'", endTime)
		sql_count += fmt.Sprintf(" and update_time<='%s'", endTime)
	}
	if beginTime != "" {
		sqlClause += fmt.Sprintf(" and a.update_time>='%s'", beginTime)
		sql_count += fmt.Sprintf(" and update_time>='%s'", beginTime)
	}
	if channel != "" {
		sqlClause += fmt.Sprintf(" and c.channel='%s'", channel)
		sql_count += fmt.Sprintf(" and channel='%s'", channel)
	}
	if source != "" {
		sqlClause += fmt.Sprintf(" and c.source like '%%%s%%'", source)
		sql_count += fmt.Sprintf(" and source like '%%%s%%'", source)
	}
	if referer != "" {
		sqlClause += fmt.Sprintf(" and referer like '%%%s%%'", referer)
		sql_count += fmt.Sprintf(" and referer like '%%%s%%'", referer)
	}
	if platform != "" {
		sqlClause += fmt.Sprintf(" and c.platform = '%s'", platform)
		sql_count += fmt.Sprintf(" and platform = '%s'", platform)
	}
	sqlClause += " group by DATE_FORMAT(a.update_time,'%Y-%m-%d'),c.source,c.channel,c.platform "
	cout_num, err := com.GetSQLHelper().GetQueryDataList(sql_count)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "")
		return
	}
	outMap := make(map[string]interface{})
	fields := []string{"datetime", "channel_name", "source", "platform", "num_regist", "num_apply", "num_upload"}
	fieldNames := []string{"日期", "代理商渠道", "渠道通路", "设备", "注册量", "新增注册量", "下载量"}
	outMap["list"] = ""
	outMap["total_num"] = "0"
	outMap["fields"] = fields
	outMap["fields_name"] = fieldNames
	orderBy := ""
	if sortField != "" {
		if sortType != "" {
			orderBy = fmt.Sprintf(" order by %s %s", sortField, sortType)
		} else {
			orderBy = fmt.Sprintf(" order by %s asc", sortField)
		}
	}
	sqlClause += orderBy
	if action != "EXPORT" {
		beginLimit := (Str2Int(page) - 1) * Str2Int(pageSize)
		endLimit := Str2Int(pageSize)
		sqlClause += fmt.Sprintf(" limit %d,%d", beginLimit, endLimit)
	}
	dataList, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(cout_num) != 0 {
		outMap["total_num"] = Str2Int(cout_num[0]["num"])
	} else {
		outMap["total_num"] = Str2Int("0")
	}
	outMap["list"] = dataList
	if action == "EXPORT" {
		contents := outMap["list"].([]map[string]string)
		fields := outMap["fields"].([]string)
		fieldNames := outMap["fields_name"].([]string)

		xlsxFile, err := MakeXslxFileWithFieldNamesFromMapList(fieldNames, fields, contents)
		if err != nil {
			log.Error(err.Error())
			HandleError(w, err.Error())
			return
		}

		userAgent := strings.ToLower(r.UserAgent())
		fileName := fmt.Sprintf("card_data_%s.xlsx", time.Now().Format("2006_01_02_15_04_05"))
		if strings.Index(userAgent, "msie") != -1 { //fileName用urlencode编码后正常
			w.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		} else if strings.Index(userAgent, "firefox") != -1 { //正常
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else if strings.Index(userAgent, "chrome") != -1 { //不能显示文件名,文件名只显示"file",而且没有扩展名
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		xlsxFile.Write(w)
		return
	}
	HandleSuccess(w, outMap)
}

func GetUploadData(w http.ResponseWriter, r *http.Request) {
	beginTime := r.FormValue("begin_time")
	endTime := r.FormValue("end_time")
	action := r.FormValue("action")
	pageSize := r.FormValue("page_row")
	pagenum := r.FormValue("page_num")
	page := r.FormValue("page")
	sortType := r.FormValue("sort_type")
	sortField := r.FormValue("sort_field")
	channel := r.FormValue("channel")
	source := r.FormValue("source")
	if beginTime == "" || endTime == "" {
		HandleError(w, "参数错误")
		return
	}
	authCond := getAuthChannelsCond(r)
	t, _ := time.Parse("2006-01-02", endTime)
	endTime = t.Add(time.Hour*24 - time.Second).Format("2006-01-02 15:04:05")
	u, _ := time.Parse("2006-01-02", beginTime)
	beginTime = u.Add(-time.Second).Format("2006-01-02 15:04:05")
	//查询数量
	time_format := "'%Y-%m-%d'"
	sql_count := fmt.Sprintf("select count(distinct DATE_FORMAT(apply_time,%s)) as num from customer_data where %s and char_2=''", time_format, authCond)
	//查询数据
	sqlClause := fmt.Sprintf("select DATE_FORMAT(apply_time,%s) AS dateTime,count(distinct mobile) as num from customer_data where %s and char_2=''", time_format, authCond)
	sqlClause_EXPORT := fmt.Sprintf("select id,name,mobile,sex,channel,source,apply_time from customer_data where %s", authCond)
	//select DATE_FORMAT(apply_time,'%Y-%m-%d') AS dateTime,count(distinct mobile) as num from customer_data where apply_time>'2019-04-10' group by DATE_FORMAT(apply_time,'%Y-%m-%d');
	if endTime != "" {
		sqlClause += fmt.Sprintf(" and apply_time<='%s'", endTime)
		sql_count += fmt.Sprintf(" and apply_time<='%s'", endTime)
		sqlClause_EXPORT += fmt.Sprintf(" and apply_time<='%s'", endTime)
	}
	if beginTime != "" {
		sqlClause += fmt.Sprintf(" and apply_time>='%s'", beginTime)
		sql_count += fmt.Sprintf(" and apply_time>='%s'", beginTime)
		sqlClause_EXPORT += fmt.Sprintf(" and apply_time>='%s'", beginTime)
	}
	if channel != "" {
		sqlClause += fmt.Sprintf(" and channel='%s'", channel)
		sql_count += fmt.Sprintf(" and channel='%s'", channel)
		sqlClause_EXPORT += fmt.Sprintf(" and channel='%s'", channel)
	}
	if source != "" {
		sqlClause += fmt.Sprintf(" and source like '%%%s%%'", source)
		sql_count += fmt.Sprintf(" and source like '%%%s%%'", source)
		sqlClause_EXPORT += fmt.Sprintf(" and source like '%%%s%%'", source)
	}
	sqlClause += " group by DATE_FORMAT(apply_time,'%Y-%m-%d')"
	sqlClause_EXPORT += fmt.Sprintf(" and char_2='' limit %s", pagenum)
	cout_num, err := com.GetSQLHelper().GetQueryDataList(sql_count)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "")
		return
	}
	outMap := make(map[string]interface{})
	fields := []string{"datetime", "num"}
	fieldNames := []string{"日期", "数量"}
	outMap["list"] = ""
	outMap["total_num"] = "0"
	outMap["fields"] = fields
	outMap["fields_name"] = fieldNames
	orderBy := ""
	if sortField != "" {
		if sortType != "" {
			orderBy = fmt.Sprintf(" order by %s %s", sortField, sortType)
		} else {
			orderBy = fmt.Sprintf(" order by %s asc", sortField)
		}
	}
	sqlClause += orderBy
	if action != "EXPORT" {
		beginLimit := (Str2Int(page) - 1) * Str2Int(pageSize)
		endLimit := Str2Int(pageSize)
		sqlClause += fmt.Sprintf(" limit %d,%d", beginLimit, endLimit)
	}
	if action == "EXPORT" {
		sqlClause = sqlClause_EXPORT
	}
	dataList, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(cout_num) != 0 {
		outMap["total_num"] = Str2Int(cout_num[0]["num"])
	} else {
		outMap["total_num"] = Str2Int("0")
	}
	outMap["list"] = dataList
	if action == "EXPORT" {
		for _, data_up := range dataList {
			data_up_data := make(map[string]interface{})
			data_up_data["char_2"] = "Y"
			id := data_up["id"]
			_, err := com.GetSQLHelper().UpdateDataByMap("customer_data", data_up_data, fmt.Sprintf(" where id=%s", id))
			if err != nil {
				log.Errorf("Save update one data to customer_data failed, data[%v]  error[%v]", data_up_data, err)
				HandleError(w, err.Error())
				return
			}
		}
		contents := outMap["list"].([]map[string]string)
		fields := []string{"id", "name", "mobile", "sex", "channel", "source", "apply_time"}
		fieldNames := []string{"用户id", "姓名", "手机号码", "性别", "通路", "渠道", "时间"}

		xlsxFile, err := MakeXslxFileWithFieldNamesFromMapList(fieldNames, fields, contents)
		if err != nil {
			log.Error(err.Error())
			HandleError(w, err.Error())
			return
		}

		userAgent := strings.ToLower(r.UserAgent())
		fileName := fmt.Sprintf("card_data_%s.xlsx", time.Now().Format("2006_01_02_15_04_05"))
		if strings.Index(userAgent, "msie") != -1 { //fileName用urlencode编码后正常
			w.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		} else if strings.Index(userAgent, "firefox") != -1 { //正常
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else if strings.Index(userAgent, "chrome") != -1 { //不能显示文件名,文件名只显示"file",而且没有扩展名
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		xlsxFile.Write(w)
		return
	}
	HandleSuccess(w, outMap)
}
