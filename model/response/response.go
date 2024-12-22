package response

import (
	"bytes"
	"strings"

	"github.com/soulteary/ip-helper/model/define"
	"github.com/soulteary/ip-helper/model/fn"
)

func RenderJSON(ipaddr string, dbInfo []string) map[string]any {
	return map[string]any{"ip": ipaddr, "info": dbInfo}
}

func RenderHTML(config *define.Config, urlPath string, globalTemplate []byte, ipaddr string, dbInfo []string) []byte {
	template := bytes.ReplaceAll(globalTemplate, []byte("%IP_ADDR%"), []byte(ipaddr))
	template = bytes.ReplaceAll(template, []byte("%DOMAIN%"), []byte(config.Domain))
	template = bytes.ReplaceAll(template, []byte("%DATA_1_INFO%"), []byte(strings.Join(fn.RemoveDuplicates(dbInfo), " ")))
	template = bytes.ReplaceAll(template, []byte("%DOCUMENT_PATH%"), []byte(urlPath))
	template = bytes.ReplaceAll(template, []byte("%ONLY_DOMAIN%"), []byte(fn.GetDomainOnly(config.Domain)))
	template = bytes.ReplaceAll(template, []byte("%ONLY_DOMAIN_WITH_PORT%"), []byte(fn.GetDomainWithPort(config.Domain)))
	return template
}
