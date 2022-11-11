package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
	log "web-server/alog"
)

type SQLHelper struct {
	db *sql.DB
}

var helper *SQLHelper

func GetSQLHelper() *SQLHelper {
	if helper == nil {
		helper = &SQLHelper{}
	}

	return helper
}

/*
	2022-10-08添加事务处理
*/
func GetSQLDB() *sql.DB {
	if helper == nil {
		helper = &SQLHelper{}
		return helper.db
	}
	return helper.db
}

func (this *SQLHelper) Init(dataSourceName string) {
	var err error
	this.db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal("Open db error. error: ", err)
	}

	this.db.SetMaxOpenConns(2000)
	this.db.SetMaxIdleConns(200)
	this.db.SetConnMaxLifetime(time.Second * 3600)
	err = this.db.Ping()
	if err != nil {
		log.Fatal("db Ping error. error: ", err)
	}
}

func (this *SQLHelper) GetDB() *sql.DB {
	if this.db != nil {
		return this.db
	}
	return nil
}

func (this *SQLHelper) Existed() bool {
	if this.db == nil {
		return false
	}
	return true
}

func (this *SQLHelper) SetDB(db *sql.DB) {
	this.db = db
}

func (this *SQLHelper) Close() {
	if this.db != nil {
		this.db.Close()
	}
}

