package network

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	. "ghost/global"
)

type RequestHandler struct {
	url      string
	method   string
	jsondata []byte
}

type ResponseHandler struct {
	//handler *RequestHandler
	text []byte
	err  error
}

type HttpClient struct {
	httpClient      *http.Client
	PendingWriteNum int
	in              chan *RequestHandler
	out             chan *ResponseHandler
}

var loginClient = &http.Client{}

func (client *HttpClient) Start() {
	Console.Info("Http Service Start ...")
	Log.Info("Http Service Start ...")
	client.init()
	go client.run()
}

func (client *HttpClient) init() {
	t := time.Duration(3) * time.Second
	client.httpClient = &http.Client{Timeout: t}

	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 1000
		Log.Info("invalid PendingWriteNum, reset to %v", client.PendingWriteNum)
	}

	client.in = make(chan *RequestHandler, client.PendingWriteNum)
	client.out = make(chan *ResponseHandler, client.PendingWriteNum)
}

func (client *HttpClient) run() {
	for {
		select {
		case myhandler, ok := <-client.in:
			if !ok {
				return
			}

			client.request(myhandler)
		case rsphandler, ok := <-client.out:
			if !ok {
				return
			}
			client.onResponse(rsphandler.text, rsphandler.err)
		}
	}
}

func (client *HttpClient) request(handler *RequestHandler) {
	req, _ := http.NewRequest(handler.method, handler.url, bytes.NewBuffer(handler.jsondata))

	go func() {
		resp, err := client.httpClient.Do(req)
		if err != nil {
			client.out <- &ResponseHandler{text: []byte(""), err: err}
		} else {
			defer resp.Body.Close()
			res, rerr := ioutil.ReadAll(resp.Body)
			client.out <- &ResponseHandler{text: res, err: rerr}
		}
	}()
}

func (client *HttpClient) onResponse(text []byte, err error) {
	if err != nil {
		Log.Warn("onResponse Read err:(%v)", err)
		return
	}
	Console.Info("%v", string(text))
}

func (client *HttpClient) Write(method string, url string, text []byte) {
	client.in <- &RequestHandler{method: method, url: url, jsondata: text}
}
