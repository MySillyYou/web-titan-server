package handler

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
)

var (
	GMysqlDb *gorm.DB
	GvaMysql GeneralDB
	GConfig  map[string]interface{}
)

type mysqlDB struct {
	cli *gorm.DB
}

const sys = "system"

// TypeMysql mysql
func TypeMysql() string {
	return "Mysql"
}

// GeneralDB 也被 Pgsql 和 Mysql 原样使用
type GeneralDB struct {
	Path         string `mapstructure:"path" json:"path" yaml:"path"`                               // 服务器地址:端口
	Port         string `mapstructure:"port" json:"msql-port" yaml:"msql-port"`                     //:端口
	Config       string `mapstructure:"config" json:"config" yaml:"config"`                         // 高级配置
	Dbname       string `mapstructure:"db-name" json:"db-name" yaml:"db-name"`                      // 数据库名
	Username     string `mapstructure:"username" json:"username" yaml:"username"`                   // 数据库用户名
	Password     string `mapstructure:"password" json:"password" yaml:"password"`                   // 数据库密码
	MaxIdleConns int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"` // 空闲中的最大连接数
	MaxOpenConns int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"` // 打开到数据库的最大连接数
	LogMode      string `mapstructure:"log-mode" json:"log-mode" yaml:"log-mode"`                   // 是否开启Gorm全局日志
	LogZap       bool   `mapstructure:"log-zap" json:"log-zap" yaml:"log-zap"`                      // 是否通过zap写入日志文件
}
type Mysql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *GeneralDB) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

func (m *GeneralDB) GetLogMode() string {
	return m.LogMode
}

// GetDevicesInfoList  search from mysql
func (m mysqlDB) GetDevicesInfoList(info DevicesSearch) (list []DevicesInfo, total int64, err error) {
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

func (m mysqlDB) SaveDeviceInfo(incomeDaily DevicesSearch) error {
	var incomeDailyOld IncomeDaily
	result := GMysqlDb.Where("device_id = ?", incomeDaily.DeviceId).Where("user_id = ?", incomeDaily.UserId).First(&incomeDailyOld)

	if result.RowsAffected <= 0 {
		err := GMysqlDb.Create(&incomeDaily).Error
		return err
	} else {
		incomeDaily.ID = incomeDailyOld.ID
		err := GMysqlDb.Save(&incomeDaily).Error
		return err
	}
}

func CreateTables() error {
	M := GMysqlDb.Migrator()
	if M.HasTable(&HourDataOfDaily{}) {
		print("HourDataOfDaily:已经存在")
	} else {
		// 不存在就创建表
		err := GMysqlDb.AutoMigrate(&HourDataOfDaily{})
		if err != nil {
			return err
		}
	}
	if M.HasTable(&IncomeDaily{}) {
		print("IncomeDaily:已经存在")
	} else {
		// 不存在就创建表
		err := GMysqlDb.AutoMigrate(&IncomeDaily{})
		if err != nil {
			return err
		}
	}
	return nil
}

func GormMysql() (MysqlDb, error) {
	m := GvaMysql
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,
		SkipInitializeWithVersion: false,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), nil); err != nil {
		return nil, err
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		GMysqlDb = db
		mysqlDB := &mysqlDB{db}
		return mysqlDB, nil
	}

}
