package process

import (
	"database/sql"
	"fmt"
	"log"
	"node/crypt"
	walletapi "node/models/walletapi"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type IA_settings struct {
	Id                int
	P_id              int
	P_type            string
	P_label           string
	Scid              string
	Ia_scid           string
	Respond_amount    int
	Ia_respond_amount int
	Out_message_uuid  int
	Out_message       string
	Api_url           string
	Ia_comment        string
	Status            bool
	Expiry            string
}

var token_balances map[string]int

func setInstanceVars() {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, _ := db.Query("SELECT name,value FROM settings WHERE name = 'install_time_utc' OR name = 'start_block' OR name = 'last_synced_block'")
	var (
		name  string
		value string
	)

	for rows.Next() {
		rows.Scan(&name, &value)
		switch name {
		case "install_time_utc":
			installed_time_utc = value
		//case "start_block":
		//	start_block, _ = strconv.Atoi(value)
		case "last_synced_block":
			last_synced_block, _ = strconv.Atoi(value)
		}
	}
}

func getLastSyncedBalance() int {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		value int
	)
	db.QueryRow("SELECT value FROM settings WHERE name = 'last_synced_balance'").Scan(&value)

	return value
}

func saveSyncedData(saved_balance int, last_synced_block int) {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE settings SET value = ? WHERE name = 'last_synced_block'")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(
		last_synced_block,
	)
	if err != nil {
		log.Fatal(err)
	}

	statement, err = db.Prepare("UPDATE settings SET value = ? WHERE name = 'last_synced_balance'")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(
		saved_balance,
	)
	if err != nil {
		log.Fatal(err)
	}
	//	rows_affected, _ := result.RowsAffected()
	//fmt.Println("update?:", rows_affected)
}

func nextCheckInTime() bool {
	now := time.Now().UTC()

	//fmt.Println("Now:", now.Format("2006-01-02 15:04:05"))

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		value string
	)
	err = db.QueryRow("SELECT value FROM settings WHERE (name = 'next_checkin_utc' AND value < ?)", now.Format("2006-01-02 15:04:05")).Scan(&value)
	switch {
	case err != nil:
		//fmt.Println("BEFORE:", now.Format("2006-01-02 15:04:05"))
		return false
	default:
		// it is later than next checkin time, set next check-in time and return true
		//	fmt.Println("AFTER:", now.Format("2006-01-02 15:04:05"))
		next_checkin := now.Add(
			+time.Duration(5) * time.Minute,
		)

		db, err = sql.Open("sqlite3", "./pong.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		statement, err := db.Prepare("UPDATE settings SET value=? WHERE name = 'next_checkin_utc'")
		if err != nil {
			log.Fatal(err)
		}

		_, err = statement.Exec(
			next_checkin.Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			log.Fatal(err)
		}

		//see if registered, if not return false.
		var (
			count string
		)
		err = db.QueryRow("SELECT COUNT(*) FROM settings WHERE name = 'web_api_id' AND NOT(value IS NULL OR value = '')").Scan(&count)
		switch {
		case err != nil:
			return false
		case count == "0":
			return false
		default:
			return true
		}

	}

}

/** TX Processing **/

func getIASettings(amount int, port int) IA_settings {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if LOGGING {
		fmt.Println("getting ia settings..")
		fmt.Println(amount)
		fmt.Println(port)
	}
	var IA_settings IA_settings
	rows, err := db.Query("SELECT ia_id, p_id, p_type, label,scid, ia_scid, respond_amount, ia_respond_amount, out_message_uuid, out_message, api_url, iaddresses.comment AS ia_comment,expiry FROM iaddresses "+
		"INNER JOIN products ON iaddresses.product_id = products.p_id  "+
		"WHERE iaddresses.ask_amount = ? AND iaddresses.port = ? AND iaddresses.status = '1'", amount, port)
	if err != nil {
		return IA_settings
	}
	var (
		ia_id             int
		p_id              int
		p_type            string
		p_label           string
		scid              string
		ia_scid           string
		respond_amount    int
		ia_respond_amount int
		out_message_uuid  int
		out_message       string
		api_url           string
		ia_comment        string
		expiry            string
	)

	for rows.Next() {
		rows.Scan(&ia_id, &p_id, &p_type, &p_label, &scid, &ia_scid, &respond_amount, &ia_respond_amount, &out_message_uuid, &out_message, &api_url, &ia_comment, &expiry)

		IA_settings.Id = ia_id
		IA_settings.P_id = p_id
		IA_settings.P_type = p_type
		IA_settings.P_label = p_label
		IA_settings.Scid = scid
		IA_settings.Ia_scid = ia_scid
		IA_settings.Respond_amount = respond_amount
		IA_settings.Ia_respond_amount = ia_respond_amount
		IA_settings.Out_message_uuid = out_message_uuid
		IA_settings.Out_message = out_message
		IA_settings.Api_url = api_url
		IA_settings.Ia_comment = ia_comment
		IA_settings.Status = true
		IA_settings.Expiry = expiry
	}

	return IA_settings

}

