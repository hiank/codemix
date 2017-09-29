package config

import (
	"path/filepath"
	"os"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type FilterConfig struct {

	DirS 		string 		`json:"dir-s"`
	DirD 		string 		`json:"dir-d"`
	Ignore 		[]string	`json:"ignore"`
	IgnorePre 	[]string 	`json:"ignore-pre"`
	MatchSuf 	[]string 	`json:"match-suf"`
}

type MixConfig struct {

	NumMax 	int 				`json:"num-max"`
	NumMin 	int 				`json:"num-min"`
	IgnoreRandname []string 	`json:"ignore-randname"`
}


type Config struct {

	F 	FilterConfig  	`json:"filter"`
    M  	MixConfig      	`json:"mix"`
}


var _config *Config
// GetConfig 获得工程配置
func GetConfig() *Config {

	if _config != nil {
		return _config
	}
	
	con := new(Config)
	
	// path, _ := os.Getwd()
	path := filepath.Dir(os.Args[0])
	// fmt.Println("path :" + path)
	buf, err := ioutil.ReadFile(path + string(os.PathSeparator) + "setting.json")
	if err != nil {
		fmt.Println("read file setting.json error :" + err.Error())
		_config = con
		return con
	}

	err = json.Unmarshal(buf, con)
	if err != nil {
		fmt.Println("unmarshal json error :" + err.Error())
		return nil
	}
	
	_config = con
	return con
}

