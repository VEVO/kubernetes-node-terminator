package main

import (
	"fmt"
	"net"

	"github.com/golang/glog"
)

type datadogClient struct {
	remoteAddr *net.UDPAddr
	connection *net.UDPConn
	title      string
	tags       string
}

func newDatadogClient(svcAddress string) *datadogClient {
	var err error

	datadogClient := &datadogClient{}
	datadogClient.remoteAddr, err = net.ResolveUDPAddr("udp", svcAddress+":8125")
	if err != nil {
		glog.Fatal(err)
	}

	datadogClient.connection, err = net.DialUDP("udp", nil, datadogClient.remoteAddr)
	if err != nil {
		glog.Fatal(err)
	}

	return datadogClient
}

func (d datadogClient) sendEvent(eventText string) error {
	var err error
	event := fmt.Sprintf("_e{%d,%d}:%s|%s|t:error|%s", len(d.title), len(eventText), d.title, eventText, d.tags)
	_, err = d.connection.Write([]byte(event))

	return err
}
