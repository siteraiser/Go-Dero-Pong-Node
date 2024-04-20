package daemonapi

//import "os"
//import "fmt"
//import "time"

import (
	"fmt"
	settings "node/models/settings"

	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc"
)

const LOGGING = false

func getClient() jsonrpc.RPCClient {
	return jsonrpc.NewClient("http://" + settings.GetWalletConn().Api + "/json_rpc")
}

func GetTxPool() (rpc.GetTxPool_Result, error) {
	var rpcClient = getClient()
	var pool_result rpc.GetTxPool_Result

	err := rpcClient.CallFor(&pool_result, "DERO.GetTxPool")
	if err != nil || pool_result.Status == "" {
		if LOGGING {
			fmt.Printf("Could not get tx pool %s\n", err)
		}
		return pool_result, err
	}

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
	var rpcClient = getClient()
	hashes := []string{tx_hash}
	var tx_result rpc.GetTransaction_Result
	params := rpc.GetTransaction_Params{
		Tx_Hashes: hashes,
	}

	err := rpcClient.CallFor(&tx_result, "DERO.GetTransaction", params)
	if err != nil || tx_result.Status == "" {
		if LOGGING {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
		}
		return tx_result, err
	}
	return tx_result, err
}

func GetBlockByHeight(height int) {

	var rpcClient = getClient()
	var block_result rpc.GetBlock_Result

	params := rpc.GetBlock_Params{
		Height: uint64(height),
	}

	err := rpcClient.CallFor(&block_result, "DERO.GetBlock", params)
	if err != nil || block_result.Status == "" {
		if LOGGING {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
		}
		return
	}
	fmt.Printf("Block:%v\n", block_result)

}

func NameToAddress(name string) (string, error) {

	var rpcClient = getClient()
	var name_to_address rpc.NameToAddress_Result

	params := rpc.NameToAddress_Params{
		Name:       name,
		TopoHeight: int64(-1),
	}

	err := rpcClient.CallFor(&name_to_address, "DERO.NameToAddress", params)
	if err != nil || name_to_address.Status == "" {
		if LOGGING {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
		}
		return "", err
	}
	fmt.Printf("Block:%v\n", name_to_address.Address)
	return name_to_address.Address, nil
}
