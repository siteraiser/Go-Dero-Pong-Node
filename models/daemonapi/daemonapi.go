package daemonapi

//import "os"
//import "fmt"
//import "time"

import (
	"context"
	"fmt"
	"log"
	settings "node/models/settings"

	"github.com/civilware/Gnomon/rwc"
	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/deroproject/derohe/rpc"
	"github.com/gorilla/websocket"
)

const LOGGING = false

/*
rpc v3... Probably don't need sockets but hey...

	func getClient() (jsonrpc.RPCClient, context.Context, context.CancelFunc) {
		client := jsonrpc.NewClient("http://" + settings.GetWalletConn().Api + "/json_rpc")
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)

		return client, ctx, cancel
	}

	func getClient() jsonrpc.RPCClient {
		return jsonrpc.NewClient("http://" + settings.GetWalletConn().Api + "/json_rpc")
	}
*/
func GetTxPool() (rpc.GetTxPool_Result, error) {

	var pool_result rpc.GetTxPool_Result

	WS, _, err := websocket.DefaultDialer.Dial("ws://"+settings.GetDaemonConn()+"/ws", nil)
	input_output := rwc.New(WS)
	RPC := jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)
	if err = RPC.CallResult(context.Background(), "DERO.GetTxPool", nil, &pool_result); err != nil {
		if LOGGING {
			log.Fatal("Error: %s ", err)
		}
	}
	WS.Close()
	/*
		fmt.Printf("Block:%v\n", pool_result.Tx_list)
		length := len(pool_result.Tx_list)
		for i := 0; i < length; i++ {
			fmt.Printf("TX in Pool: %s\n", pool_result.Tx_list[i])

		}
	*/
	return pool_result, err
}

func GetTX(tx_hash string) (rpc.GetTransaction_Result, error) {

	hashes := []string{tx_hash}
	var tx_result rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: hashes,
	}

	WS, _, err := websocket.DefaultDialer.Dial("ws://"+settings.GetDaemonConn()+"/ws", nil)

	input_output := rwc.New(WS)
	RPC := jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)
	if err = RPC.CallResult(context.Background(), "DERO.GetTransaction", params, &tx_result); err != nil {
		if LOGGING {
			log.Fatal("Could not obtain TX from daemon, err: %s ", err)
		}
		//WS.Close()
		//return tx_result, err
	}
	WS.Close()

	return tx_result, err
}

/*
func GetBlockByHeight(height int) {

	var rpcClient = getClient()
	var block_result rpc.GetBlock_Result

	params := rpc.GetBlock_Params{
		Height: uint64(height),
	}

	err := rpcClient.CallFor(&block_result, "DERO.GetBlock", params)
	if err != nil || block_result.Status == "" {
		if LOGGING {
			fmt.Printf("Err %s\n", err)
		}
		return
	}
	fmt.Printf("Block:%v\n", block_result)

}
*/
func NameToAddress(name string) (string, error) {

	var name_to_address rpc.NameToAddress_Result

	params := rpc.NameToAddress_Params{
		Name:       name,
		TopoHeight: int64(-1),
	}

	WS, _, err := websocket.DefaultDialer.Dial("ws://"+settings.GetDaemonConn()+"/ws", nil)
	input_output := rwc.New(WS)
	RPC := jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)
	if err = RPC.CallResult(context.Background(), "DERO.NameToAddress", params, &name_to_address); err != nil {
		if LOGGING {
			fmt.Printf("Error getting name to address: %s ", err)
		}
		WS.Close()
		return "", err
	}
	WS.Close()
	return name_to_address.Address, nil
}
