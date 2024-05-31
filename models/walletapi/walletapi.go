package walletapi

import (
	"encoding/base64"
	"fmt"
	settings "node/models/settings"
	"time"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/ybbus/jsonrpc"
)

const LOGGING = false

func getClient() jsonrpc.RPCClient {
	opts := &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(settings.GetWalletConn().User+":"+settings.GetWalletConn().Pass)),
		},
	}
	return jsonrpc.NewClientWithOpts("http://"+settings.GetWalletConn().Api+"/json_rpc", opts)
}

func GetAddress() string {

	rpcClient := getClient()

	var addr_result rpc.GetAddress_Result
	err := rpcClient.CallFor(&addr_result, "GetAddress")
	if err != nil || addr_result.Address == "" {
		if LOGGING {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
		}
		return ""
	}
	return addr_result.Address
	// service address can be created client side for now
}

func GetBalance() int {

	rpcClient := getClient()

	var bal_result rpc.GetBalance_Result
	err := rpcClient.CallFor(&bal_result, "GetBalance")
	if err != nil {

		if LOGGING {
			fmt.Printf("Could not obtain balance from wallet err %s\n", err)
		}
		return -1
	}
	//	fmt.Printf("Could not obtain balance from wallet err %v\n", bal_result.Balance)

	return int(bal_result.Balance)
}

/*
func GetUnlockedBalance() int {

		rpcClient := getClient()

		var bal_result rpc.GetBalance_Result
		err := rpcClient.CallFor(&bal_result, "GetBalance")
		if err != nil {
			fmt.Printf("Could not obtain balance from wallet err %s\n", err)
			return -1
		}
		//	fmt.Printf("Could not obtain balance from wallet err %v\n", bal_result.Balance)

		return int(bal_result.Unlocked_Balance)
	}
*/

func GetTokenBalance(scid string) int {
	//"dd29bf592a53d8af04f4a7877b9693c9fdca78df1e1edaa22926fa8111f43bae"
	rpcClient := getClient()

	var bal_result rpc.GetBalance_Result
	err := rpcClient.CallFor(&bal_result, "GetBalance", rpc.GetBalance_Params{SCID: crypto.HashHexToHash(scid)})
	if err != nil {
		if LOGGING {
			fmt.Printf("Could not obtain balance from wallet err %s\n", err)
		}
		return -1
	}
	//	fmt.Printf("Could not obtain balance from wallet err %v\n", bal_result.Balance)

	return int(bal_result.Unlocked_Balance)
}

func GetHeight() int {
	rpcClient := getClient()

	var height_result rpc.GetHeight_Result
	err := rpcClient.CallFor(&height_result, "GetHeight")
	if err != nil {
		if LOGGING {
			fmt.Printf("Could not obtain balance from wallet err %s\n", err)
		}
		return -1
	}
	return int(height_result.Height)
}

func MakeIntegratedAddress(d_port int, in_message string, ask_amount int, expiry string) (string, error) {
	rpcClient := getClient()
	//fmt.Printf("Expiry %s\n", expiry)
	expected_arguments := rpc.Arguments{
		{
			Name:     rpc.RPC_DESTINATION_PORT,
			DataType: rpc.DataUint64,
			Value:    uint64(d_port),
		},
		{
			Name:     rpc.RPC_COMMENT,
			DataType: rpc.DataString,
			Value:    in_message,
		},
		{
			// this service will reply to incoming request,
			Name:     rpc.RPC_NEEDS_REPLYBACK_ADDRESS,
			DataType: rpc.DataUint64,
			Value:    uint64(0),
		},
		{
			Name:     rpc.RPC_VALUE_TRANSFER,
			DataType: rpc.DataUint64,
			Value:    uint64(ask_amount), // in atomic units
		},
	}

	if expiry != "" {
		exp, _ := time.Parse("2006-01-02 15:04:05", expiry)

		expected_arguments = append(expected_arguments, rpc.Argument{
			Name:     rpc.RPC_EXPIRY,
			DataType: rpc.DataTime,
			Value:    exp.UTC(),
		})
	}

	//fmt.Printf("expected_arguments %s\n", expected_arguments)

	var addr *rpc.Address
	var addr_result rpc.GetAddress_Result
	err := rpcClient.CallFor(&addr_result, "GetAddress")
	if err != nil || addr_result.Address == "" {
		if LOGGING {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
		}
		return "", err
	}

	if addr, err = rpc.NewAddress(addr_result.Address); err != nil {
		if LOGGING {
			fmt.Printf("address could not be parsed: addr:%s err:%s\n", addr_result.Address, err)
		}
		return "", err
	}

	/*	fmt.Printf("Wallet Address: %s\n", addr)
		service_address_without_amount := addr.Clone()
		service_address_without_amount.Arguments = expected_arguments[:len(expected_arguments)-1]
		fmt.Printf("Integrated address to activate '%s', (without hardcoded amount) service: \n%s\n", "PONG NODE", service_address_without_amount.String())
	*/
	// service address can be created client side for now
	service_address := addr.Clone()
	service_address.Arguments = expected_arguments
	if LOGGING {
		fmt.Printf("Integrated address to activate '%s', service: \n%s\n", "PONG NODE", service_address.String())
	}
	return service_address.String(), err
}

func GetAllTransfers(min_height int) (rpc.Get_Transfers_Result, error) {
	rpcClient := getClient()
	var transfers rpc.Get_Transfers_Result
	err := rpcClient.CallFor(&transfers, "GetTransfers", rpc.Get_Transfers_Params{In: true, Out: true, Min_Height: uint64(min_height)})
	if err != nil && LOGGING {
		fmt.Printf("Could not obtain gettransfers from wallet: %s", err)
	}
	//fmt.Printf("Transfers %s", transfers)
	return transfers, err
}

func GetInTransfers(min_height int) (rpc.Get_Transfers_Result, error) {
	rpcClient := getClient()
	var transfers rpc.Get_Transfers_Result
	err := rpcClient.CallFor(&transfers, "GetTransfers", rpc.Get_Transfers_Params{In: true, Out: false, Coinbase: false, Min_Height: uint64(min_height)})
	if err != nil && LOGGING {
		fmt.Printf("Could not obtain gettransfers from wallet: %s", err)
	}
	//fmt.Printf("Transfers %s", transfers)
	return transfers, err
}

func Transfer(Txs []rpc.Transfer) (string, string) {
	/*var response = rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
		{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: DEST_PORT},
		{Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: "Successfully purchased pong (this could be serial/license key or download link or further)"},
	}*/
	rpcClient := getClient()
	var result rpc.Transfer_Result
	tparams := rpc.Transfer_Params{Transfers: Txs}
	err := rpcClient.CallFor(&result, "Transfer", tparams)
	if err != nil {
		if LOGGING {
			fmt.Printf("err while transfer: %s", err)
		}
		return "", err.Error()
	}
	if LOGGING {
		fmt.Printf("Transfer Successful: %s", result.TXID)
	}
	return result.TXID, ""

}

func GetTransferByTXID(txid string) (rpc.Get_Transfer_By_TXID_Result, error) {
	rpcClient := getClient()
	var transfer rpc.Get_Transfer_By_TXID_Result
	err := rpcClient.CallFor(&transfer, "GetTransferbyTXID", rpc.Get_Transfer_By_TXID_Params{TXID: txid})
	if err != nil && LOGGING {
		fmt.Printf("Could not obtain transfer from wallet: %s", err)
	}
	//fmt.Printf("Transfers %s", transfers)
	return transfer, err
}
