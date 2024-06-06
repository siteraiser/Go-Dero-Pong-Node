package process

import (
	"fmt"
	"hash/crc32"
	"reflect"
	"strconv"
	"strings"
	"time"

	"node/models/daemonapi"
	iaddresses "node/models/iaddresses"
	products "node/models/products"
	walletapi "node/models/walletapi"
	"node/models/webapi"

	_ "github.com/mattn/go-sqlite3"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/google/uuid"
)

const LOGGING = false

var messages []string
var errors []string

//var product_changes = false

var installed_time_utc string

// var start_block int
var last_synced_block int

var next_sync_block = 0
var last_balance_result = 0

//var address_submission_candidates AddressSubmissionCandidates

type AddressSubmissionCandidates struct {
	Entries []rpc.Entry
}

type AddressSubmission struct {
	Type    string
	Crc32   string
	Name    string
	Level1  string
	Level2  string
	City    string
	State   string
	Zip     string
	Country string
}

type Tx struct {
	I_id            int
	Txid            string
	P_type          string
	Buyer_address   string
	Scid            string
	Amount          int
	Respond_amount  int
	Port            int
	For_product_id  int
	For_ia_id       int
	Ia_comment      string
	Product_label   string
	Successful      bool
	Processed       bool
	Block_height    int
	Time_utc        string
	InventoryResult InvUpdateRes
}

//	Expiry          string
//
// returned from inventory update
type InvUpdateRes struct {
	Success bool
	Id_type string
	P       int
	Ia      int
}

type ResponseTx struct {
	R_id             int
	Order_id         int
	Txid             string
	Type             string
	Product_label    string
	Ia_comment       string
	Buyer_address    string
	Amount           int
	Out_scid         string
	Respond_amount   int
	Port             int
	Out_message      string
	Out_message_uuid int
	Uuid             string
	Api_url          string
	Crc32            string
	Ship_address     string
	Confirmed        bool
	Time_utc         string
	Incoming_height  int
	T_block_height   int
}

func ResetMessages() {
	messages = messages[:0]
}
func ResetErrorMessages() {
	errors = errors[:0]
}

func sendCheckIn() {

	balance_result := walletapi.GetBalance()
	balance := 0
	if balance_result > 0 {
		balance = balance_result
	} else {
		errors = append(errors, "Error getting balance or not enough funds")
		return
	}

	//Detect balance change, check later on when tx should've arrived
	height := walletapi.GetHeight()
	if last_balance_result != balance_result {
		next_sync_block = 2 + height
		last_balance_result = balance_result
		return
	}
	if height < next_sync_block {
		return
	}

	//Balance always updates, Txs can get lost tho
	saved_balance := getLastSyncedBalance()
	/*if balance_result == saved_balance {
		return
	}
	*/
	last_synced_block2 := last_synced_block

	transfers_result, err := walletapi.GetAllTransfers(last_synced_block2 + 1)
	if err != nil {
		errors = append(errors, "Wallet connection error. Couldn't get balance. \nEnsure cli wallet or Engram cyberdeck or equivalent is setup \nor logout and log back into wallet.\n")
		return
	}

	/*
		fmt.Println("--------transfers_result count-----------")
		fmt.Printf("transfers_result %v\n", len(transfers_result.Entries))

		fmt.Println("----------------------------------")
		fmt.Printf("balance_result %v\n", balance_result)
		fmt.Printf("saved_balance %v\n", saved_balance)
	*/
	for _, e := range transfers_result.Entries {
		if e.Incoming {
			//Add to saved balance..
			saved_balance = saved_balance + int(e.Amount)
			if LOGGING {
				fmt.Printf("int(e.Amount) %v\n", int(e.Amount))
			}
		} else {
			//Subtract from saved balance..
			saved_balance = saved_balance - int(e.Amount)
			saved_balance = saved_balance - int(e.Fees)
		}
		//Remember the last synced block
		last_synced_block2 = int(e.Height)
	}
	/*
		fmt.Printf("balance %v\n", balance)
		fmt.Printf("saved_balance %v\n", saved_balance)
		fmt.Println("----------------------------------")
	*/
	if saved_balance == balance && saved_balance > 0 {

		//Update sync records
		saveSyncedData(saved_balance, last_synced_block2)
	} else if saved_balance == 0 {
		errors = append(errors, "Error getting synced balance or not enough funds,\n try reloading the page and wallet to complete setup.")

	} else {
		errors = append(errors, "Missing TX, balance is not synced with amount!\nFind a full node and re-install wallet if necessary.")
	}
	if nextCheckInTime() && len(errors) == 0 {

		if LOGGING {
			fmt.Println("sending checkin")
		}
		//Then check-in
		webapi.CheckIn()
	} //else{delist...}
}

