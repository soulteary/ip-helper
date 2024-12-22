package ipInfo

import (
	"github.com/soulteary/ipdb-go"
)

type IPDB struct {
	IPIP *ipdb.City
}

func InitIPDB(ipipDB string) (IPDB, error) {
	ipip, err := ipdb.NewCity(ipipDB)
	if err != nil {
		return IPDB{}, err
	}
	return IPDB{IPIP: ipip}, nil
}
