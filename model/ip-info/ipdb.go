/**
 * Copyright (c) 2024-2025 Su Yang (soulteary)
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

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
