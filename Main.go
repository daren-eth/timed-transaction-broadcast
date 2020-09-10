package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"time"
)

type Configuration struct {
	ETHNodeURL           string                 `json:"eth_node_url"`
	GasStationURL        string                 `json:"gas_station_url"`
	GasPricedTransaction []GasPricedTransaction `json:"gas_priced_transactions"`
}

type GasPricedTransaction struct {
	GasPrice  int      `json:"gas_price"`
	SignedTxs []string `json:"signed_txs"`
}

type GasPrice struct {
	Fast    int `json:"fast"`
	Fastest int `json:"fastest"`
	SafeLow int `json:"safeLow"`
	Average int `json:"average"`
}

var lastGasPrice = math.MaxInt16
var conf = Configuration{}
var submitted = make(map[string]bool)
var txHashRegex = regexp.MustCompile("(0x[A-Fa-f0-9]{64})")

func main() {
	if !readConfig() {
		return
	}

	getGasPrice() //initial gas price
	go func() {
		//start timer to query ethgasstation and update lastGasPrice
		ticker := time.NewTicker(1 * time.Minute)
		for {
			<-ticker.C
			getGasPrice()
		}
	}()

	for {
		for _, gasTx := range conf.GasPricedTransaction {
			if lastGasPrice <= gasTx.GasPrice {
				for _, signedTx := range gasTx.SignedTxs {
					if _, ok := submitted[signedTx]; ok {
						continue
					}
					response, err := BroadcastTransaction(NewClient(conf.ETHNodeURL), signedTx)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						txHash := txHashRegex.FindString(response)
						fmt.Println(response)
						//you could add some shell call here to notify you that the tx has been submitted
						//I have an application that allows me to use push messaging to send messages directly to my mobile device
						//exec.Command("/bin/sh", "-c", "curl --include --request POST --header  https://some notification address")
						submitted[signedTx] = true
					}
				}
			}
		}
		time.Sleep(1 * time.Minute)
		readConfig()
	}
}

func readConfig() bool {
	var newConf = Configuration{}
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return false
	}

	err = json.Unmarshal(content, &newConf)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if !reflect.DeepEqual(conf, newConf) {
		conf = newConf
		fmt.Println("Config Read:\n", string(content))
	}
	return true
}

func getGasPrice() {
	client := http.Client{
		Timeout: time.Second * 20,
	}

	req, err := http.NewRequest(http.MethodGet, conf.GasStationURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var prices = GasPrice{}
	err = json.Unmarshal(body, &prices)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	lastGasPrice = prices.Average / 10
	fmt.Println("Gas Price Updated: ", lastGasPrice)
}