/*
func updateInventory(tx Tx) InvUpdateRes {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		p_and_ia_ids InvUpdateRes
		ia_id        int
		p_id         int
		inventory    int
		ia_inventory int
	)
	p_and_ia_ids.Success = false

	err = db.QueryRow(
		"SELECT ia_id,p_id,inventory,ia_inventory FROM iaddresses "+
			"INNER JOIN products ON (iaddresses.product_id = products.p_id) "+
			"WHERE iaddresses.port = ? AND iaddresses.ask_amount = ? AND iaddresses.status = '1'",
		tx.Port,
		tx.Amount,
	).Scan(&ia_id, &p_id, &inventory, &ia_inventory)

	switch {
	case err != nil:
		if LOGGING {
			fmt.Println("error getting update inventory")
		}
		return p_and_ia_ids
	}

	id := 0
	id_type := ""
	query := ""
	if inventory > 0 {
		query = "UPDATE products SET inventory = inventory - 1 WHERE p_id = ?"
		id = p_id
		id_type = "p"
	} else if ia_inventory > 0 {
		query = "UPDATE iaddresses SET ia_inventory = ia_inventory - 1	WHERE ia_id = ?"
		id = ia_id
		id_type = "ia"
	}
	if id != 0 { //Still some inventory
		statement, err := db.Prepare(query)
		if err != nil {
			fmt.Println("updating inventory prepare err-")
			log.Fatal(err)
		}
		_, err = statement.Exec(
			id,
		)
		if err != nil {
			fmt.Println("-updating inventory err-")
			log.Fatal(err)
		}
		p_and_ia_ids.Success = true
		p_and_ia_ids.Id_type = id_type
		p_and_ia_ids.P = p_id
		p_and_ia_ids.Ia = ia_id
		return p_and_ia_ids
	}
	p_and_ia_ids.Success = false
	fmt.Printf("In update.... p_and_ia_ids.Success-------------:%v", p_and_ia_ids.Success)
	return p_and_ia_ids

}
*/
//Some spaghetti for Secret Name Basis :D
func insertNewTransaction(tx *Tx) {
	//Should refactor back into the different functions as they were before the inserts started acting up...
	tx.InventoryResult.Success = false

	//see if tx exists..
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	exists := false
	var (
		count string
		//count2 string
	)
	err = db.QueryRow("SELECT COUNT(*) FROM incoming WHERE txid = ?", tx.Txid).Scan(&count)
	//_ = db.QueryRow("SELECT COUNT(*) FROM failed_incoming WHERE txid = ?", tx.Txid).Scan(&count2) //just to be safe lol technically it should be fine without this since failed transactions shoud be in the incoming table by the next go around
	switch {
	case err != nil:
		log.Fatal(err)
	case count == "1":
		exists = true
	}

	if exists || tx.Time_utc < installed_time_utc {
		return
	}
	/*	Continue checking if the transaction is a success... */
	//Is the Integrated Address expired?
	/*nah...	expired := false
	if tx.Time_utc > tx.Expiry && tx.Expiry != "" {
		expired = true
	}
	*/

	//Does seller have enough tokens?
	failed_token := false
	//Check supply, and keep a running total
	if tx.P_type == "token" {
		if _, found := token_balances[tx.Scid]; !found {
			token_balances[tx.Scid] = walletapi.GetTokenBalance(tx.Scid)
			//To do.. if token_balances[tx.Scid] <= 0 set status false... if it is an integrated address
		}
		//	fmt.Printf("\n\ntx.Scid: %v -- balance %v", tx.Scid, token_balances[tx.Scid])
		//	fmt.Printf("\n\ntx.Respond_amount %v", tx.Respond_amount)

		if tx.Respond_amount > token_balances[tx.Scid] {
			//maybe set inv to zero and send to webapi (leave status for now since it may be required to lookup the info for the response...)
			tx.InventoryResult.Success = false
			failed_token = true
		}
		if tx.Respond_amount <= token_balances[tx.Scid] {
			token_balances[tx.Scid] = token_balances[tx.Scid] - tx.Respond_amount
		}
	}

	//update inventory if possible
	var (
		ia_id        int
		p_id         int
		inventory    int
		ia_inventory int
	)

	err = db.QueryRow(
		"SELECT ia_id,p_id,inventory,ia_inventory FROM iaddresses "+
			"INNER JOIN products ON (iaddresses.product_id = products.p_id) "+
			"WHERE iaddresses.port = ? AND iaddresses.ask_amount = ? AND iaddresses.status = '1'",
		tx.Port,
		tx.Amount,
	).Scan(&ia_id, &p_id, &inventory, &ia_inventory)

	no_record := false
	switch {
	case err != nil:
		if LOGGING {
			fmt.Println("error getting update inventory")
		}
		no_record = true
		//return
	}
	err = nil
	id := 0
	id_type := ""
	if !failed_token && no_record { //inventory done if not enough irl tokens and ia is not expired.. && !expired
		if inventory > 0 {
			id = p_id
			id_type = "p"
		} else if ia_inventory > 0 {
			id = ia_id
			id_type = "ia"
		}
		if id != 0 { //Still some inventory
			if id_type == "p" {
				//fmt.Printf("Reducing Product inventory for product id:%v, currently inv is: %v\n", id, inventory)
				statement, err := db.Prepare("UPDATE products SET inventory = inventory - 1 WHERE p_id = ?")
				if err != nil {
					fmt.Println("updating inventory prepare err-")
					log.Fatal(err)
				}
				_, err = statement.Exec(
					id,
				)

			} else if id_type == "ia" {
				//fmt.Printf("Reducing INTEGRATED inventory, for ia id:%v, current inv is: %v\n", id, ia_inventory)
				statement, err := db.Prepare("UPDATE iaddresses SET ia_inventory = ia_inventory - 1	WHERE ia_id = ?")
				if err != nil {
					fmt.Println("updating inventory prepare err-")
					log.Fatal(err)
				}
				_, err = statement.Exec(
					id,
				)
			}

			//	fmt.Println("Updating inventory... id_type" + id_type + " With id: " + strconv.Itoa(id))
			tx.InventoryResult.Success = true
		} else {
			//	fmt.Println("Not updating inventory...Success false!!!!!!!!!!!!!!!!!")
			tx.InventoryResult.Success = false
		}

		if err != nil {
			fmt.Println("-updating inventory err-")
			log.Fatal(err)
		}
	}
	tx.InventoryResult.Id_type = id_type
	tx.InventoryResult.P = p_id
	tx.InventoryResult.Ia = ia_id

	//	p_and_ia_ids.Success = false

	//insert transaction and whether or not it was successful

	var for_ia_id any
	for_ia_id = nil
	successful := 0
	if tx.InventoryResult.Success { //Inventory was duducted from somewhere...
		successful = 1
		for_ia_id = tx.InventoryResult.Ia
	}
	if successful == 1 {
		statement, err := db.Prepare("INSERT INTO incoming (" +
			"txid," +
			"buyer_address," +
			"amount," +
			"port," +
			"for_product_id," +
			"for_ia_id," +
			"ia_comment," +
			"product_label," +
			"successful," +
			"processed," +
			"block_height," +
			"time_utc" +
			") VALUES " +
			"(?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}

		//	fmt.Printf("In insertNewTX 2....\"successful\"-------------:%v\n", successful)

		result, err := statement.Exec(
			tx.Txid,
			crypt.Encrypt(tx.Buyer_address),
			tx.Amount,
			tx.Port,
			tx.For_product_id,
			for_ia_id,
			tx.Ia_comment,
			tx.Product_label,
			1,
			0,
			tx.Block_height,
			tx.Time_utc,
		)
		if err != nil {
			log.Fatal(err)
		}

		if LOGGING {
			affected_rows, _ := result.RowsAffected()
			fmt.Println("Inserted Incoming...?:", affected_rows)
		}
	} else {
		// It is not allowing to insert with the correct values (mainly "successful") into the incoming table so stick em somewhere else for a bit...
		failstatement, err := db.Prepare("INSERT INTO failed_incoming (" +
			"txid," +
			"buyer_address," +
			"amount," +
			"port," +
			"for_product_id," +
			"for_ia_id," +
			"ia_comment," +
			"product_label," +
			"successful," +
			"processed," +
			"block_height," +
			"time_utc" +
			") VALUES " +
			"(?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}

		//	fmt.Printf("In Failed insertNewTX 2....\"successful\"-------------:%v\n", successful)

		result, err := failstatement.Exec(
			tx.Txid,
			crypt.Encrypt(tx.Buyer_address),
			tx.Amount,
			tx.Port,
			tx.For_product_id,
			for_ia_id,
			tx.Ia_comment,
			tx.Product_label,
			0,
			0,
			tx.Block_height,
			tx.Time_utc,
		)
		if err != nil {
			log.Fatal(err)
		}

		if LOGGING {
			affected_rows, _ := result.RowsAffected()
			fmt.Println("Inserted Incoming...?:", affected_rows)
		}

	}
	//loadincoming()

	// last_insert_id, _ := result.LastInsertId()
	return
}

/*
	func txExists(tx Tx) bool {
		//see if tx exists..
		db, err := sql.Open("sqlite3", "./pong.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		var (
			count  string
			count2 string
		)
		err = db.QueryRow("SELECT COUNT(*) FROM incoming WHERE txid = ?", tx.Txid).Scan(&count)
		err = db.QueryRow("SELECT COUNT(*) FROM failedincoming WHERE txid = ?", tx.Txid).Scan(&count2)
		switch {
		case err != nil:
			return false
		case count == "0" && count2 == "0":
			return false
		default:
			//fmt.Println(count)
			return true
		}

}
*/
func addFailedTransactionsBackIntoIncomingTable() {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var tx_list []Tx
	rows, _ := db.Query("SELECT i_id,txid,buyer_address,amount,port,for_product_id,for_ia_id,ia_comment,product_label,successful,processed,block_height,time_utc FROM failed_incoming WHERE processed = '0'")

	var (
		i_id           int
		txid           string
		buyer_address  string
		amount         int
		port           int
		for_product_id int
		for_ia_id      int
		ia_comment     string
		product_label  string
		successful     int
		processed      int
		block_height   int
		time_utc       string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&i_id, &txid, &buyer_address, &amount, &port, &for_product_id, &for_ia_id, &ia_comment, &product_label, &successful, &processed, &block_height, &time_utc)

		if LOGGING {
			//	fmt.Println("Unprocesses Txs")
			//	fmt.Println(strconv.Itoa(i_id) + ": " + txid + " - " + buyer_address + " - " + strconv.Itoa(amount) + " port: " + strconv.Itoa(port) + "Successful??????????" + strconv.Itoa(successful))
		}
		var tx Tx
		tx.I_id = i_id
		tx.Txid = txid
		tx.Buyer_address = crypt.Decrypt(buyer_address)
		tx.Amount = amount
		tx.Port = port
		tx.For_product_id = for_product_id
		tx.For_ia_id = for_ia_id
		tx.Ia_comment = ia_comment
		tx.Product_label = product_label
		tx.Successful = successful != 0
		tx.Processed = processed != 0
		tx.Block_height = block_height
		tx.Time_utc = time_utc

		tx_list = append(tx_list, tx)

	}

	for _, tx := range tx_list {
		statement, err := db.Prepare("INSERT INTO incoming (" +
			"txid," +
			"buyer_address," +
			"amount," +
			"port," +
			"for_product_id," +
			"for_ia_id," +
			"ia_comment," +
			"product_label," +
			"successful," +
			"processed," +
			"block_height," +
			"time_utc" +
			") VALUES " +
			"(?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}

		result, err := statement.Exec(
			tx.Txid,
			crypt.Encrypt(tx.Buyer_address),
			tx.Amount,
			tx.Port,
			tx.For_product_id,
			for_ia_id,
			tx.Ia_comment,
			tx.Product_label,
			0,
			0,
			tx.Block_height,
			tx.Time_utc,
		)
		if err != nil {
			log.Fatal(err)
		}

		if LOGGING {
			affected_rows, _ := result.RowsAffected()
			fmt.Println("Inserted Incoming...?:", affected_rows)
		}
	}

	for _, tx := range tx_list {

		_, err = db.Exec(
			"DELETE FROM failed_incoming WHERE i_id = ?",
			tx.I_id)
		if err != nil {
			fmt.Printf(" Error deleting failed incoming.." + err.Error())
		}

	}

}

// Create orders....
func unprocessedTxs() []Tx {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var tx_list []Tx
	rows, err := db.Query("SELECT i_id,txid,buyer_address,amount,port,for_product_id,for_ia_id,ia_comment,product_label,successful,processed,block_height,time_utc FROM incoming WHERE processed = '0'")
	if err != nil {
		return make([]Tx, 0)
	}
	var (
		i_id           int
		txid           string
		buyer_address  string
		amount         int
		port           int
		for_product_id int
		for_ia_id      int
		ia_comment     string
		product_label  string
		successful     int
		processed      int
		block_height   int
		time_utc       string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&i_id, &txid, &buyer_address, &amount, &port, &for_product_id, &for_ia_id, &ia_comment, &product_label, &successful, &processed, &block_height, &time_utc)

		//	fmt.Printf("yay:%v\n\n", &id)
		/*	if len(image) > 100 {
			imgstr = image[0:100]
		}	*/
		if LOGGING {
			fmt.Println("Unprocessed (incoming) Txs")
			fmt.Println(strconv.Itoa(i_id) + ": " + txid + " - " + buyer_address + " - " + strconv.Itoa(amount) + " port: " + strconv.Itoa(port) + "Successful??????????" + strconv.Itoa(successful))
		}
		var tx Tx
		tx.I_id = i_id
		tx.Txid = txid
		tx.Buyer_address = crypt.Decrypt(buyer_address)
		tx.Amount = amount
		tx.Port = port
		tx.For_product_id = for_product_id
		tx.For_ia_id = for_ia_id
		tx.Ia_comment = ia_comment
		tx.Product_label = product_label
		tx.Successful = successful != 0
		tx.Processed = processed != 0
		tx.Block_height = block_height
		tx.Time_utc = time_utc

		tx_list = append(tx_list, tx)

	}
	return tx_list
}

/** Save shipping Address submission (from after buy / crc32 submission) **/

func saveAddress(address_array map[string]string) bool {
	if LOGGING {
		fmt.Printf("address_array: \n %v \n", address_array)
	}
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		r_id         int
		ship_address string
	)
	err = db.QueryRow("SELECT r_id, IFNULL(ship_address, '') FROM responses WHERE crc32 = ? ORDER BY r_id DESC LIMIT 1", address_array["id"]).Scan(&r_id, &ship_address)

	switch {
	case err != nil:
		fmt.Println("Error selecting address:", err.Error())
		return false
	}
	if LOGGING {
		fmt.Println("r_id:", r_id)
		fmt.Println("ship_address:", ship_address)
	}
	//Already Submitted.

	if ship_address == address_array["txid"] {
		return false
	}
	if LOGGING {
		fmt.Println("address_array[txid]:", address_array["txid"])
	}
	statement, err := db.Prepare(
		"UPDATE responses SET " +
			"ship_address = ? " +
			"WHERE crc32 = ? AND r_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(
		address_array["txid"],
		address_array["id"],
		r_id,
	)
	return err == nil
}

