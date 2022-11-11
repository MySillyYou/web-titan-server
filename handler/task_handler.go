package handler

import (
	"fmt"
	"net/http"
	"strconv"
	log "web-server/alog"
	mql "web-server/mysql"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
	var TaskInfoOld TaskInfo
	var TaskInfos TaskInfo
	TaskInfos.UserId = r.FormValue("userId")
	TaskInfos.Cid = r.FormValue("cid")
	TaskInfos.BandwidthUp = Str2Float64(r.FormValue("bandwidth_up"))
	TaskInfos.BandwidthDown = Str2Float64(r.FormValue("bandwidth_down"))
	TaskInfos.Status = "新建"
	TaskInfos.TimeNeed = r.FormValue("time_need")
	Price := r.FormValue("price")
	if TaskInfos.Cid == "" || TaskInfos.TimeNeed == "" || Price == "" {
		HandleError(w, "缺少参数")
		return
	}
	TaskInfos.Price = StrToFloat(Price)
	result := GMysqlDb.Where("cid = ?", TaskInfos.Cid).First(&TaskInfoOld)
	if result.RowsAffected <= 0 {
		err := GMysqlDb.Create(&TaskInfos).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}
		HandleSuccess(w, "创建成功")
	} else {
		TaskInfoOld.BandwidthUp = TaskInfos.BandwidthUp
		TaskInfoOld.BandwidthDown = TaskInfos.BandwidthDown
		TaskInfoOld.TimeNeed = TaskInfos.TimeNeed
		TaskInfoOld.Price = TaskInfos.Price
		err := GMysqlDb.Save(&TaskInfoOld).Error
		if err != nil {
			log.Error(err.Error())
			HandleError(w, "参数错误")
			return
		}
		HandleSuccess(w, "修改成功")
	}

}

type TaskInfoRes struct {
	List  []TaskInfo `json:"list"`
	Count int64      `json:"count"`
}

func GetTaskInfo(w http.ResponseWriter, r *http.Request) {
	var res TaskInfoRes
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.UserId = r.FormValue("userId")
	TaskInfoSearch.Status = r.FormValue("status")
	if TaskInfoSearch.Status == "All" {
		TaskInfoSearch.Status = ""
	}
	if TaskInfoSearch.UserId == "" {
		HandleError(w, "缺少参数")
		return
	}
	list, total, err := GetTaskInfoList(TaskInfoSearch)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "参数错误")
		return
	}
	res.List = list
	res.Count = total
	HandleSuccess(w, res)
}

func GetTaskList(w http.ResponseWriter, r *http.Request) {
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.DeviceId = r.FormValue("device_id")
	TaskInfoSearch.Status = r.FormValue("status")
	//page := Str2Int(r.FormValue("page"))
	//pageRow := Str2Int(r.FormValue("page_row"))
	if TaskInfoSearch.DeviceId == "" {
		HandleError(w, "缺少参数")
		return
	}
	sqlClause := fmt.Sprintf("select date_format(created_at, '%%Y-%%m-%%d') as date, count(1) as num, sum(file_size) as file_size,sum(price) as price from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') group by date", TaskInfoSearch.DeviceId)
	fmt.Println(sqlClause)
	datas, err := mql.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		HandleCodeMsg(w, KErrorServer, KErrorMsg[KErrorServer])
		return
	}
	sqlClause = fmt.Sprintf("select count(1) as num_all, sum(file_size) as file_size_all from task_info "+
		"where device_id='%s' and status in ('已完成','已连接')", TaskInfoSearch.DeviceId)
	fmt.Println(sqlClause)
	count_all, err := mql.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		HandleCodeMsg(w, KErrorServer, KErrorMsg[KErrorServer])
		return
	}
	resp := make(map[string]interface{})
	resp["tot_num"] = count_all
	resp["data_list"] = datas
	fmt.Println(datas)
	HandleSuccess(w, resp)
}

func GetTaskListDetail(w http.ResponseWriter, r *http.Request) {
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.DeviceId = r.FormValue("device_id")
	date := r.FormValue("date")
	//page := Str2Int(r.FormValue("page"))
	//pageRow := Str2Int(r.FormValue("page_row"))
	if TaskInfoSearch.DeviceId == "" {
		HandleError(w, "缺少参数")
		return
	}
	beginTime := ""
	endTime := ""
	if len(date) == 10 {
		beginTime = date + " 00:00:00"
		endTime = date + " 23:59:59"
	}
	sqlClause := fmt.Sprintf("select date_format(created_at, '%%Y-%%m-%%d') as date,cid,file_name,file_size,bandwidth_up,bandwidth_down,ip_address,created_at,status from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') and created_at>='%s' and  created_at<='%s'", TaskInfoSearch.DeviceId, beginTime, endTime)
	fmt.Println(sqlClause)
	datas, err := mql.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		HandleCodeMsg(w, KErrorServer, KErrorMsg[KErrorServer])
		return
	}
	sqlClause = fmt.Sprintf("select count(1) as num from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') and created_at>='%s' and  created_at<='%s'", TaskInfoSearch.DeviceId, beginTime, endTime)
	fmt.Println(sqlClause)
	count_all, err := mql.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		HandleCodeMsg(w, KErrorServer, KErrorMsg[KErrorServer])
		return
	}
	resp := make(map[string]interface{})
	resp["tot_num"] = count_all
	resp["data_list"] = datas
	fmt.Println(datas)
	HandleSuccess(w, resp)
}

type TaskSearch struct {
	TaskInfo
	PageInfo
}

// DevicesInfo search from mysql
func GetTaskInfoList(info TaskSearch) (list []TaskInfo, total int64, err error) {
	// string转成int：
	limit, _ := strconv.Atoi(info.PageSize)
	page, _ := strconv.Atoi(info.Page)
	offset := limit * (page - 1)
	// 创建db
	dbTask := GMysqlDb.Model(&TaskInfo{})
	var InPages []TaskInfo
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Status != "" && info.Status != "All" {
		dbTask = dbTask.Where("status = ?", info.Status)
	}
	if info.Cid != "" {
		dbTask = dbTask.Where("cid = ?", info.Cid)
	}
	if info.UserId != "" {
		dbTask = dbTask.Where("user_id = ?", info.UserId)
	}
	err = dbTask.Count(&total).Error
	if err != nil {
		return
	}
	err = dbTask.Limit(limit).Offset(offset).Find(&InPages).Error
	return InPages, total, err
}
