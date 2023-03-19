package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
)

func CheckByDNS(domain string) bool {
	ips, err := net.LookupIP(domain)
	fmt.Println(ips, err)
	if err != nil {
		return true
	}

	return false
}

type Res struct {
	DomainList []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"domainList"`
	UnRegisteredDomainList []interface{} `json:"unRegisteredDomainList"`
	RegisteredDomainList   []interface{} `json:"registeredDomainList"`
	ErrorDomainList        []string      `json:"errorDomainList"`
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
	s := resp.DomainList[0].Status
	available := s != "registered" && s != "forbid" && (s == "unRegisteredAdd") || (s == "error")
	return available
}