func storeSameBlockAddress(address_array map[string]string, block_height int) bool {
	//Store address until order and response is ready
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//see if shipping address record already exists
	var (
		count string
	)
	_ = db.QueryRow("SELECT COUNT(*) FROM shipping_address WHERE shipping_address_txid = ? AND wallet_address = ? AND block_height = ?",
		address_array["txid"],
		address_array["buyer_address"],
		block_height,
	).Scan(&count)
	switch {
	case count != "0":
		return false
	}

	//sa_id
	statement, err := db.Prepare("INSERT INTO shipping_address (shipping_address_txid,wallet_address,block_height) VALUES (?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(
		address_array["txid"],
		address_array["buyer_address"], //Don't need to encrypt since it should only live for a nanosecond or so...
		block_height,
	)
	return err == nil
}

func getSameBlockShipping(buyer_address string, block_height int) (string, int) {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		sa_id                 int
		shipping_address_txid string
	)
	err = db.QueryRow(
		"SELECT sa_id,shipping_address_txid FROM shipping_address "+
			"WHERE wallet_address = ? AND block_height = ?",
		buyer_address,
		block_height,
	).Scan(&sa_id, &shipping_address_txid)

	switch {
	case err != nil:
		fmt.Printf("Couldn't find same block shipping- " + err.Error())
		return "", 0
	default:
		return shipping_address_txid, sa_id
	}

}
func deleteShippingRecord(sa_id int) {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(
		"DELETE FROM shipping_address WHERE sa_id = ?",
		sa_id)
	if err != nil {
		fmt.Printf("getSameBlockShipping(): Error deleting" + err.Error())
	}
}

