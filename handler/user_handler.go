package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"math/rand"
	"net/http"
	"strings"
	"time"
	log "web-server/alog"
)

const (
	NodeUnknown = iota
	NodeEdge
	NodeCandidate
	NodeScheduler
)

func Register(w http.ResponseWriter, r *http.Request) {
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