func checkTokenBalances() {
	//Reset token balances if difference detected...in case seller sent some etc...
	if len(token_balances) != 0 {
		for scid, balance := range token_balances {
			if balance != walletapi.GetTokenBalance(scid) {
				ResetTokenBalances()
			}
		}
	}
}
func ResetTokenBalances() {
	if LOGGING {
		fmt.Printf("Resetting token balances")
	}
	token_balances = make(map[string]int)
}

func expiringIAs() {
	if iaddresses.HasActiveExpires() == "true" {
		//Check for active iaddresses that are now expired, update db then push to website
		active_expired := iaddresses.GetActiveExpired()
		if len(active_expired) != 0 {
			for _, iaid := range active_expired {
				iaddresses.SetExpiredIAById(iaid)
				webapi.SubmitIAddress(iaddresses.LoadById(iaid))
			}
		}
	}
}

/*********************************/
/* Begin Processing Transactions */
/* Called from main.go loop      */
/*********************************/
func Transactions() ([]string, []string) {
	//fmt.Printf("token balance %v\n", walletapi.GetTokenBalance())

	//reset sync data
	setInstanceVars()
	checkTokenBalances()

	// check if balance and transactions match
	// if so then send a checkin
	// loadincoming()
	// loadinventory()
	// loadorders()
	// loadResponses()
	// loadOut()
	// loadSavedShipping()
	if len(errors) > 0 {
		return messages, errors
	}

	//See if new responses have confirmed
	confirmation()

	// Check incoming transfers for new sales (store in db)
	checkIncoming()

	//Check that everything is in sync and if so save the new state
	sendCheckIn()

	/*
		we no longer want to continue processing
		if there is a missing tx since it could	fail
		to attribute a same block address submission
		being that it checks for a submitted address
		just before sending out the response.

		(
			maybe have another routine to check for same
			block submissions that got missed and then add
			those to the response so that things can contiue
			to process even with a missing tx....
		)

	*/
	if len(errors) > 0 {
		return messages, errors
	}

	/*
		if everything checks out then combine physical goods orders into single orders per buyer
		(for shipping address submissions and single responses for multiple item orders)
	*/
	createOrders()

	// Build the appropriate transactions and send them
	sendTransfers(createTransferList())

	/*
		Finally set IAs as inactive.
		IAs are left active until here and checked above
		if the payment was sent before expiration.
	*/
	expiringIAs()

	return messages, errors
}