func insertOrder(order []Tx, order_type string) bool {
	if LOGGING {
		fmt.Printf("INSERTING ORDER:%v\n\n", order)
	}
	var iids []string
	var iids2 []any
	for _, tx := range order {
		iids = append(iids, strconv.Itoa(tx.I_id))
		iids2 = append(iids2, tx.I_id)
	}
	inc_ids := strings.Join(iids, ",")
	if LOGGING {
		fmt.Println("Product IDS:" + inc_ids)
		fmt.Printf("Product IDS array:%v\n\n", iids2)
	}

	in := strings.Repeat("?,", len(iids)-1) + "?"

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "INSERT INTO orders (" +
		"incoming_ids," +
		"order_type," +
		"order_status" +
		") " +
		"VALUES " +
		"(?,?,?)"

	statement, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	/*	array := append(
		[]string{},
		iids...)
	*/
	_, err = statement.Exec(
		inc_ids,
		order_type,
		"pending",
	)
	if err != nil {
		log.Fatal(err)
	}

	query = "UPDATE incoming SET " +
		"processed='1' " +
		"WHERE i_id IN (" + in + ")"
	statement, err = db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(iids2...)
	if err != nil {
		log.Fatal(err)
	}

	return err == nil
}

// Get pending orders
func getOrdersByStatusAndType(status string, o_type string) []ResponseTx {
	//FIND_IN_SET(, orders.incoming_ids)

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var response_tx_list []ResponseTx

	rows, err := db.Query(
		"SELECT o_id,order_type, "+
			"txid,buyer_address,amount,port,product_label,ia_comment,block_height "+
			"FROM orders "+
			"INNER JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%'))  "+
			"WHERE order_status = ? AND order_type = ? GROUP BY orders.incoming_ids",
		status,
		o_type,
	)

	if err != nil {
		return make([]ResponseTx, 0)
	}
	var (
		o_id          int
		order_type    string
		txid          string
		buyer_address string
		amount        int
		port          int
		product_label string
		ia_comment    string
		block_height  int
	)

	for rows.Next() {
		rows.Scan(&o_id, &order_type, &txid, &buyer_address, &amount, &port, &product_label, &ia_comment, &block_height)

		var rtx ResponseTx
		rtx.Order_id = o_id
		rtx.Txid = txid
		rtx.Type = order_type
		rtx.Buyer_address = crypt.Decrypt(buyer_address)
		rtx.Amount = amount
		rtx.Port = port
		rtx.Product_label = product_label
		rtx.Ia_comment = ia_comment
		rtx.Incoming_height = block_height
		response_tx_list = append(response_tx_list, rtx)

	}
	return response_tx_list

}

