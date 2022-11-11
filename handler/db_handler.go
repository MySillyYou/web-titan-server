package handler

import (
"fmt"
	_ "gorm.io/gorm"
)

type MysqlDb interface {
	GetDevicesInfoList(info DevicesSearch) ([]DevicesInfo, int64, error)
	SaveDeviceInfo(incomeDaily DevicesSearch) error
}

var mysqlDb MysqlDb

// NewCacheDB New Cache DB
func NewGormMysqlDB(url string, dbType string) {
	var err error

	switch dbType {
	case TypeMysql():
		mysqlDb, err = GormMysql()
	default:
		panic("unknown DB type")
	}

	if err != nil {
		e := fmt.Sprintf("NewCacheDB err:%v , url:%v", err, url)
		panic(e)
	}
}
