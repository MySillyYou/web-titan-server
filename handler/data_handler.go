package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/linguohua/titan/api/client"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	log "web-server/alog"
	com "web-server/mysql"
)

var AllM AllMinerInfo

// rpc demo
func ClientRun() {
	url := "http://192.168.0.177:3456/rpc/v0"
	ctx := context.Background()
	apiScheduler, closer, err := client.NewScheduler(ctx, url, nil)

	defer closer()
	v, err := apiScheduler.FindNodeWithBlock(ctx, "QmeUqw4FY1wqnh2FMvuc2v8KAapE7fYwu2Up4qNwhZiRk7", "119.28.56.202")
	if err != nil {
		print("sss")
	}
	a := fmt.Sprintf("%s", v)
	fmt.Println(a)
}
func (t *DeviceTask) DeviceInfoGetFromRpc(DeviceId string) (DeviceInfo DevicesInfo, err error) {
	var data RpcDevice
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.GetDevicesInfo"
	song["id"] = 3
	song["params"] = []string{DeviceId}
	//song["params"] = []string{"e_ff8f7733511411ed889b86f108eb4c09"}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", RpcAddr, reader)
	if err != nil {
		log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	//defer client.CloseIdleConnections()
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if GUpdate {
		var daily IncomeDaily
		daily.Time = GTime
		daily.DiskUsage = data.Result.DiskUsage
		daily.DeviceId = data.Result.DeviceId
		daily.PkgLossRatio = data.Result.PkgLossRatio
		daily.JsonDaily = data.Result.TodayProfit
		data.Result.TodayOnlineTime = data.Result.OnlineTime
		daily.OnlineJsonDaily = data.Result.TodayOnlineTime
		daily.Latency = data.Result.Latency
		daily.DiskUsage = data.Result.DiskUsage
		_, ok := t.DeviceIdAndUserId[daily.DeviceId]
		if ok {
			daily.UserId = t.DeviceIdAndUserId[daily.DeviceId]
		}
		err = TransferData(daily)
		if err != nil {
			log.Error(err.Error())
		}
	}
	return data.Result, nil
}

func CidInfoGetFromRpc(DeviceId string) error {
	var data RpcTask
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.GetDownloadInfo"
	song["id"] = 3
	song["params"] = []string{DeviceId}
	//song["params"] = []string{"e_ff8f7733511411ed889b86f108eb4c09"}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bytesData)
	//request, err := http.NewRequest("POST", "http://192.168.0.73:3456/rpc/v0", reader)
	request, err := http.NewRequest("POST", RpcAddr, reader)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	//defer client.CloseIdleConnections()
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	var dataSave TaskInfo
	if len(data.Result) > 0 {
		for _, taskOne := range data.Result {
			dataSave.Cid = taskOne.Cid
			dataSave.DeviceId = taskOne.DeviceId
			dataSave.FileSize = taskOne.FileSize
			dataSave.Price = taskOne.Reward
			dataSave.BandwidthUp = taskOne.BandwidthUp
			err = SaveTaskInfo(dataSave)
			if err != nil {
				log.Error(err.Error())
				continue
			}
		}
	}
	return nil
}

func AllMinerInfoGetFromRpc() {
	var data AllMinerInfo
	song := make(map[string]interface{})
	song["jsonrpc"] = "2.0"
	song["method"] = "titan.StateNetwork"
	song["id"] = 3
	song["params"] = []string{}
	//song["params"] = []string{"e_ff8f7733511411ed889b86f108eb4c09"}
	bytesData, err := json.Marshal(song)
	if err != nil {
		return
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", RpcAddr, reader)
	if err != nil {
		log.Error(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	//defer client.CloseIdleConnections()
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err.Error())
		return
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}
	DeviceMap := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &DeviceMap)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Error(err.Error())
		return
	}
	AllM = data
	return
}

// device data save

//  daily data save

func SetDailyInfo(jsonStr, date, income string) string {
	onlineJsonDailyFrom := make(map[string]interface{})
	if jsonStr != "" {
		e := json.Unmarshal([]byte(jsonStr), &onlineJsonDailyFrom)
		if e != nil {
			return ""
		}
	}
	if income == "" {
		income = "0"
	}
	onlineJsonDailyFrom[date] = Str2Float64(income)
	bytes, e := json.Marshal(onlineJsonDailyFrom)
	if e != nil {
		return ""
	}
	jsonString := string(bytes)
	return jsonString
}

