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
	// 权限验证功能
	handler.RegisterInterface()
	// Titan 管理后台页面路由
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

	// 以下路由功能待定
	// 根据用户请求获取所属地址
	http.HandleFunc("/ts", handler.GetSomeInitials)
	http.Handle("/", http.FileServer(http.Dir("../static")))
	// 静态文件展示
	http.Handle("/banner/", http.StripPrefix("/banner/", http.FileServer(http.Dir("static/banner"))))
	// 提供软件下载
	http.Handle("/app_apk/", http.StripPrefix("/app_apk/", http.FileServer(http.Dir("../static/app_apk"))))
	// 路由定向
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("../static"))))
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
	// 此处可配置权限路由（未开放）
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
	// 监听系统停止信号
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
