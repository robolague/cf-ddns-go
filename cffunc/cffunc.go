package cffunc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

//var Baseid string = os.Getenv("BASEID")

var Zone string = os.Getenv("ZONE")

var Auth_email string = os.Getenv("AUTH_EMAIL")

var Auth_key string = os.Getenv("AUTH_KEY")

var Bearer string = os.Getenv("BEARER")
var Cloudflareurl string = "https://api.cloudflare.com/client/v4/zones/" + Zone + "/dns_records"

var cfHeader = http.Header{
	"X-Auth-Email":  {Auth_email},
	"X-Auth-Key":    {Auth_key},
	"Content-Type":  {"application/json"},
	"Authorization": {Bearer},
}

type updatejson struct {
	Recordtype    string `json:"type"`
	Recordname    string `json:"name"`
	Recordcontent string `json:"content"`
	Recordttl     int    `json:"ttl"`
	Recordproxied bool   `json:"proxied"`
}

func Openandread(filename string) ([]string, error) {
	readFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	return fileLines, nil
}

func Update_ddns(domain string, id string, ipaddr string, client *http.Client) (string, error) {
	updateinfo, _ := json.Marshal(updatejson{Recordtype: "A", Recordname: domain, Recordcontent: ipaddr, Recordttl: 1, Recordproxied: true})
	req, err := http.NewRequest("PUT", Cloudflareurl+"/"+id, bytes.NewBuffer(updateinfo))
	if err != nil {
		return "", err
	}
	req.Header = cfHeader

	result, err := client.Do(req)
	if err != nil {
		return "", err
	}
	fmt.Println(result.Status)
	return result.Status, nil
}

func New_ddns(domain string, id string, ipaddr string, client *http.Client) (string, error) {
	updateinfo, _ := json.Marshal(updatejson{Recordtype: "A", Recordname: domain, Recordcontent: ipaddr, Recordttl: 1, Recordproxied: true})
	req, err := http.NewRequest("POST", Cloudflareurl+"/", bytes.NewBuffer(updateinfo))
	if err != nil {
		return "", err
	}
	req.Header = cfHeader
	result, err := client.Do(req)
	if err != nil {
		return "", err
	}
	fmt.Println(result.Status)
	return result.Status, nil
}

func Get_public_ip(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "http://ipinfo.io/ip", nil)
	if err != nil {
		return "", err
	}
	result, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	resultz, _ := io.ReadAll(result.Body)
	//resultz := "192.152.577.4"
	validornot := net.ParseIP(string(resultz))
	if validornot == nil {
		return "", errors.New("Bad IP")
	}
	return string(resultz), nil
}

func Getdomainid(domain string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", Cloudflareurl, nil)
	if err != nil {
		return "", err
	}
	query := req.URL.Query()
	query.Add("match", "all")
	query.Add("name", domain)
	req.URL.RawQuery = query.Encode()
	req.Header = cfHeader
	result, err := client.Do(req)
	defer result.Body.Close()
	if err != nil {
		return "", err
	}
	resultz, _ := io.ReadAll(result.Body)
	resultobject := gjson.GetBytes(resultz, "result.@pretty")
	domainid := resultobject.Get("#.id")
	if domainid.String() == "[]" {
		return "NoID", nil
	} else {
		return strings.Trim(domainid.String(), "[]\""), nil
	}
	return "", nil
}
