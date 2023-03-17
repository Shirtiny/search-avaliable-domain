package check

import (
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"bytes"
	"encoding/json"
	"fmt"

	"strings"
)

// isDomainAvailable 检查域名是否被注册
func CheckByDNS(domain string) bool {
	txt, err := net.LookupHost(domain)
	// fmt.Println(txt, err)
	if err != nil {
		return true // 域名未被注册
	}

	for _, line := range txt {
		if strings.Contains(line, "No match for domain") {
			return true // 域名未被注册
		}
	}

	return false // 域名已被注册
}

type Res struct {
	DomainList []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"domainList"`
	UnRegisteredDomainList []interface{}    `json:"unRegisteredDomainList"`
	RegisteredDomainList   []interface{}    `json:"registeredDomainList"`
	ErrorDomainList        []string `json:"errorDomainList"`
}

func CheckIsDomainAvailableByApi(domain string) bool {
	url := "https://www.dns.com.cn/show/domain/search/domainBatchCheckSearch.do"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("domainName", domain)
	_ = writer.WriteField("validateToken", "615595e473884258885a07fa09b96da7")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Cookie", "SESSION=347717f5-2796-43a2-9e15-1aa5afcd15dd")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp := Res{}
	e := json.Unmarshal(body, &resp)
	fmt.Printf("%+v, %+v\n", resp, e)
	not := resp.DomainList[0].Status == "error"
	return !not
}
