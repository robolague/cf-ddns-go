package main

import (
	"cfddns/cffunc"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var (
	ddnsUpdateCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddns_update_count",
			Help: "Number of DDNS updates made.",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(ddnsUpdateCount)
}

func main() {
	cfclient := http.Client{}
	ipinfoclient := http.Client{}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()

	for {
		publicip, err := cffunc.Get_public_ip(&ipinfoclient)
		if err != nil {
			error.Error(err)
		}
		fmt.Println("The public IP is:", publicip)
		dnsnames, _ := cffunc.Openandread("dnslist")
		for _, line := range dnsnames {
			///primary logic goes here
			fmt.Println(line)
			domainid, err := cffunc.Getdomainid(line, &cfclient)
			if err != nil {
				fmt.Println(err)
			} else if domainid == "NoID" {
				fmt.Println("No domain ID, creating new record")
				cffunc.New_ddns(line, domainid, publicip, &cfclient)
				ddnsUpdateCount.WithLabelValues("new_record").Inc()
			} else {
				fmt.Println("The domain ID string is", domainid, ", updating record.")
				cffunc.Update_ddns(line, domainid, publicip, &cfclient)
				ddnsUpdateCount.WithLabelValues("updated_record").Inc()
			}
		}
		time.Sleep(2 * time.Hour)
	}
}