/*
	通过sqlClause来获取不定字段的数据
	@sqlClause sql语句
	@return 返回map类型的[]，一行记录就是一个map，map里的key即为字段名。
		注意，字段名区分大小写，这里的字段名全部小写；如果失败返回nil, error
*/
func (this *SQLHelper) GetQueryDataList(sqlClause string, args ...interface{}) ([]map[string]string, error) {
	rows, err := this.db.Query(sqlClause, args...)
	if err != nil {
		log.Error("query error: ", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	dataList := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		data := make(map[string]string)
		for i, col := range values {
			//			if col == nil {
			//				continue
			//			}

			key := columns[i]
			key = strings.ToLower(key)
			data[key] = string(col)

		}
		//		log.Info(&data)
		dataList = append(dataList, data)
	}

	return dataList, nil
}

/*
	执行sqlClause
	@sqlClause sql语句
	@return 返回两个参数，int64是执行成功的记录行数，error是sql执行结果。
		失败返回0, err
*/
func (this *SQLHelper) ExecSqlClause(sqlClause string) (int64, error) {
	result, err := this.db.Exec(sqlClause)
	if err != nil {
		log.Error("exec sql error: ", err)
		return 0, err
	}
	rowAffected, err := result.RowsAffected()

	return rowAffected, err
}

/*
	通过sqlClause插入数据
	@sqlClause sql语句
	@return 返回两个参数，int64是插入成功后的自增ID，error是sql执行结果。
		失败返回0, err
*/
func (this *SQLHelper) Insert(sqlClause string) (int64, error) {
	result, err := this.db.Exec(sqlClause)
	if err != nil {
		log.Error("exec sql error: ", err)
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()

	return lastInsertId, err
}

/*
	通过map插入数据
	@data 需要插入的数据，data-key为表的字段名，data-value为对应key的表项值
	@table 表名
	@return 返回两个参数，int64是插入成功后的自增ID，error是sql执行结果。
		失败返回0, err
*/
func (this *SQLHelper) InsertMapData(data map[string]string, table string, replace bool) (int64, error) {

	if data == nil || len(data) <= 0 {
		return 0, errors.New("InsertMapData data is nil")
	}

	keyClause := "("
	valueClause := "("
	replaceClause := ""
	for key, value := range data {
		if strings.Contains(key, "`") {
			keyClause += fmt.Sprintf("%s,", key)
		} else {
			keyClause += fmt.Sprintf("`%s`,", key)
		}
		if value == "null" || value == "NULL" {
			valueClause += "null,"
		} else {
			valueClause += "'" + value + "',"
		}
		replaceClause += key + "=values(" + key + "),"
	}

	keyClause = keyClause[:len(keyClause)-1] + ")"
	valueClause = valueClause[:len(valueClause)-1] + ")"
	replaceClause = replaceClause[:len(replaceClause)-1]

	sqlClause := ""
	if replace {
		sqlClause = fmt.Sprintf(`insert into %s %s values %s on duplicate key update %s;`, table, keyClause, valueClause, replaceClause)
	} else {
		sqlClause = fmt.Sprintf(`insert into %s %s values %s;`, table, keyClause, valueClause)
	}
	//log.Info(sqlClause)

	lastInsertId, err := this.Insert(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}

	return lastInsertId, nil

}

/*
	通过[]map批量插入数据
	@dataList 需要插入的数据，列表中包含多个map，每个map的key为表的字段名，value为对应key的表项值
	@table 表名
	@return 返回两个参数，int64是插入成功后的第一条记录的自增ID，error是sql执行结果。
		失败返回0, err
*/
func (this *SQLHelper) InsertListData(dataList []map[string]string, table string, replace bool) (int64, error) {

	if dataList == nil || len(dataList) <= 0 {
		return 0, errors.New("InsertListData dataList is nil")
	}

	//由于golang的map顺序是不固定的，所以增加一个key列表，用来存放key
	keyList := make([]string, 0)
	keyClause := "("
	replaceClause := ""
	for key, _ := range dataList[0] {
		if strings.Contains(key, "`") {
			keyClause += fmt.Sprintf("%s,", key)
		} else {
			keyClause += fmt.Sprintf("`%s`,", key)
		}
		replaceClause += key + "=values(" + key + "),"
		keyList = append(keyList, key)
	}

	keyClause = keyClause[:len(keyClause)-1] + ")"
	replaceClause = replaceClause[:len(replaceClause)-1]

	valueClause := ""
	for _, data := range dataList {

		tempClause := "("
		//从key列表中查找每个key对应的值
		for _, key := range keyList {
			if data[key] == "null" || data[key] == "NULL" {
				tempClause += "null,"
			} else {
				tempClause += "'" + data[key] + "',"
			}
		}

		tempClause = tempClause[:len(tempClause)-1] + ")"
		valueClause += tempClause + ","
	}

	valueClause = valueClause[:len(valueClause)-1]

	sqlClause := ""
	if replace {
		sqlClause = fmt.Sprintf(`insert into %s %s values %s on duplicate key update %s;`, table, keyClause, valueClause, replaceClause)
	} else {
		sqlClause = fmt.Sprintf(`insert into %s %s values %s;`, table, keyClause, valueClause)
	}
	//log.Info(sqlClause)

	lastInsertId, err := this.Insert(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}

	return lastInsertId, nil

}

/*
 * 通过map插入不定字段
 */
func (this *SQLHelper) InsertDataByMap(tableName string, insertMap map[string]interface{}) (int64, error) {
	mpLen := len(insertMap)
	valueList := make([]interface{}, mpLen)
	fieldListStr := " "
	tmpStr := " "
	i := 0
	for key, value := range insertMap {
		valueList[i] = value
		fieldListStr = fieldListStr + key
		tmpStr = tmpStr + "?"
		if i < mpLen-1 {
			fieldListStr = fieldListStr + ","
			tmpStr = tmpStr + ","
		}
		i++
	}
	sqlCluse := "insert into " + tableName + "(" + fieldListStr + ") values(" + tmpStr + ")"
	result, err := this.db.Exec(sqlCluse, valueList...)
	if err != nil {
		log.Error("SQL Error, SQL:", sqlCluse, err)
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()

	return lastInsertId, err
}

func (this *SQLHelper) UpdateDataByMap(tableName string, dataMap map[string]interface{}, sqlCondition string) (int64, error) {
	sqlCluse := `update ` + tableName + ` set `
	fieldLen := len(dataMap)
	valueList := make([]interface{}, fieldLen)
	t := 0
	for key, value := range dataMap {
		valueList[t] = value
		sqlCluse = sqlCluse + key + "=?"
		if t < fieldLen-1 {
			sqlCluse = sqlCluse + ","
		}
		t++
	}
	sqlCluse = sqlCluse + sqlCondition
	result, err := this.db.Exec(sqlCluse, valueList...)
	if err != nil {
		log.Error("SQL Error, SQL:", sqlCluse, err)
		return 0, err
	}
	rowAffected, err := result.RowsAffected()

	return rowAffected, err
}
