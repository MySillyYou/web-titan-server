package handler

import (
	log "web-server/alog"
	mql "web-server/mysql"
	red "web-server/redis"
	"web-server/sms"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"github.com/gorilla/sessions"
)
//const KMaxAge = 86400

var store = sessions.NewCookieStore([]byte("user_info"))

var gConfig map[string]interface{}
var gAuthMenu map[string]interface{}

func Init(config, authMenu map[string]interface{}) bool {

	gConfig = config
	gAuthMenu = authMenu

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(gConfig["max_age"].(float64)),
		HttpOnly: true,
	}

	if !red.Redis.Existed() {
		// 初始化redis
		red.Redis.Init(config["redis_addr"].(string), int(config["redis_db"].(float64)))
	}

	return true
}

func RegisterInterface() {

	//登录模块
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/get_code", GetLoginVerifyCode)
	//RegisterAuthHandler("/modify_pwd", ModifyPwd)
}

//func FlashTTL(r *http.Request, w http.ResponseWriter) error {
//
//	session, err := store.Get(r, "user_info")
//	if err != nil {
//		return err
//	}
//	// 删除session，并重新生成
//	store.Clear(r)
//	newSession, err := store.Get(r, "user_info")
//	if err != nil {
//		return err
//	}
//	newSession.Values = session.Values
//	store.Save(r, w, newSession)
//	return nil
//}

//func CheckUserHandler(h http.Handler) http.Handler {
//	return http.HandlerFunc(
//		func(w http.ResponseWriter, r *http.Request) {
//
//			defer func() {
//				if e, ok := recover().(error); ok {
//					HandleError(w, "服务异常")
//					log.Error(e.Error())
//				}
//			}()
//
//			r.FormValue("")
//			path := r.URL.Path
//			if gConfig["is_login"].(bool) {
//				r.Form["operator"] = []string{"test"}
//				h.ServeHTTP(w, r)
//				return
//			}
//
//			session, err := store.Get(r, "user_info")
//			if err != nil {
//				log.Error("user_info not exist")
//				HandleAuthError(w)
//				return
//			}
//
//			session2FormValue := func(ssKey string) bool {
//				var (
//					ok      bool
//					ssValue interface{}
//				)
//				if ssValue, ok = session.Values[ssKey]; !ok {
//					log.Errorf("%s not exist", ssKey)
//					return false
//				}
//				r.Form[ssKey] = []string{ssValue.(string)}
//				return true
//			}
//
//			if !session2FormValue("operator") {
//				HandleAuthError(w)
//				return
//			}
//			if !session2FormValue("role") {
//				HandleAuthError(w)
//				return
//			}
//			if !session2FormValue("channels") {
//				HandleAuthError(w)
//				return
//			}
//			if !session2FormValue("menus") {
//				HandleAuthError(w)
//				return
//			}
//
//			// 刷新过期时间
//			FlashTTL(r, w)
//
//			//			log.Debug("role", r.Form["role"], r.Form["operator"])
//
//			pathAuth := checkPathAuth(r.FormValue("menus"), path)
//			if !pathAuth {
//				HandleAuthError(w)
//				return
//			}
//
//			h.ServeHTTP(w, r)
//		})
//}

/*
	检测path路径的权限
*/
func checkPathAuth(menus string, path string) bool {
	tempAuth, ok := gAuthMenu["auth_menus"].(map[string]interface{})
	if !ok {
		log.Error("gAuthMenu auth_menus format error")
		return false
	}

	pathAuth, ok := tempAuth[path].(string)
	if !ok {
		log.Errorf("gAuthMenu auth_menus %s format error", path)
		return false
	}

	authMenuList := strings.Split(menus, ",")
	pathMenuList := strings.Split(pathAuth, ",")

	for _, authMenu := range authMenuList {
		for _, pathMenu := range pathMenuList {
			if authMenu == pathMenu {
				return true
			}
		}
	}

	return false
}

func getAllMenus() string {

	tempAuth, ok := gAuthMenu["fields"].(map[string]interface{})
	if !ok {
		return ""
	}

	tempList := make([]string, 0, len(tempAuth))
	for key := range tempAuth {
		tempList = append(tempList, key)
	}

	return strings.Join(tempList, ",")
}

