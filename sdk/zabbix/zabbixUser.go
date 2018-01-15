package zabbix

import (
	"bytes"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/kstrwind/lib-go/larix"
)

// ZBXInternalUser for zabbix internal user check
var ZBXInternalUser = map[string]int{
	"guest": 1,
	"Admin": 1,
}

// IsZBXUser check if userName is zabbix user
func (z *ZBXClient) IsZBXUser(userName string) bool {
	if ZBXInternalUser[userName] == 1 {
		return true
	}
	return false
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
		JSONRPC: ZBXJSONVersion,
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

// GenUserPasswd to create a passwd for user
// for user init to zabbix should with passwd
func (z *ZBXClient) GenUserPasswd(uName string) string {
	var passwd bytes.Buffer
	passwd.WriteString(uName)
	passwd.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	// + 2位随机
	r := rand.New(rand.NewSource(time.Now().Unix()))
	passwd.WriteString(string(r.Intn(10)))
	passwd.WriteString(string(r.Intn(10)))
	return passwd.String()
}

// ZBXUser define zabbix user object
type ZBXUser struct {
	UserID        string `json:"userid"`
	Alias         string `json:"alias"`
	AttemptClock  string `json:"attempt_clock"`
	AttemptFailed string `json:"attempt_failed"`
	AttemptIP     string `json:"attempt_ip"`
	AutoLogin     string `json:"autologin"`
	AutoLogout    string `json:"autologout"`
	Lang          string `json:"lang"`
	Name          string `json:"name"`
	Refresh       string `json:"refresh"`
	RowsPerPage   string `json:"rows_per_page"`
	SurName       string `json:"surname"`
	Theme         string `json:"theme"`
	Type          string `json:"type"`
	URL           string `json:"url"`
}

// ZBXMUser define a complete user for monitor
type ZBXMUser struct {
	ZBXUser
	UsrGrps []*ZBXUserGroup `json:"zabbixusergroup"`
	Medias  []*ZBXMedia     `json:"medias"`
}
