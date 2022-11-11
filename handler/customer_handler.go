package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	log "web-server/alog"
	httpUtil "web-server/http"
	com "web-server/mysql"
	red "web-server/redis"
	csms "web-server/sms"
	util "web-server/utils"
)

func GetPhoneCode(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	//bank := r.FormValue("bank")
	if phone == "" {
		HandleError(w, KErrorMsg[KErrorNoArg])
		return
	}
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	verifyCode := fmt.Sprintf("%06d", rander.Intn(1000000))
	log.Debug("verifyCode", verifyCode)

	expireTime := int64(300) //默认5分钟
	redisKey := "verify_Code" + phone

	ttl, err := red.Redis.GetTTL(redisKey)
	if err == nil && (expireTime-ttl) < 60 { //发送间隔小于1分钟
		HandleError(w, "验证码已发送，请稍后查看")
		return
	}

	err = red.Redis.SetString(redisKey, verifyCode) // 随机的验证码写入redis缓存
	if err != nil {
		log.Errorf("GetPhoneCode save verify code failed, error[%v]", err)
		HandleError(w, "获取验证码失败")
		return
	}
	fmt.Println(verifyCode)
	red.Redis.SetTTL(redisKey, expireTime) // 设置验证码过期时间
	//if gConfig["sms"].(bool) {
	//	if bank == "pufa" {
	//		err = sms.SendTencentSMSCode("1400162782", "184d88f63ea11ca656db27a28bddbd7d", "浦发银行", "235573", phone, verifyCode)
	//	}
	//	if bank == "guangda" {
	//		err = sms.SendTencentSMSCode("1400165367", "7d0fded13083fce311138837718894a2", "光大银行", "238326", phone, verifyCode)
	//	}
	//	if bank == "xingye" {
	//		err = sms.SendTencentSMSCode("1400165369", "ff599d7b94a2e937e495f1f9c046ffc2", "兴业银行", "238330", phone, verifyCode)
	//	}
	//} else {
	//	err = nil // for debug
	//}
	//appKey, phone, smsName, smsTemplateCode, code, secret
	if gConfig["sms"].(bool) {
		//err = csms.UsingAlidayuSendSMS("24692602", phone, "因地利", "SMS_90180036", verifyCode, "c279b2fea52239946ec15b4ac4659c3b")
		err = csms.UsingAlidayuSendSMS("24692602", phone, "好幸贷", "SMS_110255035", verifyCode, "c279b2fea52239946ec15b4ac4659c3b")
		if err != nil {
			log.Errorf("GetPhoneCode send sms message failed, error[%v]", err)
			HandleError(w, "短信验证码不正确，请稍后再试!")
			return
		}
	}
	HandleSuccess(w, "发送成功")
}

func GetIndexInfo(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	fmt.Println(phone)
	var dataRes IndexPageRes
	dataRes.StorageT = 1080.99
	dataRes.BandwidthMb = 666.99
	// AllMinerNum MinerInfo
	dataRes.AllCandidate = AllM.AllCandidate
	dataRes.AllEdgeNode = AllM.AllEdgeNode
	dataRes.AllVerifier = AllM.AllVerifier
	// OnlineMinerNum MinerInfo
	dataRes.OnlineCandidate = 11
	dataRes.OnlineEdgeNode = 252
	dataRes.OnlineVerifier = 88
	// Devices
	dataRes.AbnormalNum = 12
	dataRes.OfflineNum = 12
	dataRes.OnlineNum = 12
	dataRes.TotalNum = 36
	// Profit
	dataRes.CumulativeProfit = 36
	dataRes.MonthProfit = 36
	dataRes.SevenDaysProfit = 36
	dataRes.YesterdayProfit = 36
	HandleSuccess(w, dataRes)
}