func checkForPendingResponseByOrderId(order_id int) map[string]string {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	response := make(map[string]string)
	var (
		out_message string
		uuid        string
	)
	err = db.QueryRow("SELECT out_message,uuid FROM responses WHERE order_id = ? AND confirmed = '0'", order_id).Scan(&out_message, &uuid) /* confirmed = '0' or 0??????????*/

	switch {
	case err != nil:
		return response
	}

	response["out_message"] = out_message
	response["uuid"] = uuid
	return response
}

func markOrderAsProcessed(order_id int) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE orders SET order_status=? WHERE o_id=?")
	if err != nil {
		log.Fatal(err)
	}

	result, err := statement.Exec(
		"confirmed",
		order_id,
	)
	if err != nil {
		log.Fatal(err)
	}
	rows_affected, _ := result.RowsAffected()
	return rows_affected != 0
}

func getOrderDetails(order_id int) []map[string]string {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var detail_set []map[string]string

	rows, err := db.Query(
		"SELECT product_label,ia_comment, "+
			"txid,buyer_address,amount,port,product_label,ia_comment "+
			"FROM orders "+
			"INNER JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%')) "+
			"WHERE orders.o_id = ?",
		order_id,
	)

	if err != nil {
		return make([]map[string]string, 0)
	}
	var (
		product_label string
		ia_comment    string
	)
	//fmt.Println(rows)

	for rows.Next() {
		rows.Scan(&product_label, &ia_comment)

		order_details := make(map[string]string)
		order_details["product_label"] = product_label
		order_details["ia_comment"] = ia_comment
		detail_set = append(detail_set, order_details)

	}
	return detail_set

}