func confirmation() {

	/*******************************/
	/** Check if pending response **/
	/** transfers have confirmed  **/
	/*******************************/
	t_block_height := 0

	height_res := walletapi.GetHeight()

	if height_res > 0 {
		t_block_height = height_res - 1
	}

	now_utc := time.Now().UTC()
	time_utc := now_utc.Add(
		-time.Duration(36) * time.Second,
	)

	unConfirmed := unConfirmedResponses()
	if LOGGING {
		fmt.Printf("unConfirmed:%v\n", unConfirmed)
	}
	confirmed_txns := []string{}
	//go through the responses that haven't been confirmed.
	//keep old txids in a csv and check if any of those have confirmed, if so update the txid with the first one that confirms...

	for _, response := range unConfirmed {
		//make sure the response is at least one block old before checking.
		if response["time_utc"].(string) < time_utc.Format("2006-01-02 15:04:05") && response["t_block_height"].(int) < t_block_height {

			check_transaction_result, err := walletapi.GetTransferByTXID(response["txid"].(string))
			//succesfully confirmed
			if !reflect.ValueOf(check_transaction_result.Entry).IsZero() && err == nil {
				if LOGGING {
					fmt.Printf("Confirmed TX: Marking as Confirmed!.... \n%v\n", response["txid"].(string))
				}

				markResAsConfirmed(response["txid"].(string))
				confirmed_txns = append(confirmed_txns, response["txid"].(string))
			} else {

				//not found in wallet yet, check with daemon

				tx_pool_result, tx_pool_err := daemonapi.GetTxPool()
				pool_array := []string{}
				if tx_pool_err == nil {
					pool_array = tx_pool_result.Tx_list
				}

				var txid_found bool = false
				for _, x := range pool_array {
					if x == response["txid"] {
						txid_found = true
						break
					}
				}

				if !txid_found {
					tx_result, getResponseErr := daemonapi.GetTX(response["txid"].(string))

					if len(tx_result.Txs) != 0 {
						if (!tx_result.Txs[0].In_pool && tx_result.Txs[0].ValidBlock == "") || tx_result.Txs[0].Ignored {
							//failed
							markOrderAsPending(response["txid"].(string))
						}
						//If it didn't fail then wait for it to show up in wallet to confirm (do nothing).
					} else if getResponseErr.Error() == "" {
						markOrderAsPending(response["txid"].(string))
					} else {
						errors = append(errors, getResponseErr.Error())
					}
					//If it didn't fail then wait for it to show up in wallet to confirm (do nothing).
				}

				if tx_pool_err != nil {
					errors = append(errors, "Error fetching tx pool \n"+tx_pool_err.Error())
				}
			}
		}
	}
	if LOGGING {
		fmt.Printf("Confirmed TXS:********** \n%v\n", confirmed_txns)
	}
	for _, txid := range confirmed_txns {
		//Get record for freshly confirmed transfer with txid
		confirmed_incoming := getConfirmedInc(txid)
		for _, record := range confirmed_incoming {
			//does not inform when there is an inactive ia since it uses product id 0 and can't find the details...
			messages = append(messages, string(record["type"].(string)+" confirmed for order txid:\n"+record["txid"].(string)))

			//send post message to web api here...
			if record["out_message_uuid"].(int) == 1 && record["type"].(string) == "sale" {
				webapi.SendNewTx(record)
			}
		}
	}

}

