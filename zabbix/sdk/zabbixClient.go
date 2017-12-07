package sdk

import (
	"errors"
	"strings"

	"github.com/kstrwind/lib-go/larix"
)

// ZBX default configure set
const (
	ZBXTimeOutMS int64 = 2000
)

// ZBXHeaders set default zabbix headers
// Note: keys must in lower-case letters
var ZBXHeaders = map[string]string{
	"content-type": "application/json-rpc",
}

// ZBXConf define a conf fields map for conf to decode
type ZBXConf struct {
	IP             string `ini:"ip"`
	Port           int    `ini:"port"`
	URI            string `ini:"uri"`
	User           string `ini:"user"`
	Passwd         string `ini:"passwd"`
	Headers        string `ini:"headers"`
	BaseTemplateID string `ini:"base_template_id"`
	TimeOutMs      int64  `ini:"timeout_ms"`
}

// ZBXClient define for create a new ZBXClient
type ZBXClient struct {
	IP             string            `json:"ip"`
	Port           int               `json:"port"`
	URI            string            `json:"uri"`
	User           string            `json:"user"`
	Passwd         string            `json:"passwd"`
	TimeOutMs      int64             `json:"timeout"`
	Headers        map[string]string `json:"headers"`
	sessionid      string
	httpClient     *larix.HttpClient
	id             int    //req id
	BaseTemplateID string `json:"_"`
}

// ZBXRequest define zabbix request body
type ZBXRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Auth    string      `json:"auth,omitempty"`
	ID      int         `json:"id"`
}

// ZBXErrorResponse define zabbix api error fields
type ZBXErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// ZBXResponse define zabbix response
type ZBXResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Error   ZBXErrorResponse `json:"error,omitempty"`
	ID      int              `json:"id"`
}

// ZBXInit init a zabbix client by conf
func ZBXInit(conf *ZBXConf) (*ZBXClient, error) {
	if conf == nil {
		return nil, errors.New("conf is nil")
	}

	var zc = &ZBXClient{}
	if conf.IP == "" {
		return nil, errors.New("IP field is empty")
	}
	zc.IP = conf.IP

	if conf.Port == 0 {
		return nil, errors.New("port field is empty")
	}
	zc.Port = conf.Port

	if conf.URI == "" {
		return nil, errors.New("server uri field is empty")
	}
	zc.URI = conf.URI

	if conf.User == "" {
		return nil, errors.New("user field is empty")
	}
	zc.User = conf.User

	if conf.Passwd == "" {
		return nil, errors.New("password is empty")
	}
	zc.Passwd = conf.Passwd

	zc.TimeOutMs = conf.TimeOutMs
	if conf.TimeOutMs == 0 {
		zc.TimeOutMs = ZBXTimeOutMS
	}

	//Headers set
	zc.Headers = ZBXHeaders
	if conf.Headers != "" {
		tmp := strings.Split(conf.Headers, ",")
		for _, headerStr := range tmp {
			if headerStr == "" {
				continue
			}
			tmp2 := strings.Split(headerStr, ":")
			key := strings.ToLower(strings.TrimSpace(tmp2[0]))
			value := strings.TrimSpace(tmp2[1])
			zc.Headers[key] = value
		}
	}

	zc.BaseTemplateID = conf.BaseTemplateID

	zc.httpClient = &larix.HttpClient{
		Ip:         zc.IP,
		Port:       zc.Port,
		Headers:    zc.Headers,
		Timeout_ms: zc.TimeOutMs,
		Host:       "",
	}

	zc.id = 0
	zc.sessionid = ""
	return zc, nil
}