func updateResponseTX(rtx ResponseTx) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE responses SET " +
		"txid=?, " +
		"time_utc=?, " +
		"t_block_height=? " +
		"WHERE order_id=?")
	if err != nil {
		log.Fatal(err)
	}

	result, err := statement.Exec(
		rtx.Txid,
		rtx.Time_utc,
		rtx.T_block_height,
		rtx.Order_id,
	)
	if err != nil {
		log.Fatal(err)
	}
	rows_affected, _ := result.RowsAffected()
	return rows_affected != 0

}

func saveResponse(rtx ResponseTx) bool {
	//See if a response record exists, if so just update the txid (this way only the original uuid is used)
	responseRecord := checkForPendingResponseByOrderId(rtx.Order_id)
	if len(responseRecord) != 0 {
		updateResponseTX(rtx)
		return true
	}

	//No record, insert one.
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("INSERT INTO responses (" +
		"order_id," +
		"txid," +
		"type," +
		"buyer_address," +
		"out_amount," +
		"port," +
		"out_message," +
		"out_message_uuid," +
		"uuid," +
		"api_url," +
		"out_scid," +
		"crc32," +
		"ship_address," +
		"time_utc," +
		"t_block_height," +
		"confirmed" +
		") VALUES " +
		"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	result, err := statement.Exec(
		rtx.Order_id,
		rtx.Txid,
		rtx.Type,
		crypt.Encrypt(rtx.Buyer_address),
		rtx.Respond_amount,
		rtx.Port,
		rtx.Out_message,
		rtx.Out_message_uuid,
		rtx.Uuid,
		rtx.Api_url,
		rtx.Out_scid,
		rtx.Crc32,
		rtx.Ship_address,
		rtx.Time_utc,
		rtx.T_block_height,
		0,
	)
	if err != nil {
		log.Fatal(err)
	}
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Inserted Response...?:", affected_rows)
	}
	// last_insert_id, _ := result.LastInsertId()
	return affected_rows != 0

	//txids,

}

// check responses to ensure they went through, if not mark as not processed
func unConfirmedResponses() []map[string]any {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var unconfirmed_responses []map[string]any
	rows, err := db.Query(
		"SELECT txid,time_utc,t_block_height FROM responses " +
			"JOIN orders ON orders.o_id = responses.order_id  " +
			"WHERE responses.confirmed = '0' AND orders.order_status = 'confirmed'") // WHERE processed = '0'
	if err != nil {
		if LOGGING {
			fmt.Printf("Error: %v\n", err)
		}
		return make([]map[string]any, 0)

	}

	var (
		txid           string
		time_utc       string
		t_block_height int
	)
	if LOGGING {
		fmt.Println("unConfirmedResponses")
		//fmt.Println(rows)
	}
	for rows.Next() {
		rows.Scan(&txid, &time_utc, &t_block_height)
		if LOGGING {
			fmt.Println(
				"\ntxid: " + txid +
					"\ntime_utc: " + time_utc +
					"\nt_block_height: " + strconv.Itoa(t_block_height))
		}
		response := make(map[string]any)
		response["txid"] = txid
		response["time_utc"] = time_utc
		response["t_block_height"] = t_block_height
		unconfirmed_responses = append(unconfirmed_responses, response)

	}
	/*
		if len(unconfirmed_responses) == 0 {
			log.Fatal(err)
		}
	*/
	return unconfirmed_responses
}

