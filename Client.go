package main

import (
	"sync"
	"encoding/json"
	"time"
	"bytes"
	"io/ioutil"
	"net/http"
	"log"
)

type Client struct {
	sync.Mutex
	url      string
	client   *http.Client
	VMconfig *RPCVMConfig
}

type RPCRequest struct {
	Id      int64         `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	Id      int64           `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error"`
}

type RPCVMConfig struct {
	DisableStorage bool `json:"disableStorage"`
	DisableMemory  bool `json:"disableMemory"`
	DisableStack   bool `json:"disableStack"`
	FullStorage    bool `json:"fullStorage"`
}

func NewClient(url string) *Client {
	client := &Client{url: url}
	client.client = &http.Client{}
	client.VMconfig = &RPCVMConfig{
		DisableStorage: false,
		DisableMemory:  true,
		DisableStack:   true,
		FullStorage:    false,
	}
	return client
}

var baseId int64 = 0
var requestIndex int64 = 0
func (self *Client) GetUniqueId() int64 {
	//if the current time is greater than the previously used time seed plus the number of requests made at that time
	//then any new request made at time seed current cannot collide with any open requests so reset the index and seed
	if baseId + requestIndex < time.Now().UnixNano() {
		baseId = time.Now().UnixNano()
		requestIndex = 0
	} else {
		requestIndex += 1
	}
	return baseId + requestIndex
}

func (self *Client) MakeRequest(id int64, method string, params []interface{}) *RPCRequest {

	jsonReq := &RPCRequest{
		Id:      id,
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}

	return jsonReq
}

func (self *Client) MakeRequestUnique(method string, params []interface{}) *RPCRequest {
	jsonReq := &RPCRequest{
		Id:      self.GetUniqueId(),
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}

	return jsonReq
}

func (self *Client) BatchRequest(requests []*RPCRequest) ([]RPCResponse, error) {

	reqJSON, _ := json.Marshal(requests)
	req, err := http.NewRequest("POST", self.url, bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Length", (string)(len(reqJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcresp []RPCResponse
	err = json.Unmarshal(body, &rpcresp)
	if err != nil {
		log.Println(string(reqJSON))
		log.Println(string(body))
		return nil, err
	}

	return rpcresp, nil
}

func (self *Client) PostRequest(requests *RPCRequest) (*RPCResponse, error) {

	reqJSON, _ := json.Marshal(requests)
	req, err := http.NewRequest("POST", self.url, bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Length", (string)(len(reqJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcresp RPCResponse
	err = json.Unmarshal(body, &rpcresp)
	if err != nil {
		log.Println(string(reqJSON))
		log.Println(string(body))
		return nil, err
	}
	return &rpcresp, nil
}

//NOTE: PostRpcRequest is here to return a non nil rpcresp, and err object vs. PostRequest above
//For now, the response is converted to a string and returned as the first value.
//TODO: think about how to structure the return values here.
func (self *Client) PostRpcRequest(requests *RPCRequest) (string, *RPCResponse, error) {

	reqJSON, _ := json.Marshal(requests)
	req, err := http.NewRequest("POST", self.url, bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Length", (string)(len(reqJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := self.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var rpcresp RPCResponse
	err = json.Unmarshal(body, &rpcresp)
	if err != nil {
		log.Println(string(reqJSON))
		log.Println(string(body))
	}
	return string(body), &rpcresp, err
}
