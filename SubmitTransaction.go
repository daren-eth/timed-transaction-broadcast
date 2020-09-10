package main

func BroadcastTransaction(client *Client, trans string) (string, error) {
	request := client.MakeRequestUnique("eth_sendRawTransaction", []interface{}{trans})
	responseAsString, _, err := client.PostRpcRequest(request)
	return responseAsString, err
}