// GetUserDeviceInfo 设备总览
func GetUserDeviceInfo(w http.ResponseWriter, r *http.Request) {
	from := r.FormValue("from")
	to := r.FormValue("to")
	var pa DevicesSearch
	pa.UserId = r.FormValue("userId")
	fmt.Println(pa.UserId)
	DeviceId := r.FormValue("device_id")
	pa.DeviceId = DeviceId
	pa.DeviceStatus = r.FormValue("device_status")
	list, total, err := GetDevicesInfoList(pa)
	if err != nil {
		log.Error("args error")
		HandleError(w, KErrorMsg[KErrorNoArg])
	}
	var dataList []DevicesInfo
	var res DevicesInfoPage
	var dataRes IndexUserDeviceRes
	for _, data := range list {
		err = getProfitByDeviceId(&data, &res)
		dataRes.CumulativeProfit += data.CumuProfit
		dataRes.TodayProfit += data.TodayProfit
		dataRes.SevenDaysProfit += data.SevenDaysProfit
		dataRes.YesterdayProfit += data.YesterdayIncome
		dataRes.MonthProfit += data.MonthProfit
		if err != nil {
			log.Error("getProfitByDeviceId：", data.DeviceId)
		}
		dataList = append(dataList, data)
	}

	// Devices
	dataRes.AbnormalNum = res.Abnormal
	dataRes.OfflineNum = res.Offline
	dataRes.OnlineNum = res.Online
	dataRes.TotalNum = total
	dataRes.TotalBandwidth = res.BandwidthMb
	// Profit
	var p IncomeDailySearch
	p.DateFrom = from
	p.DateTo = to
	//p.UserId = userId
	p.UserId = pa.UserId
	m := GetIncomeAllList(p)
	dataRes.DailyIncome = m
	HandleSuccess(w, dataRes)
}

func timeFormat(p IncomeDailySearch) (m map[string]interface{}) {
	timeNow := time.Now().Format("2006-01-02")
	// 默认两周的数据
	dd, _ := time.ParseDuration("-24h")
	FromTime := time.Now().Add(dd * 14).Format("2006-01-02")
	if p.DateFrom == "" && p.Date == "" {
		p.DateFrom = FromTime
	}
	if p.DateTo == "" && p.Date == "" {
		p.DateTo = timeNow
	}
	p.DateFrom = p.DateFrom + " 00:00:00"
	p.DateTo = p.DateTo + " 23:59:59"
	m = getDaysData(p)

	return m
}

func timeFormatHour(p IncomeDailySearch) (m map[string]interface{}) {
	timeNow := time.Now().Format("2006-01-02")
	// 默认两周的数据
	dd, _ := time.ParseDuration("-24h")
	FromTime := time.Now().Add(dd * 14).Format("2006-01-02")
	if p.DateFrom == "" && p.Date == "" {
		p.DateFrom = FromTime
	}
	if p.DateTo == "" && p.Date == "" {
		p.DateTo = timeNow
	}
	if p.Date == "" {
		p.Date = time.Now().Format("2006-01-02")
	}
	p.DateFrom = p.Date + " 00:00:00"
	p.DateTo = p.Date + " 23:59:59"
	m = getDaysDataHour(p)

	return m
}

