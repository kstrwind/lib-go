package larix

/**
 * conf interface , plugins frame, support for multi
 * types conf
 *
 **/
import (
	"fmt"
	"os"
	"strings"

	ini "github.com/go-ini/ini"
)

//define conftype
type LarixConf map[string]map[string]interface{}

// use file suffix to determine dial  parser
const (
	CONF_INI int = iota
	CONF_YAML
	CONF_GO
	CONF_OVER
)

var dialConf map[string]int = map[string]int{
	"ini":  CONF_INI,
	"yml":  CONF_YAML,
	"yaml": CONF_YAML,
	"go":   CONF_GO,
}

type Conf struct {
	//configfiles
	ConfigFiles map[int][]string

	//current parse conf content
	ConfigCache LarixConf
}

//Conf instance
var _conf *Conf = nil

func ConfInit(configs []string) {
	if len(configs) < 1 {
		os.Stdout.WriteString("config files not set")
		return
	}

	if _conf != nil {
		return
	}

	_conf = &Conf{
		ConfigFiles: make(map[int][]string),
		ConfigCache: LarixConf{},
	}

	// pasre config file type
	for _, file := range configs {
		tmp_res := strings.Split(file, ".")
		r_suffix := tmp_res[len(tmp_res)-1]
		r_type, exists := dialConf[r_suffix]
		if !exists {
			os.Stdout.WriteString("config file [" + file + "] type [" + r_suffix + "] not support")
		} else {
			/*
				_, exists := _conf.ConfigFiles[r_type]
				if !exists {
					_conf.ConfigFiles[r_type] = make([]string)
				}*/
			_conf.ConfigFiles[r_type] = append(_conf.ConfigFiles[r_type], file)
		}
	}

	//load conf
	for c_type, c_files := range _conf.ConfigFiles {
		if c_type == CONF_INI {
			loadIni(c_files)
		} else if c_type == CONF_YAML {
			loadYaml(c_files)
		} else if c_type == CONF_GO {
			loadGo(c_files)
		} else {
			os.Stderr.WriteString(fmt.Sprintf("unknoen conf type %d, will not load", c_type))
		}
	}

	return
}

func loadIni(configs []string) {
	//step1: load config files
	//current not close fgf
	cfg := ini.Empty()

	for _, file := range configs {
		//check if file exists
		_, err := os.Stat(file)
		//here we can't use os.IsExist, because when file exist, err is nil, ca't use it to check file exists
		if err != nil && os.IsNotExist(err) {
			os.Stderr.WriteString("config file [" + file + "] not exists")
			continue
		}

		err = cfg.Append(file)
		if err != nil {
			os.Stderr.WriteString("load config file [" + file + "] failed")
		}
	}

	//parse config
	sections := cfg.SectionStrings()

	for _, section := range sections {
		lower_section := strings.ToLower(section)
		keys := cfg.Section(section).KeyStrings()

		_, exists := _conf.ConfigCache[lower_section]
		if !exists {
			_conf.ConfigCache[lower_section] = make(map[string]interface{})
		}

		for _, key := range keys {
			lower_key := strings.ToLower(key)
			_conf.ConfigCache[lower_section][lower_key] = cfg.Section(section).Key(key).Value()

		}
	}

}

func loadYaml(configs []string) {
}

func loadGo(configs []string) {
}

//get all sections
func GetSectionsKeys() []string {
	res := []string{}
	if _conf == nil || len(_conf.ConfigCache) == 0 {
		return res
	}

	for key, _ := range _conf.ConfigCache {
		res = append(res, key)
	}
	//sort.Sort(res)
	return res
}

//Get section all keys and values
func GetSection(section string) map[string]interface{} {
	//res := make(map[string]interface{})
	if _conf == nil || len(_conf.ConfigCache) == 0 {
		return map[string]interface{}{}
	}

	res, exists := _conf.ConfigCache[section]
	if !exists {
		return map[string]interface{}{}
	}

	return res
}