/************************/
/******  incoming  ******/
/************************/
func checkIncoming() {
	//Get transfers and save them if they are new and later than the db creation time.
	export_transfers_result, err := walletapi.GetInTransfers(last_synced_block) //
	//fmt.Printf("export_transfers_result:%v\n", export_transfers_result)

	if err != nil {
		errors = append(errors, "Wallet connection error. Couldn't get incoming txns.\nEnsure cyberdeck or equivalent is setup.")
	} else if len(export_transfers_result.Entries) != 0 {
		for _, e := range export_transfers_result.Entries {
			//fmt.Printf("%v\n", reflect.ValueOf(e.Payload_RPC).IsZero())
			// reflect.ValueOf(e.Payload_RPC).IsZero() &&  &&e.Payload_RPC.Has(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress)
			if !reflect.ValueOf(e.Payload_RPC).IsZero() {

				tx, err_str := makeTxObject(e)

				if err_str == "" { //insert TX

					//Is an Integrated Address buy transaction, save it...
					insertNewTransaction(&tx) //and do inventory first...

					//check type of inventory update... product or iaddress
					if tx.InventoryResult.Success {

						//product_changes = true //set changes to true to reload the products (not implemented yet...)

						if tx.InventoryResult.Id_type == "p" {
							webapi.SubmitProduct(products.LoadById(tx.InventoryResult.P), false)
						} else {
							webapi.SubmitIAddress(iaddresses.LoadById(tx.InventoryResult.Ia))
						}

					}
				} else {
					//Likely an address submission since it isn't an integrated address.
					//and if not then save in array to check if it is a later submission to a response.
					if LOGGING {
						fmt.Println("-Address Submission-----")
					}
					address_array := GetAddressArray(e)

					if len(address_array) > 8 { //10 total really... 9 for same block ids?... count includes txid and buyer_wallet in addition to the rest of the fields
						AddressSubmission := GetAddressSubmission(address_array)
						if LOGGING {
							fmt.Printf("AAddressSubmission type:%v\n", AddressSubmission.Type)
						}
						if AddressSubmission.Type == "block" {
							//Not implemented yet....
							//storeAddress

							if storeSameBlockAddress(address_array, int(e.Height)) {
								messages = append(messages, "Shipping address submitted with order.")
							}
							//Save to incoming address table for later addition to response... when genereated (if successful...)
							//(address_tx, wallet_address, block)??
						} else if AddressSubmission.Type == "crc32" {
							if saveAddress(address_array) {
								messages = append(messages, "Shipping address submitted by buyer.")
							}
						}

					}

				}
			}
		}

		//Add failed transactions back into incoming table... because things don't work as they should
		addFailedTransactionsBackIntoIncomingTable()

	}

}

func makeTxObject(entry rpc.Entry) (Tx, string) {

	var tx Tx
	tx.Txid = entry.TXID
	tx.Amount = int(entry.Amount)
	tx.Block_height = int(entry.Height)

	//time_layout := "2022-02-03T17:51:16.006+01:00"
	//tm, _ := time.Parse(time_layout, )
	tm := entry.Time.UTC()
	tx.Time_utc = tm.Format("2006-01-02 15:04:05")

	has_reply_back_addr := false
	if entry.Payload_RPC.Has(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress) {
		tx.Buyer_address = entry.Payload_RPC.Value(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress).(rpc.Address).String()
		has_reply_back_addr = true
	}
	if entry.Payload_RPC.Has(rpc.RPC_DESTINATION_PORT, rpc.DataUint64) {

		//	fmt.Printf("tx.Port:\n%v", strconv.FormatUint((entry.Payload_RPC.Value(rpc.RPC_DESTINATION_PORT, rpc.DataUint64).(uint64)), 10))
		tx.Port = int(entry.Payload_RPC.Value(rpc.RPC_DESTINATION_PORT, rpc.DataUint64).(uint64))
	}
	if !has_reply_back_addr {
		return tx, "error"
	}

	tx.For_product_id = 0
	tx.Product_label = "Inactive I.A."
	tx.Ia_comment = "Inactive I.A."

	ia_settings := getIASettings(tx.Amount, tx.Port)

	expired := false

	if tx.Time_utc > ia_settings.Expiry && ia_settings.Expiry != "" {
		expired = true
	}

	if ia_settings != (IA_settings{}) && !expired { //!reflect.ValueOf(ia_settings).IsZero()
		if LOGGING {
			fmt.Println("Found active I. Address Settings for incoming transaction")
		}
		tx.For_product_id = ia_settings.P_id
		tx.Product_label = ia_settings.P_label
		tx.Ia_comment = ia_settings.Ia_comment

		//tx.Expiry = ia_settings.Expiry not being used
		//token settings...
		product := products.LoadById(ia_settings.P_id)
		tx.P_type = product.P_type
		tx.Scid = product.Scid
		tx.Respond_amount = product.Respond_amount

		//Use Integrated Address respond amount if defined.
		if ia_settings.Ia_respond_amount > 0 {
			tx.Respond_amount = ia_settings.Ia_respond_amount
		}

		//Use Integrated Address scid if defined.
		if ia_settings.Ia_scid != "" {
			tx.Scid = ia_settings.Ia_scid
		}

	}

	return tx, ""

}