func getDaysDataHour(p IncomeDailySearch) (returnMapList map[string]interface{}) {
	list, _, err := GetIncomeDailyHourList(p)
	if err != nil {
		return
	}
	returnMap := make(map[string]interface{})
	queryMapTo := make(map[string]float64)
	pkgLossRatioTo := make(map[string]float64)
	latencyTo := make(map[string]float64)
	onlineJsonDailyTo := make(map[string]float64)
	natTypeTo := make(map[string]float64)
	diskUsageTo := make(map[string]float64)
	incomeHourBefore := float64(0)
	onlineHourBefore := float64(0)
	firstData := true
	for _, v := range list {
		//timeStr:=v.Time.Format(TimeFormatMD)
		timeStr := v.Time.Format(TimeFormatHM)
		// 每天第一条数据用于对比在线时长和收入
		if firstData {
			incomeHourBefore = v.JsonDaily
			onlineHourBefore = v.OnlineJsonDaily
			firstData = false
			continue
		}
		timeMinStr := v.Time.Format(TimeFormatM)
		if timeMinStr == "00" {
			queryMapTo[timeStr] = v.JsonDaily - incomeHourBefore
			incomeHourBefore = v.JsonDaily
			onlineJsonDailyTo[timeStr] = v.OnlineJsonDaily - onlineHourBefore
			onlineHourBefore = v.OnlineJsonDaily
		}
		if timeMinStr == "00" || timeMinStr == "30" {
			pkgLossRatioTo[timeStr] = v.PkgLossRatio * 100
			latencyTo[timeStr] = v.Latency
			natTypeTo[timeStr] = v.NatType
			diskUsageTo[timeStr] = v.DiskUsage
		}
	}
	returnMap["income"] = queryMapTo
	returnMap["online"] = onlineJsonDailyTo
	returnMap["pkgLoss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["natType"] = natTypeTo
	returnMap["diskUsage"] = diskUsageTo
	returnMapList = returnMap
	return
}

func getDaysData(p IncomeDailySearch) (returnMapList map[string]interface{}) {
	list, _, err := GetIncomeDailyList(p)
	if err != nil {
		return
	}
	returnMap := make(map[string]interface{})
	queryMapTo := make(map[string]float64)
	pkgLossRatioTo := make(map[string]float64)
	latencyTo := make(map[string]float64)
	onlineJsonDailyTo := make(map[string]float64)
	natTypeTo := make(map[string]float64)
	diskUsageTo := make(map[string]float64)
	for _, v := range list {
		timeStr := v.Time.Format(TimeFormatMD)
		//timeStr:=v.Time.Format(TimeFormatHM)
		queryMapTo[timeStr] += v.JsonDaily
		pkgLossRatioTo[timeStr] = v.PkgLossRatio
		latencyTo[timeStr] = v.Latency
		onlineJsonDailyTo[timeStr] = v.OnlineJsonDaily
		natTypeTo[timeStr] = v.NatType
		diskUsageTo[timeStr] = v.DiskUsage
	}
	returnMap["income"] = queryMapTo
	returnMap["online"] = onlineJsonDailyTo
	returnMap["pkgLoss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["natType"] = natTypeTo
	returnMap["diskUsage"] = diskUsageTo
	returnMapList = returnMap
	return
}

// getYearMonthToDay 查询指定年份指定月份有多少天
// @params year int 指定年份
// @params month int 指定月份
func getYearMonthToDay(year int, month int) int {
	// 有31天的月份
	day31 := map[int]struct{}{
		1:  {},
		3:  {},
		5:  {},
		7:  {},
		8:  {},
		10: {},
		12: {},
	}
	if _, ok := day31[month]; ok {
		return 31
	}
	// 有30天的月份
	day30 := map[int]struct{}{
		4:  {},
		6:  {},
		9:  {},
		11: {},
	}
	if _, ok := day30[month]; ok {
		return 30
	}
	// 计算是平年还是闰年
	if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
		// 得出2月的天数
		return 29
	}
	// 得出2月的天数
	return 28
}

func GetAllMinerInfo(w http.ResponseWriter, r *http.Request) {
	HandleSuccess(w, AllM)
}

func Retrieval(w http.ResponseWriter, r *http.Request) {
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.UserId = r.FormValue("userId")
	TaskInfoSearch.Status = r.FormValue("status")
	TaskInfoSearch.Cid = r.FormValue("cid")
	var res RetrievalPageRes
	list, total, err := GetTaskInfoList(TaskInfoSearch)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "参数错误")
		return
	}
	res.List = list
	res.Count = total
	// 后续通过调度器动态获取
	res.StorageT = AllM.StorageT
	res.BandwidthMb = AllM.BandwidthMb
	// AllMinerNum MinerInfo
	res.AllCandidate = AllM.AllCandidate
	res.AllEdgeNode = AllM.AllEdgeNode
	res.AllVerifier = AllM.AllVerifier
	HandleSuccess(w, res)
}

func GetDevicesInfo(w http.ResponseWriter, r *http.Request) {
	var res DevicesInfoPage
	var p DevicesSearch
	p.UserId = r.FormValue("userId")
	fmt.Println(p.UserId)
	p.DeviceId = r.FormValue("device_id")
	p.DeviceStatus = r.FormValue("device_status")
	list, total, err := GetDevicesInfoList(p)
	if err != nil {
		log.Error("args error")
		HandleError(w, KErrorMsg[KErrorNoArg])
	}
	var dataList []DevicesInfo
	for _, data := range list {
		err = getProfitByDeviceId(&data, &res)
		if err != nil {
			log.Error("getProfitByDeviceId：", data.DeviceId)
		}
		dataList = append(dataList, data)
	}
	res.List = dataList
	res.Count = total
	res.AllDevices = total
	HandleSuccess(w, res)
}