func markResAsConfirmed(txid string) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE responses SET " +
		"confirmed=? " +
		"WHERE txid=?")
	if err != nil {
		log.Fatal(err)
	}

	result, err := statement.Exec(
		1,
		txid,
	)
	if err != nil {
		log.Fatal(err)
	}
	rows_affected, _ := result.RowsAffected()
	return rows_affected != 0

}

func getTXCollectionIds(txid string) []int {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var order_ids []int
	rows, err := db.Query(
		"SELECT order_id FROM responses WHERE txid = ?", txid) // WHERE processed = '0'
	if err != nil {
		if LOGGING {
			fmt.Printf("Error: %v\n", err)
		}
		return make([]int, 0)

	}

	var (
		order_id int
	)

	for rows.Next() {
		rows.Scan(&order_id)
		order_ids = append(order_ids, order_id)
	}

	return order_ids

}

// couldn't find confirmation, set order to be retried...
func markOrderAsPending(txid string) bool {
	ids := getTXCollectionIds(txid)
	var sids []string
	for _, id := range ids {
		sids = append(sids, strconv.Itoa(id))
	}

	ids_string := strings.Join(sids, ",")

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE orders SET order_status = 'pending' WHERE o_id IN(" + ids_string + ")")
	if err != nil {
		log.Fatal(err)
	}

	result, err := statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	rows_affected, _ := result.RowsAffected()

	return rows_affected != 0

}

func getConfirmedInc(txid string) []map[string]any {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var record_set []map[string]any
	rows, err := db.Query(
		"SELECT  responses.txid AS txid2, type, out_message_uuid, uuid, for_ia_id, api_url, responses.out_message AS response_out_message,for_product_id FROM responses "+
			"INNER JOIN orders ON responses.order_id = orders.o_id  "+
			"JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%')) "+
			//	"JOIN products ON incoming.for_product_id = products.p_id  "+
			"WHERE responses.txid =  ?", txid)
	if err != nil {
		fmt.Println(err.Error())
		return make([]map[string]any, 0)
	}

	record := make(map[string]any)
	var (
		txid2                string
		type2                string
		out_message_uuid     int
		uuid                 string
		for_ia_id            int
		api_url              string
		response_out_message string
		for_product_id       int
	)
	//fmt.Println(rows)
	for rows.Next() {
		rows.Scan(&txid2, &type2, &out_message_uuid, &uuid, &for_ia_id, &api_url, &response_out_message, &for_product_id)

		record["txid"] = txid2
		record["type"] = type2
		record["out_message_uuid"] = out_message_uuid
		record["uuid"] = uuid
		record["for_ia_id"] = for_ia_id
		record["api_url"] = api_url
		record["response_out_message"] = response_out_message
		//record["for_product_id"] = for_product_id
		record_set = append(record_set, record)

	}
	return record_set

}

/********** TESTING **************/
// Create orders....
func loadincoming() {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT i_id,txid,buyer_address,amount,port,for_product_id,for_ia_id,ia_comment,product_label,successful,processed,block_height,time_utc FROM incoming") // WHERE processed = '0'
	if err != nil {
		log.Fatal(err)
	}
	var (
		i_id           int
		txid           string
		buyer_address  string
		amount         int
		port           int
		for_product_id int
		for_ia_id      int
		ia_comment     string
		product_label  string
		successful     int
		processed      int
		block_height   int
		time_utc       string
	)

	for rows.Next() {
		rows.Scan(&i_id, &txid, &buyer_address, &amount, &port, &for_product_id, &for_ia_id, &ia_comment, &product_label, &successful, &processed, &block_height, &time_utc)

		fmt.Println(strconv.Itoa(i_id) +
			": " + txid +
			" - " + crypt.Decrypt(buyer_address) +
			" - " + strconv.Itoa(amount) +
			" port: " + strconv.Itoa(port) +
			" for pid: " + strconv.Itoa(for_product_id) +
			" for_ia_id: " + strconv.Itoa(for_ia_id) +
			" ia_comment: " + ia_comment +
			"product_label : " + product_label +
			"successful:" + strconv.Itoa(successful) +
			" block: " + strconv.Itoa(block_height))

	}
	return
}

// Get pending orders
func loadorders() {
	//FIND_IN_SET(, orders.incoming_ids)

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT o_id,order_type, "+
			"txid,buyer_address,amount,port,product_label,ia_comment "+
			"FROM orders "+
			"INNER JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%')) "+
			"WHERE order_status = ? AND order_type = ? GROUP BY orders.incoming_ids",
		"pending",
		"token_sale",
	)

	if err != nil {
		//return response_tx_list
	}
	var (
		o_id          int
		order_type    string
		txid          string
		buyer_address string
		amount        int
		port          int
		product_label string
		ia_comment    string
	)

	for rows.Next() {
		rows.Scan(&o_id, &order_type, &txid, &buyer_address, &amount, &port, &product_label, &ia_comment)

		fmt.Println(strconv.Itoa(o_id) + ": ORDERTYPE" + order_type + " TXID " + txid)

	}
	return

}

