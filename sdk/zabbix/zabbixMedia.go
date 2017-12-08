package zabbix

// ZBXMedia define zabbix media object
type ZBXMedia struct {
	MediaTypeID string `json:"mediatypeid"`
	SendTo      string `json:"sendto"`
	Active      string `json:"active"`
	Severity    string `json:"severity"`
	Period      string `json:"period"`
}

// ZBXMediaType define zabbix media type
type ZBXMediaType struct {
	MediaTypeID string `json:"mediatypeid"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ExecPath    string `json:"exec_path"`
	GsrmModem   string `json:"gsm_modem"`
	Passwd      string `json:"passwd"`
	SMTPEmal    string `json:"smtp_email"`
	SMTPHelo    string `json:"smtp_helo"`
	SMTPServer  string `json:"smtp_server"`
	Status      string `json:"status"`
	UserName    string `json:"username"`
	ExecParams  string `json:"exec_params"`
}
