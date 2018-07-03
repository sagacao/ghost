package report

type reportData struct {
	User    uint32 `json:"roleid"`
	Account string `json:"userid"` //plat account
	Server  uint32 `json:"serverid"`
	Level   uint32 `json:"lev"`
	Prof    uint32 `json:"prof"`
	Sex     uint32 `json:"sex"`
}
