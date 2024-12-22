package ipInfo

import (
	"log"

	"github.com/soulteary/ipdb-go"
)

type IPDB struct {
	IPIP *ipdb.City
}

func InitIPDB() IPDB {
	db, err := ipdb.NewCity("./data/ipipfree.ipdb")
	if err != nil {
		log.Fatal(err)
	}
	return IPDB{IPIP: db}
}