func getRoleMenus(role string) string {

	tempAuth, ok := gAuthMenu["role_auth"].(map[string]interface{})
	if !ok {
		return ""
	}

	auth, ok := tempAuth[role].(string)
	if !ok {
		return ""
	}

	return auth
}

// 登录
func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	//password := r.FormValue("password")
	verifyCode := r.FormValue("code")

	if username == "" {
		log.Error("need username:", username)
		HandleError(w, "参数错误")
		return
	}

	log.Debug("Login:", username, verifyCode)

	redisKey := gConfig["redis_prefix"].(string) + username

	if gConfig["sms"].(bool) {
		// 首先核对验证码
		realCode, err := red.Redis.GetString(redisKey)
		if err != nil {
			log.Error("Login check verify code failed, error:", err)
			HandleError(w, "验证码错误")
			return
		}
		if verifyCode != realCode {
			log.Errorf("Login username[%s] user code[%s] the real code[%s] is different", username, verifyCode, realCode)
			HandleError(w, "验证码输入错误")
			return
		}
	}
	//sqlClause := fmt.Sprintf("select mobile,role,display_name,status from operators where operator = '%s'", username)
	sqlClause := fmt.Sprintf("select mobile,role,display_name,menus,channels,status from operators where operator=?")
	mapRet, err := mql.GetSQLHelper().GetQueryDataList(sqlClause, username)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "用户名错误")
		return
	}

	if len(mapRet) <= 0 {
		log.Error("len(mapRet) <= 0")
		HandleError(w, "用户名错误")
		return
	}

	if mapRet[0]["status"] != "Y" {
		log.Errorf("Login failed, username[%s] 账户已禁用", username)
		HandleError(w, "账户已禁用，请联系管理员")
		return
	}

	session, err := store.Get(r, "user_info")
	if err != nil {
		log.Error(err.Error())
		HandleError(w, err.Error())
		return
	}

	role := mapRet[0]["role"]
	menus := mapRet[0]["menus"]
	if role == KRoleAdmin && menus == "" {
		menus = getAllMenus()
	}

	session.Values["operator"] = username
	session.Values["role"] = role
	session.Values["channels"] = mapRet[0]["channels"]
	session.Values["menus"] = menus
	err = store.Save(r, w, session)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, err.Error())
		return
	}
	cookie := &http.Cookie{}
	cookie.Name = "SESSIONID"
	cookie.Value = username
	cookie.MaxAge = int(gConfig["max_age"].(float64))
	http.SetCookie(w, cookie)

	outMap := make(map[string]interface{})
	outMap["role"] = role
	outMap["display_name"] = mapRet[0]["display_name"]
	outMap["menus"] = menus

	// 记录最后登录时间
	sqlClause = fmt.Sprintf("update operators set login_time=now() where operator='%s'", username)
	_, err =mql.GetSQLHelper().ExecSqlClause(sqlClause)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, err.Error())
		return
	}

	HandleSuccess(w, outMap)

	// 登录成功，删除本次使用的验证码
	red.Redis.DelKey(redisKey)

	opMap := make(map[string]interface{})
	opMap["operator"] = username

}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user_info")
	// var operator interface{}
	// var ok bool
	// if operator, ok = session.Values["operator"]; !ok {
	// 	HandleError(w, "")
	// 	return
	// }
	if err != nil {
		log.Error(err.Error())
		HandleError(w, err.Error())
		return
	}
	for key := range session.Values {
		delete(session.Values, key)
	}
	store.Save(r, w, session)

	HandleSuccess(w, "")
}