/* Address Submission Stuff */
func GetAddressArray(entry rpc.Entry) map[string]string {

	address_string := ""
	address_array := make(map[string]string)
	//time_layout := "2006-01-02 15:04:05"
	//time_utc, _ := time.Parse(time_layout, )
	if entry.Payload_RPC.Has(rpc.RPC_COMMENT, rpc.DataString) {
		address_string = entry.Payload_RPC.Value(rpc.RPC_COMMENT, rpc.DataString).(string)
	}

	if address_string == "" || entry.Time.UTC().Format("2006-01-02 15:04:05") < installed_time_utc {
		return address_array
	}

	//address_string = "id$108160166?n$First Last?l1$555 Some Road?l2$?c1$The Town / City?s$NY?z$12345?c2$US"
	sections := strings.Split(address_string, "?")
	for _, part := range sections {
		before, after, _ := strings.Cut(part, "$")
		address_array[before] = after
	}
	address_array["txid"] = entry.TXID

	if entry.Payload_RPC.Has(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress) {
		address_array["buyer_address"] = entry.Payload_RPC.Value(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress).(rpc.Address).String()

	}

	address_array["buyer_address"] = entry.Sender
	if LOGGING {
		fmt.Printf("SENDER2:\n%v", address_array)
	}
	return address_array
}

/* Address Submission Stuff */
func GetAddressSubmission(address_array map[string]string) AddressSubmission {
	var name, value string

	if _, found := address_array["id"]; found {
		name = "crc32"
		value = address_array["id"]
	} else {
		name = "block"
		value = ""
	}
	var addressSubmission AddressSubmission = AddressSubmission{
		Name:    address_array["n"],
		Level1:  address_array["l1"],
		Level2:  address_array["l2"],
		City:    address_array["c1"],
		State:   address_array["s"],
		Zip:     address_array["z"],
		Country: address_array["c2"],
		Type:    name,
		Crc32:   value,
	}

	return addressSubmission
}

