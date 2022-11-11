package utils

import (
	log "web-server/alog"
	"errors"
	"fmt"
	"net/http"

	"regexp"

	"github.com/gorilla/context"
)

// ParamTable ...
type ParamTable map[string]ParamEntry

// ParamEntry ...
// [Flag] param from body or head
// [IsOptional] if this param is optional, it's true
// [CheckInject] whether need to check inject
type ParamEntry struct {
	Flag        int // from body: 0, head: 1
	IsOptional  bool
	CheckInject bool // need to check inject: true
}

const (
	// KParamFromBody get value from http body
	KParamFromBody = iota
	// KParamFromHead get value from http head
	KParamFromHead
)

var (
	errKeyNotFound     = errors.New("key not found")
	errEmptyEntryKey   = errors.New("empty EntryKey")
	errEmptyEntryValue = errors.New("empty EntryValue")
	errEmptyRequest    = errors.New("empty request")
)

// GetValuesFromParamTable get value and do some checking
// Ps: "github.com/gorilla/context" is used to share value between the life time of r(*http.Request)
func GetValuesFromParamTable(params ParamTable, r *http.Request) error {
	if r == nil {
		return errEmptyRequest
	}
	log.Info(GetChannelIDLogMsg(r))
	for key, entry := range params {
		value := ""
		switch entry.Flag {
		case KParamFromBody:
			value = r.FormValue(key)
		case KParamFromHead:
			value = r.Header.Get(key)
		}

		// if not optional, check empty
		if value == "" && entry.IsOptional == false {
			return errors.New(errMsgValEmpty(key))
		}

		// check inject
		if entry.CheckInject == true && CheckInject(value) {
			return errors.New(errMsgValInject(key))
		}

		// use context to share
		context.Set(r, key, value)
		log.Infof("%s=%s ", key, value)
	}
	return nil
}

// CheckInject 防SQL注入的方法
func CheckInject(src string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|
			char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		log.Errorf("in PreventInject:: regexp.Compile Error[%s]", err)
		return true
	}

	return re.MatchString(src)
}

func errMsgValEmpty(key string) string {
	return fmt.Sprintf("[%s] should not be empty", key)
}

func errMsgValInject(key string) string {
	return fmt.Sprintf("[%s] should not inject any SQL", key)
}
