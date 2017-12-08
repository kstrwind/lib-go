package zabbix

// ZBXUserGroup define properties in zabbix user group
type ZBXUserGroup struct {
	UsrGrpid    string `json:"usrgrpid"`
	Name        string `json:"name"`
	DebugMode   string `json:"debug_mode"`
	GUIAccess   string `json:"gui_access"`
	UsersStatus string `json:"users_status"`
}

// ZBXPermission define zabbix permission
type ZBXPermission struct {
	ID         string `json:"id"`
	Permission string `json:"permission"`
}

// ZBXMUserGroup define zabbix user group for monitor
type ZBXMUserGroup struct {
	ZBXUserGroup
	Users  []*ZBXUser       `json:"users,omitempty"`
	Rights []*ZBXPermission `json:"rights,omitempty"`
}
