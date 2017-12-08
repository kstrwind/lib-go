package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"zabbix/sync_worker/global"

	"github.com/kstrwind/lib-go/larix"
)

// ZBX default configure set
const (
	ZBXTimeOutMS    int64  = 2000
	ZBXJSONVersion  string = "2.0"
	ZBXDefaultRetry uint32 = 3
)

// ZBXHeaders set default zabbix headers
// Note: keys must in lower-case letters
var ZBXHeaders = map[string]string{
	"content-type": "application/json-rpc",
}

// ZBXInternalUser for zabbix internal user check
var ZBXInternalUser = map[string]int{
	"guest": 1,
	"Admin": 1,
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
	Result  interface{}      `json:"result"`
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

// ZBXID to init a request id
func (z *ZBXClient) ZBXID() int {
	z.id++
	return z.id
}

// HasLogin check if zabbix client has login
// return true for login, false for not login
func (z *ZBXClient) HasLogin() bool {
	if z.sessionid == "" {
		return false
	}
	return true
}

// SessionID get current zabbix client session
func (z *ZBXClient) SessionID() string {
	return z.sessionid
}

// IsZBXGroup check if groupid is zabbix self group
// Note: not check if group type is internal
func (z *ZBXClient) IsZBXGroup(groupID string) (bool, error) {
	rGroupid, err := strconv.Atoi(groupID)
	//for groupid check failed
	if err != nil {
		return false, err
	}
	if rGroupid < 1 || rGroupid == 2 || rGroupid == 3 || rGroupid >= 15 {
		return false, nil
	}
	return true, nil
}

// IsZBXUser check if userName is zabbix user
func (z *ZBXClient) IsZBXUser(userName string) bool {
	if ZBXInternalUser[userName] == 1 {
		return true
	}
	return false
}

// GenPasswd to create a passwd for user
// for user init to zabbix should with passwd
func (z *ZBXClient) GenPasswd(uName string) string {
	var passwd bytes.Buffer
	passwd.WriteString(uName)
	passwd.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	// + 2位随机
	r := rand.New(rand.NewSource(time.Now().Unix()))
	passwd.WriteString(string(r.Intn(10)))
	passwd.WriteString(string(r.Intn(10)))
	return passwd.String()
}

// UserLogin  login zabbix api server
func (z *ZBXClient) UserLogin() error {
	reqBody := ZBXRequest{
		JSONRPC: ZBXJSONVersion,
		Method:  "user.login",
		ID:      z.ZBXID(),
	}
	reqBody.Params = map[string]string{
		"user":     z.User,
		"password": z.Passwd,
	}

	res, err := z.request(reqBody, ZBXDefaultRetry)
	if err != nil {
		logInfo := map[string]interface{}{
			"message": "Zabbix login failed",
			"user":    z.User,
			"error":   err.Error(),
		}
		larix.LogFatal(logInfo)
	}

	ok := false
	if z.sessionid, ok = res.Result.(string); !ok || z.sessionid == "" {
		logInfo := map[string]interface{}{
			"message": "Zabbix login failed for sessionid not found",
			"res":     res.Result,
		}
		larix.LogFatal(logInfo)
		return errors.New("login return no sessionid")
	}

	return nil
}

// UserLogout to logout a zabbix client from zabbix api server
func (z *ZBXClient) UserLogout() error {
	//check if has login
	if !z.HasLogin() {
		return nil
	}
	reqBody := ZBXRequest{
		JSONRPC: global.ZBX_JSONRPC_VERSION,
		Method:  "user.logout",
		ID:      z.ZBXID(),
		Auth:    z.sessionid,
	}

	res, err := z.request(reqBody, ZBXDefaultRetry)
	if err != nil {
		logInfo := map[string]interface{}{
			"message": "Zabbix logout failed",
			"user":    z.User,
			"error":   err.Error(),
		}
		larix.LogFatal(logInfo)
	}

	if logoutRes, ok := res.Result.(bool); !ok || !logoutRes {
		logInfo := map[string]interface{}{
			"message": "Zabbix logout failed for return failed",
			"user":    z.User,
			"res":     res.Result,
		}
		larix.LogFatal(logInfo)
		return errors.New("logout return failed")
	}

	//logout succ
	z.sessionid = ""
	return nil
}

// request for request a zabbix server
func (z *ZBXClient) request(reqBody interface{}, retry uint32) (*ZBXResponse, error) {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	if retry < 1 {
		retry = 1
	}

	var reqRes []byte
	for {
		if retry < 1 {
			break
		}
		retry--
		reqBodyReader := bytes.NewBuffer(reqBodyBytes)

		reqRes, err = z.httpClient.Request("POST", z.URI, reqBodyReader)

		if err != nil {
			logInfo := map[string]interface{}{
				"message": "Zabbix logout failed",
				"error":   err.Error(),
				"retry":   retry,
			}
			larix.LogFatal(logInfo)
			//sleep a while and then retry
			time.Sleep(global.ZBX_HTTP_RETRY_INTERNAL_MS * time.Millisecond)
			continue
		}
		break
	}
	// if retry and still failed
	if err != nil {
		return nil, err
	}

	// res data decode
	var resData *ZBXResponse
	err = json.Unmarshal(reqRes, resData)
	if err != nil {
		return nil, err
	}

	//check if zabbix server return error
	if resData.Error.Code != 0 {
		return nil, errors.New(resData.Error.Message)
	}
	return resData, nil
}