func (t *DeviceTask) SaveDeviceInfo(Df string) error {
	data, err := t.DeviceInfoGetFromRpc(Df)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if data.DeviceId == "" {
		log.Println("接受空数据", Df)
		return nil
	}

	var DeviceInfoOld DevicesInfo
	// 统计数据以web服务器为准，将接收的统计清零
	data.SevenDaysProfit = 0
	data.MonthProfit = 0
	data.YesterdayIncome = 0
	result := GMysqlDb.Where("device_id = ?", data.DeviceId).First(&DeviceInfoOld)
	if result.RowsAffected <= 0 {
		data.CreatedAt = time.Now()
		err := GMysqlDb.Create(&data).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	} else {
		data.ID = DeviceInfoOld.ID
		//data.UserId = DeviceInfoOld.UserId
		//data.CreatedAt = DeviceInfoOld.CreatedAt
		data.UpdatedAt = time.Now()
		//err := GMysqlDb.Save(&data).Error
		err := GMysqlDb.Updates(&data).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil

}

func SaveTaskInfo(data TaskInfo) error {
	//data, err := CidInfoGetFromRpc("e_ff8f7733511411ed889b86f108eb4c09")
	//if err != nil {
	//	log.Error(err.Error())
	//	return err
	//}
	if data.DeviceId == "" {
		log.Println("接受空数据", data)
		return nil
	}

	var DeviceInfoOld TaskInfo
	data.TimeDone = time.Now()
	result := GMysqlDb.Where("device_id = ?", data.DeviceId).Where("cid = ?", data.Cid).First(&DeviceInfoOld)
	if result.RowsAffected <= 0 {
		data.CreatedAt = time.Now()
		err := GMysqlDb.Create(&data).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	} else {
		data.ID = DeviceInfoOld.ID
		//data.UserId = DeviceInfoOld.UserId
		//data.CreatedAt = DeviceInfoOld.CreatedAt
		data.UpdatedAt = time.Now()
		//err := GMysqlDb.Save(&data).Error
		err := GMysqlDb.Updates(&data).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil

}

func TransferData(daily IncomeDaily) error {
	if daily.DeviceId == "" {
		log.Println("接受空数据")
		return nil
	}
	var dailyOld IncomeDaily
	daily.UpdatedAt = time.Now()
	result := GMysqlDb.Where("device_id = ?", daily.DeviceId).Where("time = ?", daily.Time).First(&dailyOld)
	if result.RowsAffected <= 0 {
		daily.CreatedAt = time.Now()
		err := GMysqlDb.Create(&daily).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	} else {
		daily.ID = dailyOld.ID
		//data.UserId = DeviceInfoOld.UserId

		//err := GMysqlDb.Save(&data).Error
		err := GMysqlDb.Updates(&daily).Error
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil

}

func (t *DeviceTask) SavaIncomeDailyInfo(daily IncomeOfDaily) {
	if daily.DeviceId == "" {
		log.Println("每日数据空")
		return
	}
	var dailyOld IncomeOfDaily
	daily.UpdatedAt = time.Now()
	_, ok := t.DeviceIdAndUserId[daily.DeviceId]
	if ok {
		daily.UserId = t.DeviceIdAndUserId[daily.DeviceId]
	}
	result := GMysqlDb.Where("device_id = ?", daily.DeviceId).Where("time = ?", daily.Time).First(&dailyOld)
	if result.RowsAffected <= 0 {
		daily.CreatedAt = time.Now()
		err := GMysqlDb.Create(&daily).Error
		if err != nil {
			log.Error(err.Error())
			return
		}
	} else {
		daily.ID = dailyOld.ID
		err := GMysqlDb.Updates(&daily).Error
		if err != nil {
			log.Error(err.Error())
			return
		}
	}
	return

}

func (t *DeviceTask) FormatIncomeDailyList(deviceId string) {
	timeNow := time.Now().Format("2006-01-02")
	DateFrom := timeNow + " 00:00:00"
	DateTo := timeNow + " 23:59:59"
	sqlClause := fmt.Sprintf("select user_id,date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily_r "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", deviceId, DateFrom, DateTo)
	datas, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, data := range datas {
		var InPage IncomeOfDaily
		InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		InPage.DiskUsage = Str2Float64(data["disk_usage"])
		InPage.NatType = Str2Float64(data["nat_ratio"])
		InPage.JsonDaily = Str2Float64(data["hour_income"])
		InPage.OnlineJsonDaily = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		InPage.Latency = Str2Float64(data["latency"])
		InPage.DeviceId = deviceId
		InPage.UserId = data["user_id"]
		t.SavaIncomeDailyInfo(InPage)
	}
	return
}

func (t *DeviceTask) CountDataByUser(userId string) {
	dd, _ := time.ParseDuration("-24h")
	timeBase := time.Now().Add(dd * 1).Format("2006-01-02")
	DateFrom := timeBase + " 00:00:00"
	DateTo := timeBase + " 23:59:59"
	sqlClause := fmt.Sprintf("select user_id, sum(income) as income from income_daily_r "+
		"where  time>='%s' and time<='%s' and user_id='%s' group by user_id;", DateFrom, DateTo, userId)
	datas, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, data := range datas {
		var InPage IncomeOfDaily
		InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		InPage.DiskUsage = Str2Float64(data["disk_usage"])
		InPage.NatType = Str2Float64(data["nat_ratio"])
		InPage.JsonDaily = Str2Float64(data["hour_income"])
		InPage.OnlineJsonDaily = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		InPage.Latency = Str2Float64(data["latency"])
		InPage.UserId = data["user_id"]
		t.SavaIncomeDailyInfo(InPage)
	}
	return
}

func (t *DeviceTask) UpdateYesTodayIncome(DeviceId string) {
	dd, _ := time.ParseDuration("-24h")
	timeBase := time.Now().Add(dd * 1).Format("2006-01-02")
	timeNow := time.Now().Format("2006-01-02")
	DateFrom := timeBase + " 00:00:00"
	DateTo := timeBase + " 23:59:59"
	dataY := QueryDataByDate(DeviceId, DateFrom, DateTo)
	timeBase = time.Now().Add(dd * 6).Format("2006-01-02")
	DateFrom = timeBase + " 00:00:00"
	DateTo = timeNow + " 23:59:59"
	dataS := QueryDataByDate(DeviceId, DateFrom, DateTo)
	timeBase = time.Now().Add(dd * 29).Format("2006-01-02")
	DateFrom = timeBase + " 00:00:00"
	dataM := QueryDataByDate(DeviceId, DateFrom, DateTo)
	dataA := QueryDataByDate(DeviceId, "", "")
	DateFrom = timeNow + " 00:00:00"
	DateTo = timeNow + " 23:59:59"
	dataT := QueryDataByDate(DeviceId, DateFrom, DateTo)
	var dataUpdate DevicesInfo
	dataUpdate.YesterdayIncome = 0
	dataUpdate.SevenDaysProfit = 0
	dataUpdate.MonthProfit = 0
	dataUpdate.CumuProfit = 0
	dataUpdate.TodayOnlineTime = 0
	dataUpdate.TodayProfit = 0
	if len(dataY) > 0 {
		dataUpdate.YesterdayIncome = Str2Float64(dataY["income"])
	}
	if len(dataS) > 0 {
		dataUpdate.SevenDaysProfit = Str2Float64(dataS["income"])
	}
	if len(dataM) > 0 {
		dataUpdate.MonthProfit = Str2Float64(dataM["income"])
	}
	if len(dataA) > 0 {
		dataUpdate.CumuProfit = Str2Float64(dataA["income"])
	}
	if len(dataT) > 0 {
		dataUpdate.TodayProfit = Str2Float64(dataT["income"])
		dataUpdate.TodayOnlineTime = Str2Float64(dataT["online_time"])
	}
	dataUpdate.UpdatedAt = time.Now()
	_, ok := t.DeviceIdAndUserId[DeviceId]
	if ok {
		dataUpdate.UserId = t.DeviceIdAndUserId[DeviceId]
	}
	//err := GMysqlDb.Save(&data).Error
	var dataOld DevicesInfo
	result := GMysqlDb.Where("device_id = ?", DeviceId).First(&dataOld)
	if result.RowsAffected <= 0 {
		dataUpdate.CreatedAt = time.Now()
		err := GMysqlDb.Create(&dataUpdate).Error
		if err != nil {
			log.Error(err.Error())
			return
		}
	} else {
		dataOld.YesterdayIncome = dataUpdate.YesterdayIncome
		dataOld.SevenDaysProfit = dataUpdate.SevenDaysProfit
		dataOld.MonthProfit = dataUpdate.MonthProfit
		dataOld.CumuProfit = dataUpdate.CumuProfit
		dataOld.UpdatedAt = dataUpdate.UpdatedAt
		dataOld.TodayOnlineTime = dataUpdate.TodayOnlineTime
		dataOld.TodayProfit = dataUpdate.TodayProfit
		if dataUpdate.UserId != "" {
			dataOld.UserId = dataUpdate.UserId
		}
		err := GMysqlDb.Save(&dataOld).Error
		if err != nil {
			log.Error(err.Error())
			return
		}
	}
	return
}

func QueryDataByDate(DeviceId, DateFrom, DateTo string) map[string]string {

	sqlClause := fmt.Sprintf("select sum(income) as income,online_time from income_daily_r "+
		"where  time>='%s' and time<='%s' and device_id='%s' group by user_id;", DateFrom, DateTo, DeviceId)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select sum(income) as income,online_time from income_daily_r "+
			"where device_id='%s' group by user_id;", DeviceId)
	}
	//fmt.Println(sqlClause)
	data, err := com.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	if len(data) > 0 {
		return data[0]
	}
	return nil
}

var (
	GDevice       *DeviceTask
	GWg           *sync.WaitGroup
	GUpdateTagNew string
	GUpdate       bool
	GUpdateTask   bool
	GTime         time.Time
)

type DeviceTask struct {
	Done              chan struct{}
	RunInterval       int64
	DeviceIds         []string
	DeviceIdAndUserId map[string]string
}

func (t *DeviceTask) Initial() {
	t.Done = make(chan struct{}, 1)
	t.RunInterval = int64(gConfig["run_interval"].(float64))
	t.DeviceIdAndUserId = make(map[string]string)
	t.getDeviceIds()
	//data, _ := CidInfoGetFromRpc(t.DeviceIds[0])
	//fmt.Println(data)
	today := time.Now().Format(TimeFormatYMD)
	GUpdateTagNew = today
	GUpdate = false
	GUpdateTask = false
	//t.DeviceIds = append(t.DeviceIds, "e_9453993051df11edb9d0902e1671f843")
}

func (t *DeviceTask) getDeviceIds() {
	var info DevicesSearch
	list, _, err := GetDevicesInfoList(info)
	if err != nil {
		log.Error("args error")
		return
	}
	for _, deviceId := range list {
		t.DeviceIds = append(t.DeviceIds, deviceId.DeviceId)
		if deviceId.UserId != "" && deviceId.DeviceId != "" {
			t.DeviceIdAndUserId[deviceId.DeviceId] = deviceId.UserId
		}
	}

	//t.DeviceIds = append(t.DeviceIds, "e_9453993051df11edb9d0902e1671f843")
	return
}

func (t *DeviceTask) itemRun() {
	defer GWg.Done()
	ticker := time.Tick(time.Duration(t.RunInterval) * time.Second)
	for {
		select {
		case <-t.Done:
			log.Infof("device Run once loop end")
			return
		default:
		}

		//nowHour := time.Now().Hour()
		nowMin := time.Now().Minute()
		if nowMin == 00 || nowMin == 10 || nowMin == 20 || nowMin == 30 || nowMin == 40 || nowMin == 45 {
			GTime = time.Now()
			GUpdate = true
		}
		//today := time.Now().Format(TimeFormatYMD)
		//if GUpdateTagNew == "" || GUpdateTagNew != today {
		//	GUpdate = true
		//	GUpdateTagNew = today
		//}
		for _, deviceId := range t.DeviceIds {
			err := t.SaveDeviceInfo(deviceId)
			if err != nil {
				log.Infof("wrong msg %v", err)
				<-ticker
				continue
			}
			if GUpdate {
				// 定时任务更新每日设备参数信息
				t.FormatIncomeDailyList(deviceId)
				// 定时任务更新统计收入信息
				t.UpdateYesTodayIncome(deviceId)
				// 定时更新全网数据
				AllMinerInfoGetFromRpc()
				// 更新设备完成任务
				err := CidInfoGetFromRpc(deviceId)
				if err != nil {
					log.Infof("wrong msg %v", err)
					<-ticker
					continue
				}
			}

		}
		GUpdate = false

		//log.Debug("device Run once loop end")
		<-ticker
	}
}

func Run() {
	GWg.Add(1)
	go GDevice.itemRun()
	GWg.Wait()
	log.Debug("run loop end")
}
