package utils

import (
	log "web-server/alog"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
)

//load config
func LoadConfig(config interface{}, configFile string) error {
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0666)
	if err != nil {
		log.Error("open file error: ", err)
		return err
	}
	defer file.Close()

	str, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error("read all error: ", err)
		return err
	}

	if err = json.Unmarshal(str, config); err != nil {
		log.Error("Unmarshal error: ", err)
		return err
	}
	return nil
}

func ParsePbFromTextFile(filePath string, pb proto.Message) error {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("ReadFile error. error: ", err)
		return err
	}
	if err := proto.UnmarshalText(string(fileBytes), pb); err != nil {
		log.Error("UnmarshalText error. error: ", err)
		return err
	}
	return nil
}
