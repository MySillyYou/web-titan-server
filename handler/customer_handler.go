package handler

import (
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
	p.DeviceId = r.FormValue("device_id")
	var res IncomeDailyRes
	m := timeFormat(p)
	res.DailyIncome = m
	res.DeviceDiagnosis = "优秀"
	HandleSuccess(w, res)
}

func GetDeviceDiagnosisHour(w http.ResponseWriter, r *http.Request) {
	var p IncomeDailySearch
	p.DeviceId = r.FormValue("device_id")
	p.Date = r.FormValue("date")
	var res IncomeDailyRes
	m := timeFormatHour(p)
	res.DailyIncome = m
	res.DeviceDiagnosis = "优秀"
	HandleSuccess(w, res)
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

func GetSomeInitials(w http.ResponseWriter, r *http.Request) {

	ip := httpUtil.GetRequestIP(r)
	//city := util.GetCityFromIP(ip)
	loc := util.GetIPLocationFromBaiDu(ip)
	city := strings.Split(loc, "_")[1]
	log.Debugf("checkIPCityLimit ip:%s city:%s", ip, city)
	fmt.Println(loc)

	HandleSuccess(w, "")
}