/**********************/
/******  ORDERS  ******/
/**********************/
func createOrders() {
	//Make array of unprocessed transactions

	not_processed := unprocessedTxs()

	if LOGGING {
		fmt.Printf("Creating Orders, not_processed ORDERS:\n%v\n", not_processed)
		fmt.Println("---------------------------------")
	}
	/********************************/
	/** Create Orders from new txs **/
	/********************************/

	var tx_list = make(map[string][]Tx)

	for _, tx := range not_processed {
		settings := getIASettings(tx.Amount, tx.Port)

		//Ensure it was a successful incoming transaction.

		if tx.Successful {
			//Was found and had enough inventory.$settings['scid'] == '' && $settings['ia_scid'] == ''

			// Categorize transactions based on product type
			switch settings.P_type {
			case "physical":
				tx_list["physical_sales"] = append(
					tx_list["physical_sales"],
					tx,
				)
			case "digital":
				tx_list["digital_sales"] = append(
					tx_list["digital_sales"],
					tx,
				)
			case "token":
				tx_list["token_sales"] = append(
					tx_list["token_sales"],
					tx,
				)
			}
			/*
				else if($settings['p_type'] == 'smartcontract'){
					$tx_list['sc_sales'][] = $tx;
				}
			*/
		} else if settings != (IA_settings{}) {
			//No inventory$settings['scid'] == '' && $settings['ia_scid'] == ''

			if settings.P_type == "physical" || settings.P_type == "digital" || settings.P_type == "token" {
				tx_list["refunds"] = append(tx_list["refunds"], tx)
			}
			/*
				else if($settings['p_type'] == 'token'){
					$tx_list['token_refunds'][] = $tx;
				}else if($settings['p_type'] == 'smartcontract'){
					$tx_list['sc_refunds'][] = $tx;
				}
			*/
		} else {
			//No mathcing products / I. Addresses found
			tx_list["refunds"] = append(tx_list["refunds"], tx)
		}
	}

	if len(tx_list) != 0 && LOGGING {
		fmt.Printf("Combining / sorting Orders, tx_list:\n%v\n", tx_list)
		fmt.Println("---------------------------------")
	}
	//Combine orders from same wallet and block
	if physical_sales, found := tx_list["physical_sales"]; found {
		var (
			heights = make(map[int][]Tx)
			blocks  = make(map[int]map[string][]Tx)
		)

		for _, tx := range physical_sales {
			heights[tx.Block_height] = append(heights[tx.Block_height], tx)
		}

		for height, tx_array := range heights {

			//	txObj := reflect.VisibleFields(reflect.TypeOf(Tx{}))
			blocks[height] = make(map[string][]Tx)

			for _, tx := range tx_array {
				blocks[height][tx.Buyer_address] = append(blocks[height][tx.Buyer_address], tx)
			}
		}

		//	$orders = [];
		for _, addresses := range blocks {
			for _, tx_array := range addresses {
				insertOrder(tx_array, "physical_sale")
			}
		}
	}
	// Define the order types and their corresponding transaction types
	var orderTypes = []struct {
		txType    string
		orderType string
	}{
		{"digital_sales", "digital_sale"},
		{"token_sales", "token_sale"},
		{"refunds", "refund"},
	}

	// Create sales and refund orders for each type
	for _, order := range orderTypes {
		createSeparateOrders(tx_list, order.txType, order.orderType)
	}
	if LOGGING {
		fmt.Println("Done Inserting Orders")
	}
}

func createSeparateOrders(tx_list map[string][]Tx, txType, orderType string) {
	if transactions, found := tx_list[txType]; found {
		var orders []Tx
		for _, tx := range transactions {
			orders = append(orders, tx)
			insertOrder(orders, orderType)
		}
	}
}

func createTransferList() ([]rpc.Transfer, []ResponseTx) {
	//Find pending orders and create a transfer list.
	var (
		transfer_list     []rpc.Transfer
		pending_orders    []ResponseTx
		transfer          rpc.Transfer
		updatedResponseTx ResponseTx
	)
	order_types := []string{"physical_sale", "digital_sale", "token_sale", "refund"}
	for _, t := range order_types {
		pending_orders = append(pending_orders, getOrdersByStatusAndType("pending", t)...)
	}

	for i, responseTx := range pending_orders {
		settings := getIASettings(responseTx.Amount, responseTx.Port)
		switch responseTx.Type {
		case "physical_sale", "digital_sale":
			updatedResponseTx, transfer = createTransfer(responseTx, settings)
		case "token_sale":
			updatedResponseTx, transfer = createTokenTransfer(responseTx, settings)
		case "refund":
			updatedResponseTx, transfer = createRefundTransfer(responseTx, settings)
		}
		transfer_list = append(transfer_list, transfer)
		pending_orders[i] = updatedResponseTx
	}
	if len(pending_orders) != 0 && LOGGING {
		fmt.Printf("PENDING ORDERS:\n%v\n", pending_orders[0].Type)
		fmt.Println("---------------------------------")
		fmt.Printf("transfer_list:\n%v\n", transfer_list)
		fmt.Println("---------------------------------")
	}
	return transfer_list, pending_orders
}