func loadinventory() {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		p_id         int
		ia_id        int
		inventory    int
		ia_inventory int
	)

	err = db.QueryRow(
		"SELECT ia_id,p_id,inventory,ia_inventory FROM iaddresses "+
			"INNER JOIN products ON (iaddresses.product_id = products.p_id) "+
			"WHERE iaddresses.port = ? AND iaddresses.ask_amount = ? AND iaddresses.status = '1'",
		12345,
		1000,
	).Scan(&p_id, &ia_id, &inventory, &ia_inventory)

	switch {
	case err != nil:
		fmt.Println("error getting update inventory")
		log.Fatal(err)

	}

	fmt.Println(strconv.Itoa(p_id) + strconv.Itoa(ia_id) + " ia_id: " + strconv.Itoa(ia_id) + " inventory: " + strconv.Itoa(inventory) + " ia_inventory" + strconv.Itoa(ia_inventory))

}

// Get pending orders
func loadResponses() {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT r_id, order_id, txid, buyer_address, out_message, ship_address,order_status,confirmed,time_utc,t_block_height FROM responses JOIN orders ON orders.o_id = responses.order_id WHERE orders.order_status = 'confirmed'") //WHERE responses.confirmed = '1' AND orders.order_status = 'confirmed'

	if err != nil {
		//return response_tx_list
	}
	var (
		r_id           int
		order_id       int
		txid           string
		buyer_address  string
		out_message    string
		ship_address   string
		order_status   string
		confirmed      int
		time_utc       string
		t_block_height int
	)

	for rows.Next() {
		rows.Scan(&r_id, &order_id, &txid, &buyer_address, &out_message, &ship_address, &order_status, &confirmed, &time_utc, &t_block_height)

		fmt.Println(strconv.Itoa(r_id) +
			": order_id: " + strconv.Itoa(order_id) +
			" txid: " + txid +
			" buyer_address: " + crypt.Decrypt(buyer_address) +
			" out_message: " + out_message +
			" ship_address: " + ship_address +
			" order_status: " + order_status +
			" confirmed: " + strconv.Itoa(confirmed) +

			" time_utc: " + order_status +
			" t_block_height: " + strconv.Itoa(t_block_height))
	}
	return

}

// Get pending orders
func loadSavedShipping() {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT sa_id,shipping_address_txid, wallet_address, block_height " +
			"FROM shipping_address ")

	if err != nil {
		//return response_tx_list
	}
	var (
		sa_id                 int
		shipping_address_txid string
		wallet_address        string
		block_height          string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&sa_id, &shipping_address_txid, &wallet_address, &block_height)

		//	fmt.Printf("yay:%v\n\n", &id)

		fmt.Println(strconv.Itoa(sa_id) + ": shipping_address_txid: " + shipping_address_txid + " wallet_address: " + wallet_address + " block_height: " + block_height)

	}
	return

}

// Get pending orders
func loadOut() {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT product_label, ia_comment, amount, responses.out_message AS res_out_message, out_amount, responses.buyer_address AS res_buyer_address, ship_address, responses.txid AS res_txid, responses.time_utc AS res_time_utc " +
			"FROM incoming " +
			"RIGHT JOIN orders ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%')) " +
			"INNER JOIN responses ON (orders.o_id = responses.order_id) " +
			"WHERE responses.type = 'sale' OR responses.type = 'token_sale' OR responses.type = 'sc_sale' ")

	if err != nil {
		log.Fatal(err)
	}
	var (
		product_label     string
		ia_comment        string
		amount            int
		res_out_message   string
		out_amount        int
		res_buyer_address string
		ship_address      string
		res_txid          string
		res_time_utc      string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&product_label, &ia_comment, &amount, &res_out_message, &out_amount, &res_buyer_address, &ship_address, &res_txid, &res_time_utc)

		/*	"SELECT txid,time_utc,t_block_height FROM responses " +
			"JOIN orders ON orders.o_id = responses.order_id  " +
			"WHERE responses.confirmed = '0' AND orders.order_status = 'confirmed'")	*/

		fmt.Println(
			"\n product_label: " + product_label +
				"\n ia_comment: " + ia_comment +
				"\n amount: " + strconv.Itoa(amount) +
				"\n res_out_message: " + res_out_message +
				"\n out_amount: " + strconv.Itoa(out_amount) +
				"\n buyer_address: " + crypt.Decrypt(res_buyer_address) +
				"\n ship_address: " + ship_address +
				"\n txid: " + res_txid +
				"\n time_utc: " + res_time_utc)

	}
	return

}