// 修改密码
func ModifyPwd(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	oldPWD := r.FormValue("old_pwd")
	newPWD := r.FormValue("new_pwd")
	confirmPWD := r.FormValue("confirm_pwd")

	if username == "" || oldPWD == "" || newPWD == "" || confirmPWD == "" {
		log.Error("arg eror username:", username, "old_pwd:", oldPWD, "new_pwd:", newPWD, "confirm_pwd:", confirmPWD)
		HandleError(w, "修改失败，参数错误")
		return
	}

	if newPWD != confirmPWD {
		log.Error("new_pwd != confirm_pwd", "new_pwd:", newPWD, "confirm_pwd:", confirmPWD)
		HandleError(w, "修改失败，新密码两次不一致")
		return
	}

	sqlClause := fmt.Sprintf("select mobile, role from operators where operator = '%s' and password = '%s'", username, oldPWD)
	mapRet, err := mql.GetSQLHelper().GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "修改失败，用户名或密码错误")
		return
	}

	if len(mapRet) <= 0 {
		log.Error("len(mapRet) <= 0")
		HandleError(w, "修改失败，原密码不正确")
		return
	}

	sqlClause = fmt.Sprintf("update operators set password='%s' where operator = '%s' and password = '%s'", newPWD, username, oldPWD)
	_, err = mql.GetSQLHelper().ExecSqlClause(sqlClause)
	if err != nil {
		log.Error(err.Error())
		HandleError(w, "修改失败，服务异常")
		return
	}

	HandleSuccess(w, "修改成功")

}

//func RegisterAuthHandler(pattern string, f func(http.ResponseWriter, *http.Request)) {
//	http.Handle(pattern, CheckUserHandler(http.HandlerFunc(f)))
//}

// 获取登录验证码
func GetLoginVerifyCode(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		log.Error("GetLoginVerifyCode username is null")
		HandleError(w, "账号不能为空")
		return
	}
	//sqlClause := fmt.Sprintf("select mobile,status from operators where operator='%s'", username)
	sqlClause := fmt.Sprintf("select mobile,status from operators where operator=?")
	rows, err := mql.GetSQLHelper().GetQueryDataList(sqlClause, username)
	if err != nil || len(rows) < 1 || rows[0]["mobile"] == "" {
		log.Errorf("GetLoginVerifyCode get mobile failed, username[%s] error[%v]", username, err)
		HandleError(w, "手机号码未注册")
		return
	}

	if rows[0]["status"] != "Y" {
		log.Errorf("GetLoginVerifyCode get mobile failed, username[%s] 账户已禁用", username)
		HandleError(w, "账户已禁用，请联系管理员")
		return
	}

	expireTime := int64(300) //默认5分钟
	redisKey := gConfig["redis_prefix"].(string) + username

	ttl, err := red.Redis.GetTTL(redisKey)
	if err == nil && (expireTime-ttl) < 60 { //发送间隔小于1分钟
		HandleError(w, "验证码已发送，请耐心等待")
		return
	}

	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	verifyCode := fmt.Sprintf("%06d", rander.Intn(1000000))
	log.Debug("verifyCode", verifyCode)
	err = red.Redis.SetString(redisKey, verifyCode) // 随机的验证码写入redis缓存
	if err != nil {
		log.Errorf("GetLoginVerifyCode save verify code failed, error[%v]", err)
		HandleError(w, "获取验证码失败")
		return
	}
	red.Redis.SetTTL(redisKey, expireTime) // 设置验证码过期时间
	if gConfig["sms"].(bool) {
		//err = sms.SendSMSMessage(fmt.Sprintf("您的登录验证码是%s, 5分钟内有效【因地利】", verifyCode), rows[0]["mobile"])
		err = sms.UsingAlidayuSendSMS("24692602", rows[0]["mobile"], "因地利", "SMS_90180036", verifyCode, "c279b2fea52239946ec15b4ac4659c3b")
	} else {
		err = nil // for debug
	}
	if err != nil {
		log.Errorf("GetLoginVerifyCode send sms message failed, error[%v]", err)
		HandleError(w, "获取验证码失败")
		return
	}

	HandleSuccess(w, "")
}

//检测验证码
func CheckVerifyCode(mobile, code string) bool {

	redisKey := "verify_Code" + mobile

	// 核对验证码
	realCode, err := red.Redis.GetString(redisKey)
	if err != nil {
		log.Error("check verify code failed, error:", err.Error())
		return false
	}
	if code != realCode {
		log.Errorf("CheckVerifyCode mobile[%s] user code[%s] the real code[%s] is different", mobile, code, realCode)
		return false
	}
	return true
}