func sendTransfers(transfer_list []rpc.Transfer, pending_orders []ResponseTx) {
	/***************************************/
	/** Combine Regular Product Transfers **/
	/***************************************/

	var responseTXID,
		payload_result,
		err string // string defaults to ""

	var t_block_height int
	/*
		Does combined transfers, scid transfers may require separate transfers in case of refund required...
	*/
	if len(transfer_list) != 0 {

		//Make sure wallet is working
		height := walletapi.GetHeight()
		if height > 0 {
			t_block_height = height
		}

		fmt.Println("First T_block_height:" + strconv.Itoa(t_block_height))

		if t_block_height > 0 {
			//try the transfer
			payload_result, err = walletapi.Transfer(transfer_list)
			//Get the actual blockheight or just increment by 1 if it fails since we need to have a height to check for confirmation
			height_result := walletapi.GetHeight()
			if height_result > 0 {
				t_block_height = height_result
			} else {
				t_block_height += 1
			}
		}

		fmt.Println("Second T_block_height:" + strconv.Itoa(t_block_height))

		if payload_result != "" && err == "" {
			responseTXID = payload_result
		} else {
			if err != "" {
				errors = append(errors, "Error: "+err)
			} else {
				errors = append(errors, "Unkown Transfer Error")
			}
		}

		//....
		if len(errors) == 0 && responseTXID != "" {
			for _, tx := range pending_orders {

				//Save as time to use for waiting for a confirmation check
				now_utc := time.Now().UTC()
				time_utc := now_utc.Format("2006-01-02 15:04:05")

				//Mark incoming transaction as processed.
				//In the next check cycle it can be set to unprocessed above if response is not confirmed, then it is reprocessed.
				//Inventory is done once when it is first inserted.
				result := markOrderAsProcessed(tx.Order_id)

				if result {

					//Find same block shipping addresses add to the responsetx struct and then delete the record at some point, immediately for now...
					var shipping_record_id int
					if tx.Type == "sale" { //if type is a sale (OR not token_sale)... check for stored address with same block and wallet.
						tx.Ship_address, shipping_record_id = getSameBlockShipping(tx.Buyer_address, tx.Incoming_height)
					}

					//Add the new values
					tx.Txid = responseTXID
					tx.Time_utc = time_utc
					tx.T_block_height = t_block_height

					if saveResponse(tx) && shipping_record_id != 0 {
						deleteShippingRecord(shipping_record_id)
					}
					var message_part string
					if tx.Type == "sale" {
						detail_set := getOrderDetails(tx.Order_id)
						for _, details := range detail_set {
							message_part += details["product_label"] + " " + details["ia_comment"] + ", "
						}
						message_part = strings.Trim(message_part, ", ")
					} else {
						message_part = tx.Product_label //+ " " + details.Ia_comment
					}

					messages = append(messages, tx.Txid+"\nResponse initiated \n"+message_part)
				}
			}
		}
	}
}

// Regular response (message w/ Dero) physical or digital
func createTransfer(rtx ResponseTx, settings IA_settings) (ResponseTx, rpc.Transfer) {

	checkIASettings := func(settings IA_settings) (amount int, unique_identifier, comment string) {

		//Use Integrated Address respond amount if defined.
		amount = settings.Respond_amount
		if settings.Ia_respond_amount > 0 {
			amount = settings.Ia_respond_amount
		}

		unique_identifier = ""
		//See if use uuid is selected, generate one if so.
		if settings.Out_message_uuid == 1 {
			unique_identifier = uuid.New().String()
			settings.Out_message = settings.Out_message + unique_identifier
		}

		//Use original out message if not a uuid (usually a link or some text)...
		comment = settings.Out_message

		//Check for a pending response for this incoming tx
		pending_response := checkForPendingResponseByOrderId(rtx.Order_id)
		if len(pending_response) != 0 {
			// Found a previous repsonse, use that instead of a new one
			// 	(in case of double response we want the same confirmation number for address submission)
			comment = pending_response["out_message"]
			unique_identifier = pending_response["uuid"]
		}

		return
	}

	amount, unique_identifier, comment := checkIASettings(settings)

	//Send Response to buyer
	var transfer rpc.Transfer = rpc.Transfer{
		// SCID: 0000000000000000000000000000000000000000000000000000000000000000,
		Destination: rtx.Buyer_address,
		Amount:      uint64(amount),
		Payload_RPC: rpc.Arguments{
			{
				Name:     rpc.RPC_COMMENT,
				DataType: rpc.DataString,
				Value:    comment,
			},
		},
	}

	// $transfer_object=(object)$transfer;
	// update unprocessed array
	// $tx['ia_comment'] = $settings['ia_comment'];
	// rtx.Out_scid = transfer.SCID
	var value string
	if unique_identifier == "" {
		value = "1"
	} else {
		table := crc32.MakeTable(crc32.IEEE) //0
		value = strconv.Itoa(int(crc32.Checksum([]byte(unique_identifier), table)))
	}

	rtx = ResponseTx{
		Respond_amount:   int(transfer.Amount),
		Out_message:      comment,
		Out_message_uuid: settings.Out_message_uuid,
		Uuid:             unique_identifier,
		Api_url:          settings.Api_url,
		Type:             "sale",
		Crc32:            value,
	}

	return rtx, transfer
}

