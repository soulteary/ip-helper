package ipInfo

import (
	"github.com/soulteary/ip-helper/model/fn"
)

func (db IPDB) FindByIPIP(ip string) []string {
	info, err := db.IPIP.Find(ip, "CN")
	if err != nil {
		info = []string{"未找到 IP 地址信息"}
	}
	return fn.RemoveDuplicates(info)
}
