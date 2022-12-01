package main

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

//type CFResults struct {
//	Result []struct {
//		ID        string `json:"id" jsonschema:"required"`
//		ZoneID    string `json:"zone_id"`
//		ZoneName  string `json:"zone_name"`
//		Name      string `json:"name"`
//		Type      string `json:"type"`
//		Content   string `json:"content"`
//		Proxiable bool   `json:"proxiable"`
//		Proxied   bool   `json:"proxied"`
//		TTL       int    `json:"ttl"`
//		Locked    bool   `json:"locked"`
//		Meta      struct {
//			AutoAdded           bool   `json:"auto_added"`
//			ManagedByApps       bool   `json:"managed_by_apps"`
//			ManagedByArgoTunnel bool   `json:"managed_by_argo_tunnel"`
//			Source              string `json:"source"`
//		} `json:"meta"`
//		CreatedOn  time.Time `json:"created_on"`
//		ModifiedOn time.Time `json:"modified_on"`
//	} `json:"result"`
//	Success    bool          `json:"success"`
//	Errors     []interface{} `json:"errors"`
//	Messages   []interface{} `json:"messages"`
//	ResultInfo struct {
//		Page       int `json:"page"`
//		PerPage    int `json:"per_page"`
//		Count      int `json:"count"`
//		TotalCount int `json:"total_count"`
//		TotalPages int `json:"total_pages"`
//	} `json:"result_info"`
//}

func main() {
	cfclient := http.Client{}
	ipinfoclient := http.Client{}
	publicip, err := get_public_ip(&ipinfoclient)
	if err != nil {
		error.Error(err)
	}
	fmt.Println("The public IP is:", publicip)
	dnsnames, _ := openandread("dnslist")
	for _, line := range dnsnames {
		///primary logic goes here
		fmt.Println(line)
		domainid, err := getdomainid(line, &cfclient)
		if err != nil {
			fmt.Println(err)
		} else if domainid == "NoID" {
			fmt.Println("No domain ID, creating new record")
			new_ddns(line, domainid, publicip, &cfclient)
		} else {
			fmt.Println("The domain ID string is", domainid, ", updating record.")
			update_ddns(line, domainid, publicip, &cfclient)
		}
	}
}

func openandread(filename string) ([]string, error) {
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

func update_ddns(domain string, id string, ipaddr string, client *http.Client) (string, error) {
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

func new_ddns(domain string, id string, ipaddr string, client *http.Client) (string, error) {
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

func get_public_ip(client *http.Client) (string, error) {
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

func getdomainid(domain string, client *http.Client) (string, error) {
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