// Token Response
func createTokenTransfer(rtx ResponseTx, settings IA_settings) (ResponseTx, rpc.Transfer) {

	useIASettings := func(settings IA_settings) (int, string, string) {
		//Use Integrated Address respond amount if defined.
		respond_amount := settings.Respond_amount
		if settings.Ia_respond_amount > 0 {
			respond_amount = settings.Ia_respond_amount
		}

		//Use Integrated Address scid if defined.
		out_scid := settings.Scid
		if settings.Ia_scid != "" {
			out_scid = settings.Ia_scid
		}

		transfer_out_message := settings.Out_message

		//Use scid as out message if message is null. (not likely going to be seen anyway lol)
		if settings.Out_message == "" {
			transfer_out_message = out_scid
		}

		return respond_amount, out_scid, transfer_out_message
	}

	amount, scid, comment := useIASettings(settings)

	if LOGGING {
		fmt.Println("scid string:" + scid)
	}

	//Send Response to buyer
	var transfer rpc.Transfer = rpc.Transfer{
		Destination: rtx.Buyer_address,
		Amount:      uint64(amount),
		SCID:        crypto.HashHexToHash(scid),
		Payload_RPC: rpc.Arguments{
			{
				Name:     rpc.RPC_COMMENT,
				DataType: rpc.DataString,
				Value:    comment,
			},
		},
	}

	//	fmt.Printf("scid:%v\n", transfer.SCID)
	rtx = ResponseTx{
		Respond_amount:   int(transfer.Amount),
		Out_message:      comment,
		Out_message_uuid: 0,
		Uuid:             "",
		Api_url:          "",
		Out_scid:         scid,
		Crc32:            "",
		Type:             "token_sale",
	}
	return rtx, transfer
}

// Regular response (message w/ Dero)
func createRefundTransfer(rtx ResponseTx, settings IA_settings) (ResponseTx, rpc.Transfer) {

	//Send Response to buyer
	//fmt.Printf("Destination:", rtx.Buyer_address)
	transfer_out_message := "Refund for: " + settings.P_label + "-" + settings.Ia_comment
	if len(transfer_out_message) > 110 {
		transfer_out_message = transfer_out_message[0:110]
	}

	var transfer rpc.Transfer = rpc.Transfer{
		Amount:      uint64(rtx.Amount),
		Destination: rtx.Buyer_address,
		Payload_RPC: rpc.Arguments{
			{
				Name:     rpc.RPC_COMMENT,
				DataType: rpc.DataString,
				Value:    transfer_out_message,
			},
		},
	}

	rtx = ResponseTx{
		Respond_amount:   int(transfer.Amount),
		Out_message:      transfer_out_message,
		Out_message_uuid: 0,
		Uuid:             "",
		Api_url:          "",
		Crc32:            "", //should actually be null to match the php version...
		Type:             "refund",
		//rtx.Out_scid = transfer.SCID
	}

	return rtx, transfer
}