func GetDeviceDiagnosisDaily(w http.ResponseWriter, r *http.Request) {
	var p IncomeDailySearch
	from := r.FormValue("from")
	to := r.FormValue("to")
	p.DateFrom = from
	p.DateTo = to
	//p.UserId = r.FormValue("userId")
	p.DeviceId = r.FormValue("device_id")
	var res IncomeDailyRes
	m := timeFormat(p)
	res.DailyIncome = m
	//res.DefYesterday = "31.1%"
	//res.YesterdayProfit = 12.33
	//res.SevenDaysProfit = 32.33
	//res.CumulativeProfit = 212.33
	//res.MonthProfit = 112.33
	//res.TodayProfit = 112.33
	//res.OnlineTime = "12h"
	res.DeviceDiagnosis = "优秀"
	HandleSuccess(w, res)
}

func GetDeviceDiagnosisHour(w http.ResponseWriter, r *http.Request) {
	var p IncomeDailySearch
	//p.UserId = r.FormValue("userId")
	p.DeviceId = r.FormValue("device_id")
	p.Date = r.FormValue("date")
	var res IncomeDailyRes
	m := timeFormatHour(p)
	res.DailyIncome = m
	res.DeviceDiagnosis = "优秀"
	HandleSuccess(w, res)
}

//func timeFormatHour(p IncomeDailySearch) (m map[string]interface{}) {
//	// timeNow := time.Now().Format("2006-01-02")
//	if p.DateFrom == "" && p.Date == "" && p.DateTo == "" {
//		p.Date = "2022-06-30"
//	}
//	returnMapList := make(map[string]interface{})
//	// 单日数据
//	if p.Date != "" {
//		getHoursData(p, &returnMapList)
//	}
//	return returnMapList
//}

func getHoursData(p IncomeDailySearch, returnMapList *map[string]interface{}) {
	listHour, _, err := GetHourDailyList(p)
	if err != nil {
		return
	}
	onlineJsonDailyFrom := make(map[string]interface{})
	onlineJsonDaily := make(map[string]interface{})
	e := json.Unmarshal([]byte(listHour.OnlineJsonDaily), &onlineJsonDailyFrom)
	if e != nil {
		return
	}
	pkgLossRatioFrom := make(map[string]interface{})
	pkgLossRatio := make(map[string]interface{})
	e = json.Unmarshal([]byte(listHour.PkgLossRatio), &pkgLossRatioFrom)
	if e != nil {
		return
	}
	latencyFrom := make(map[string]interface{})
	latency := make(map[string]interface{})
	e = json.Unmarshal([]byte(listHour.Latency), &latencyFrom)
	if e != nil {
		return
	}
	natTypeFrom := make(map[string]interface{})
	natType := make(map[string]interface{})
	e = json.Unmarshal([]byte(listHour.NatType), &natTypeFrom)
	if e != nil {
		return
	}
	for i := 1; i <= 24; i++ {
		stringFrom := strconv.Itoa(i)
		if len(stringFrom) < 2 {
			stringFrom = "0" + stringFrom
		}
		onlineJsonDaily[stringFrom+":00"] = onlineJsonDailyFrom[stringFrom]
		pkgLossRatio[stringFrom+":00"] = pkgLossRatioFrom[stringFrom]
		latency[stringFrom+":00"] = latencyFrom[stringFrom]
		natType[stringFrom+":00"] = natTypeFrom[stringFrom]

	}
	returnMap := make(map[string]interface{})
	returnMap["online"] = onlineJsonDaily
	returnMap["pkgLoss"] = pkgLossRatio
	returnMap["latency"] = latency
	returnMap["natType"] = natType
	returnMap["diskUsage"] = "127.2G/500.0G"
	*returnMapList = returnMap
	return
}

