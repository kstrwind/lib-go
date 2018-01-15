package zabbix

import (
	"fmt"
	"testing"
)

func Test_ClientLogin(t *testing.T) {
	tstConf := &ZBXConf{
		IP:        "192.168.56.101",
		Port:      8003,
		URI:       "/api_jsonrpc.php",
		User:      "Admin",
		Passwd:    "zabbix",
		TimeOutMs: 1000,
	}
	zCase, err := ZBXInit(tstConf)
	if err != nil {
		fmt.Println("zabbix init failed:", err.Error())
		return
	}
	fmt.Println("Zabbix init conf succ")
	err = zCase.UserLogin()
	if err != nil {
		fmt.Println("zabbix login failed:", err.Error())
	}
	fmt.Println("Zabbix user login succ")
	fmt.Println("sessionid is:", zCase.SessionID())
	err = zCase.UserLogout()
	if err != nil {
		fmt.Println("zabbix logout failed:", err.Error())
		t.Log(err)
	}
	fmt.Println("Zabbix logout succ:", zCase.SessionID())

}
