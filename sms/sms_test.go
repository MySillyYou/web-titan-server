package sms

import (
	"fmt"
	"testing"
	log "web-server/alog"
)

func Test_SMS(t *testing.T) {
	fmt.Println("result:", SendMessageSIOO("17620163567", `您的贷款申请已经收到，我们的工作人员会在24小时内与您联系核实资料，回复T退订。`))
	//fmt.Println("result:", SendMessageSIOO("18718526381", `您的贷款申请已经收到，我们的工作人员会在24小时内与您联系核实资料，回复T退订。`))

	return

	// phone list
	phoneList := make([]string, 0)
	phoneList = append(phoneList, "18718526381")
	/*
		phoneList = append(phoneList, "18666076776")
		phoneList = append(phoneList, "18680459353")
		phoneList = append(phoneList, "18825040603")
		phoneList = append(phoneList, "18825161126")
		phoneList = append(phoneList, "13794342581")
		phoneList = append(phoneList, "15121198474")
		phoneList = append(phoneList, "13570506413")
		phoneList = append(phoneList, "18924109315")
		phoneList = append(phoneList, "15814522320")
		phoneList = append(phoneList, "15920137576")
		phoneList = append(phoneList, "13450905116")
		phoneList = append(phoneList, "13544170226")
		phoneList = append(phoneList, "15112475883")
		phoneList = append(phoneList, "13189030495")
	*/

	// smsContent := "【好信用服务】祝大家天天开心，有钱花，有肉吃"

	for _, phone := range phoneList {
		if err := UsingAlidayuSendSMS("23572784", phone, "好幸用", "SMS_68630002", "", "e277d6b54bdf83a6a37d95d4d0ec6ca4"); err != nil {
			// send msg from 明传无线
			// if err := SendSMSMessage(smsContent, phone); err != nil {
			log.Errorf("send msg to %s error. error_info: %s", phone, err.Error())
		}
		log.Errorf("send msg to %s success", phone)
	}

	//	str := "-1"
	//	rege, _ := regexp.Compile(`^(-*?\d*)(,|)\d*$`)
	//	result := rege.ReplaceAllString(str, `$1`)
	//	fmt.Println(result)
}
