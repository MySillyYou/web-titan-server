package main

import (
	"context"
	"encoding/json"
	gcon "github.com/gorilla/context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"

	//"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"syscall"
	"time"
	log "web-server/alog"
	"web-server/handler"
	mql "web-server/mysql"
)

func loadConfig(path string) (map[string]interface{}, error) {

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("read file error: ", err.Error())
		return nil, err
	}

	var config map[string]interface{}
	if err = json.Unmarshal(file, &config); err != nil {
		log.Error("Unmarshal config error: ", err.Error())
		return nil, err
	}
	var sqlPath handler.GeneralDB
	if err = json.Unmarshal(file, &sqlPath); err != nil {
		return nil, err
	}
	handler.GvaMysql = sqlPath
	return config, nil
}

func httpRoute() {

	handler.RegisterInterface()
	http.HandleFunc("/api/get_miner_info", handler.GetAllMinerInfo)
	http.HandleFunc("/api/get_retrieval", handler.Retrieval)
	http.HandleFunc("/api/get_user_device_info", handler.GetUserDeviceInfo)

	http.HandleFunc("/api/get_index_info", handler.GetIndexInfo)
	http.HandleFunc("/api/get_device_info", handler.GetDevicesInfo)
	http.HandleFunc("/api/get_diagnosis_days", handler.GetDeviceDiagnosisDaily)
	http.HandleFunc("/api/get_diagnosis_hours", handler.GetDeviceDiagnosisHour)

	// 设备绑定
	http.HandleFunc("/api/device_biding", handler.DeviceBiding)
	http.HandleFunc("/api/register", handler.Register)
	http.HandleFunc("/api/device_create", handler.DeviceCreate)
	//http.HandleFunc("/api/save_daily_infos", handler.SaveDailyInfos)

	// 任务
	http.HandleFunc("/api/create_task", handler.CreateTask)
	http.HandleFunc("/api/get_task", handler.GetTaskInfo)
	http.HandleFunc("/api/get_task_list", handler.GetTaskList)
	http.HandleFunc("/api/get_task_detail", handler.GetTaskListDetail)

	//操作流程记录
	http.HandleFunc("/onMobEventStart", handler.StartActFlow)

	//handler.RegisterAuthHandler("/admin/regist_query_inner", handler.QueryInnerRegisterList)
	//点击统计
	//handler.RegisterAuthHandler("/admin/query_total_data", handler.GetTotalData)
	//handler.RegisterAuthHandler("/admin/query_total_data", handler.GetTotalDataa)
	//handler.RegisterAuthHandler("/admin/query_detail_data", handler.GetDetailData)
	//handler.RegisterAuthHandler("/admin/query_inner_data", handler.GetInnerRegistData)
	//handler.RegisterAuthHandler("/admin/query_upload_data", handler.GetUploadData)

	http.Handle("/", http.FileServer(http.Dir("../static")))
	http.Handle("/banner/", http.StripPrefix("/banner/", http.FileServer(http.Dir("../static/banner"))))
	http.Handle("/app_apk/", http.StripPrefix("/app_apk/", http.FileServer(http.Dir("../static/app_apk"))))
	http.Handle("/mipmap-xxxhdpi/", http.StripPrefix("/mipmap-xxxhdpi/", http.FileServer(http.Dir("../static/logo/mipmap-xxxhdpi"))))
	http.Handle("/banner_xxxhdpi/", http.StripPrefix("/banner_xxxhdpi/", http.FileServer(http.Dir("../static/logo/banner_xxxhdpi"))))
	//http.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("../static/dist"))))
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../static"))))
	//http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../static/dist"))))
	//图片路径
	//http.Handle("/banner_slices/mipmap-hdpi/2000-banner.png", http.StripPrefix("/banner_slices/mipmap-hdpi/", http.FileServer(http.Dir("../banner_slices/mipmap-hdpi/"))))
}

var (
	gQuit chan bool
)

func main() {

	rand.Seed(time.Now().UnixNano())

	gQuit = make(chan bool)

	config, err := loadConfig("./conf/config.json")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	handler.GConfig = config
	//初始化MySQL
	mql.GetSQLHelper().Init(config["sql_addr"].(string))
	defer mql.GetSQLHelper().Close()
	handler.NewGormMysqlDB("", "Mysql")
	handler.RpcAddr = config["rpc_addr"].(string)
	Init()
	//handler.Conn, err = jsonrpc.Dial("tcp", rpcAddr)
	//if err != nil {
	//	log.Fatal(err)
	//}

	httpRoute()

	stop()
	srv := startHttpServer()

	<-gQuit
	close(gQuit)

	//最长等待10秒钟，超时将关闭连接
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//关闭http server
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Error(err.Error()) // failure/timeout shutting down the server gracefully
	}

	log.Info("stop server successfully!")
}

func Init() {

	authConfig, err := loadConfig("./conf/auth_menu.json")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	ok := handler.Init(handler.GConfig, authConfig)
	if !ok {
		log.Fatal("handler initial error")
		return
	}
	handler.GDevice = &handler.DeviceTask{}
	handler.GDevice.Initial()
	handler.GWg = &sync.WaitGroup{}
}

func startHttpServer() *http.Server {

	listenPort := handler.GConfig["port"].(string)
	log.Debug("starting server and listening port", listenPort)

	srv := &http.Server{Addr: ":" + listenPort, Handler: gcon.ClearHandler(http.DefaultServeMux)}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Errorf("Httpserver: ListenAndServe() error: %s", err.Error())
		}
	}()
	go handler.Run()
	// returning reference so caller can call Shutdown()
	return srv
}

func stop() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Infof("received ctrl+c(%v)\n", sig)
			handler.GDevice.Done <- struct{}{}
			log.Info("stopping server ...")
			gQuit <- true
			//			close(gDone)
		}
	}()
}