// GetDevicesInfoList DevicesInfo search from mysql
func GetDevicesInfoList(info DevicesSearch) (list []DevicesInfo, total int64, err error) {
	// string转成int：
	limit, _ := strconv.Atoi(info.PageSize)
	page, _ := strconv.Atoi(info.Page)
	offset := limit * (page - 1)
	// 创建db
	db := GMysqlDb.Model(&DevicesInfo{})
	var InPages []DevicesInfo
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DeviceId != "" {
		db = db.Where("device_id = ?", info.DeviceId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.UserId != "" {
		db = db.Where("user_id = ?", info.UserId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DeviceStatus != "" && info.DeviceStatus != "allDevices" {
		db = db.Where("device_status = ?", info.DeviceStatus)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Find(&InPages).Error
	return InPages, total, err
}
func getProfitByDeviceId(rt *DevicesInfo, dt *DevicesInfoPage) error {
	switch rt.DeviceStatus {
	case "online":
		dt.Online += 1
	case "offline":
		dt.Offline += 1
	case "abnormal":
		dt.Abnormal += 1

	}
	dt.BandwidthMb += rt.BandwidthUp
	return nil
}

func GetIncomeDailyHourList(info IncomeDailySearch) (list []IncomeDaily, total int64, err error) {
	// string转成int：
	limit, _ := strconv.Atoi(info.PageSize)
	page, _ := strconv.Atoi(info.Page)
	offset := limit * (page - 1)
	// 创建db
	db := GMysqlDb.Model(&IncomeDaily{})
	var InPages []IncomeDaily
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DeviceId != "" {
		db = db.Where("device_id = ?", info.DeviceId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.UserId != "" {
		db = db.Where("user_id = ?", info.UserId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DateFrom != "" {
		db = db.Where("time >= ?", info.DateFrom)
	}
	if info.DateTo != "" {
		db = db.Where("time <= ?", info.DateTo)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Find(&InPages).Error
	err = db.Limit(limit).Offset(offset).Find(&InPages).Error
	return InPages, total, err
}

func GetIncomeDailyList(info IncomeDailySearch) (list []IncomeOfDaily, total int64, err error) {
	// string转成int：
	limit, _ := strconv.Atoi(info.PageSize)
	page, _ := strconv.Atoi(info.Page)
	offset := limit * (page - 1)
	// 创建db
	db := GMysqlDb.Model(&IncomeOfDaily{})
	var InPages []IncomeOfDaily
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DeviceId != "" {
		db = db.Where("device_id = ?", info.DeviceId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.UserId != "" {
		db = db.Where("user_id = ?", info.UserId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DateFrom != "" {
		db = db.Where("time >= ?", info.DateFrom)
	}
	if info.DateTo != "" {
		db = db.Where("time <= ?", info.DateTo)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Find(&InPages).Error
	err = db.Limit(limit).Offset(offset).Find(&InPages).Error
	return InPages, total, err
}

//func GetIncomeDailyList(info IncomeDailySearch) (list []IncomeOfDaily, total int64, err error) {
//		// string转成int：
//		//limit, _ := strconv.Atoi(info.PageSize)
//		//page, _ := strconv.Atoi(info.Page)
//		//offset := limit * (page - 1)
//		var InPages []IncomeOfDaily
//		sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, count(1) as num, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily_r "+
//			"where device_id='%s' and time>='%s' and time<='%s' group by date", info.DeviceId, info.DateFrom, info.DateTo)
//		if info.UserId != "" {
//			sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, count(1) as num, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily_r "+
//				"where user_id='%s' and time>='%s' and time<='%s' group by date", info.UserId, info.DateFrom, info.DateTo)
//		}
//		fmt.Println(sqlClause)
//		datas, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
//		for _, data := range datas {
//			fmt.Println(data)
//			var InPage IncomeOfDaily
//			InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
//			InPage.DiskUsage = Str2Float64(data["disk_usage"])
//			InPage.NatType = Str2Float64(data["nat_ratio"])
//			InPage.JsonDaily = Str2Float64(data["hour_income"])
//			InPage.OnlineJsonDaily = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
//			InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
//			InPage.Latency = Str2Float64(data["latency"])
//			InPages = append(InPages, InPage)
//			total += Str2Int64(data["num"])
//		}
//
//		return InPages, total, err
//	}

func GetIncomeAllList(info IncomeDailySearch) (list []map[string]interface{}) {
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, , sum(income) as income from income_daily_r "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", info.DeviceId, info.DateFrom, info.DateTo)
	if info.UserId != "" {
		sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, sum(income) as income from income_daily_r "+
			"where user_id='%s' and time>='%s' and time<='%s' group by date", info.UserId, info.DateFrom, info.DateTo)
	}
	datas, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("error[%v]", err)
		return
	}
	var mapIncomeList []map[string]interface{}
	for _, data := range datas {
		mapIncome := make(map[string]interface{})
		mapIncome["date"] = data["date"]
		mapIncome["income"] = Str2Float64(data["income"])
		mapIncomeList = append(mapIncomeList, mapIncome)
	}
	return mapIncomeList
}

func GetIncomeDailyListO(info IncomeDailySearch) (list []IncomeOfDaily, total int64, err error) {
	// string转成int：
	//limit, _ := strconv.Atoi(info.PageSize)
	//page, _ := strconv.Atoi(info.Page)
	//offset := limit * (page - 1)
	var InPages []IncomeOfDaily
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, count(1) as num, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily_r "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", info.DeviceId, info.DateFrom, info.DateTo)
	if info.UserId != "" {
		sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, count(1) as num, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily_r "+
			"where user_id='%s' and time>='%s' and time<='%s' group by date", info.UserId, info.DateFrom, info.DateTo)
	}
	fmt.Println(sqlClause)
	datas, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	for _, data := range datas {
		fmt.Println(data)
		var InPage IncomeOfDaily
		InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		InPage.DiskUsage = Str2Float64(data["disk_usage"])
		InPage.NatType = Str2Float64(data["nat_ratio"])
		InPage.JsonDaily = Str2Float64(data["hour_income"])
		InPage.OnlineJsonDaily = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		InPage.Latency = Str2Float64(data["latency"])
		InPages = append(InPages, InPage)
		total += Str2Int64(data["num"])
	}

	return InPages, total, err
}

func GetHourDailyList(info IncomeDailySearch) (list HourDataOfDaily, total int64, err error) {
	// 创建db
	db := GMysqlDb.Model(&HourDataOfDaily{})
	var InPages HourDataOfDaily
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Date == "" {
		log.Error("参数错误")
		return
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.DeviceId != "" {
		db = db.Where("device_id = ?", info.DeviceId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.UserId != "" {
		db = db.Where("user_id = ?", info.UserId)
	}
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Date != "" {
		db = db.Where("date = ?", info.Date)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Find(&InPages).First(&InPages).Error
	return InPages, total, err
}

func customerActFlow(referer, event_id, user_id, pro_id, platform string) bool {
	if event_id == "" {
		log.Error("args error")
		return false
	}
	if user_id == "" {
		log.Error("args error")
		return false
	}
	tableName := "customer_data"
	sqlClause := fmt.Sprintf("select source,channel from %s where id='%s'", tableName, user_id)
	data_cus, erro := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if erro != nil {
		log.Errorf("Getbanner query one data to [%v] failed, data[%v]  error[%v]", tableName, data_cus, erro)
		return false
	}
	if len(data_cus) == 0 {
		log.Debug("信息有误")
		return false
	}
	//如果pro_id不为空代表是申请了贷款
	pro_name := ""
	if pro_id != "" && event_id != "details_page" {
		sqlClause := fmt.Sprintf("select  product_name,logo_path,available_num,loan_date from  loan_product_info where id='%s' ", pro_id)
		data_list_apply, sql_err := com.GetSQLHelper().GetQueryDataList(sqlClause)
		if sql_err != nil {
			log.Errorf(" CustomerFeedback sqlClause[%s] error[%v]", sqlClause, sql_err)
			return false
		}
		if len(data_list_apply) != 0 {
			data_apply := make(map[string]interface{})
			data_apply["product_id"] = data_list_apply[0]["id"]
			data_apply["product_name"] = data_list_apply[0]["product_name"]
			pro_name = data_list_apply[0]["product_name"]
			data_apply["available_num"] = data_list_apply[0]["available_num"]
			data_apply["customer_id"] = user_id
			data_apply["logo_path"] = data_list_apply[0]["logo_path"]
			nowTime := time.Now().Format("2006-01-02 15:04:05")
			data_apply["update_time"] = nowTime
			data_apply["loan_date"] = data_list_apply[0]["loan_date"]
			data_apply["state"] = 2
			data_apply["record_type"] = 3
			tableName := "scan_apply_record"
			data_apply["commit_time"] = nowTime
			_, err := com.GetSQLHelper().InsertDataByMap(tableName, data_apply)
			if err != nil {
				log.Errorf("Save insert one data to [%v] failed, data[%v]  error[%v]", tableName, data_apply, err)
				return false
			}
		}
	}

	data := make(map[string]interface{})
	data["user_id"] = user_id
	data["platform"] = platform
	data["channel"] = data_cus[0]["channel"]
	data["source"] = data_cus[0]["source"]
	data["referer"] = referer
	data["event_id"] = event_id
	data["pro_id"] = pro_id
	data["pro_name"] = pro_name
	data["event_times"] = "1"
	tableName = "act_flow"
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	data["update_time"] = nowTime
	_, err := com.GetSQLHelper().InsertDataByMap(tableName, data)
	if err != nil {
		log.Errorf("Save insert one data to data[%s] failed, data[%v]  error[%v]", tableName, data, err)
		return false
	}
	return true
}

func StartActFlow(w http.ResponseWriter, r *http.Request) {
	event_id := r.FormValue("event_id")
	platform := r.FormValue("platform")
	//channel := r.FormValue("channel")
	source := r.FormValue("source")
	if event_id == "" {
		log.Error("args error")
		HandleError(w, KErrorMsg[KErrorNoArg])
		return
	}

	data := make(map[string]interface{})
	data["event_id"] = event_id
	data["platform"] = platform
	data["source"] = source
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	data["update_time"] = nowTime
	data["event_times"] = "1"
	tableName := "act_flow"
	_, err := com.GetSQLHelper().InsertDataByMap(tableName, data)
	if err != nil {
		log.Errorf("Save insert one data to data[%s] failed, data[%v]  error[%v]", tableName, data, err)
		return
	}
	HandleSuccess(w, "")
}

// 后续有需要,继续在此结构体中增加字段
type ExtendedInfo struct {
	OriginTableName   string
	OriginId          string
	OriginPageUrl     string
	OriginProtocolUrl string
	UserAgent         string
	OriginSource      string
}

// 在item_reuse_history表中保存user_agent字段
// 在item_reuse_history表中保存原始的source字段
func UserIsExist(table, mobile string) string {
	sqlClause := fmt.Sprintf("select id from %s where mobile='%s'order by id desc limit 1", table, mobile)
	rows, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("userIsExist sqlClause[%s] error[%v]", sqlClause, err)
		return ""
	}
	if len(rows) <= 0 {
		return ""
	}
	//differ := GetTimeDiffer(rows[0]["apply_time"], time_now)
	//fmt.Println(differ)
	//if differ >= 5 {
	//	return -1
	//} else {
	//	return -2
	//}
	return rows[0]["id"]
}

func StrToInt(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return -1
	}
	return num
}

func GetTimeDiffer(start_time, end_time string) int64 {
	var minute_time int64
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", start_time, time.Local)
	t2, err := time.ParseInLocation("2006-01-02 15:04:05", end_time, time.Local)
	if err == nil && t1.Before(t2) {
		diff := t2.Unix() - t1.Unix() //
		minute_time = diff / 60
		return minute_time
	} else {
		return minute_time
	}
}

func checkCityLimit(city, channel, cityLevel string) bool {

	//sqlClause := fmt.Sprintf("select * from agent_info where channel='%s'", channel)
	sqlClause := util.GetEscapeSqlClause("select * from agent_info where channel='%s'", channel)
	log.Debug(sqlClause)
	dataList, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err)
		return true
	}

	for _, data := range dataList {
		cities := data[cityLevel]
		if cities == "" || city == "" {
			continue
		}
		tempList := strings.Split(cities, ",")
		for _, tempCity := range tempList {
			if strings.Contains(city, tempCity) || strings.Contains(tempCity, city) {
				return true
			}
		}
	}

	return false
}

func GetSomeInitial(w http.ResponseWriter, r *http.Request) {

	channel := r.FormValue("channel")

	if channel == "" {
		log.Error(KErrorNoArg)
		HandleError(w, KErrorMsg[KErrorNoArg])
		return
	}

	ip := httpUtil.GetRequestIP(r)
	//city := util.GetCityFromIP(ip)
	loc := util.GetIPLocationFromBaiDu(ip)
	city := strings.Split(loc, "_")[1]
	log.Debugf("checkIPCityLimit ip:%s city:%s", ip, city)

	needSwitch := 0

	ok := checkCityLimit(city, channel, "first_citys")
	if ok {
		needSwitch = 1
	}

	outMap := make(map[string]interface{})
	outMap["need_switch"] = needSwitch

	HandleSuccess(w, outMap)
}
