package loadout

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	//"math/rand"
	"strconv"

	"node/crypt"
)

const LOGGING = false

type RecordList struct {
	Items []Record
}
type Record struct {
	P_id              int
	Product_label     string
	Ia_comment        string
	Amount            int
	Out_amount        int
	Ship_address      string
	Res_out_message   string
	Res_buyer_address string
	Res_txid          string
	Res_time_utc      string
}

type OrderList struct {
	Orders []Order
}

type Order struct {
	O_id         int
	Incoming_ids string
	Count        int
	Status       string
	I_time       string
	R_type       string
	Items        []Record
}

// Get orders full loadout, not being used
func loadOut() RecordList {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT product_label, ia_comment, amount, out_amount, ship_address, responses.out_message AS res_out_message, responses.buyer_address AS res_buyer_address, responses.txid AS res_txid, responses.time_utc AS res_time_utc " +
			"FROM incoming " +
			"RIGHT JOIN orders ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%'))  " +
			"INNER JOIN responses ON (orders.o_id = responses.order_id) " +
			"WHERE responses.type = 'sale' OR responses.type = 'token_sale' OR responses.type = 'sc_sale' ")

	if err != nil {
		log.Fatal(err)
	}
	var record_list RecordList
	var (
		product_label     string
		ia_comment        string
		amount            int
		out_amount        int
		ship_address      string
		res_out_message   string
		res_buyer_address string
		res_txid          string
		res_time_utc      string
	)

	for rows.Next() {
		rows.Scan(&product_label, &ia_comment, &amount, &out_amount, &ship_address, &res_out_message, &res_buyer_address, &res_txid, &res_time_utc)

		fmt.Println(
			"\n product_label: " + product_label +
				"\n ia_comment: " + ia_comment +
				"\n amount: " + strconv.Itoa(amount) +
				"\n out_amount: " + strconv.Itoa(out_amount) +
				"\n ship_address: " + ship_address +
				"\n res_out_message: " + res_out_message +
				"\n buyer_address: " + crypt.Decrypt(res_buyer_address) +
				"\n txid: " + res_txid +
				"\n time_utc: " + res_time_utc)

		var record Record
		record.Product_label = product_label
		record.Ia_comment = ia_comment
		record.Amount = amount
		record.Out_amount = out_amount
		record.Ship_address = ship_address
		record.Res_out_message = res_out_message
		record.Res_buyer_address = crypt.Decrypt(res_buyer_address)
		record.Res_txid = res_txid
		record.Res_time_utc = res_time_utc

		record_list.Items = append(record_list.Items, record)

	}
	return record_list

}

// Get orders
func LoadOrders() OrderList {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(
		"SELECT o_id,incoming_ids,order_status,  incoming.time_utc AS i_time,  responses.type AS r_type FROM orders " +
			"LEFT JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%'))  " +
			"INNER JOIN responses ON (orders.o_id = responses.order_id) " +
			"WHERE responses.type = 'sale' OR responses.type = 'token_sale' GROUP BY o_id ORDER BY o_id DESC") //OR responses.type = 'sc_sale'

	if err != nil {
		log.Fatal(err)
	}
	var order_list OrderList
	var (
		o_id         int
		incomimg_ids string
		order_status string
		i_time       string
		r_type       string
	)
	//order_ids := []int{}
	//imgstr := ""
	fmt.Println("")
	//fmt.Printf("\nrows: %v", rows)
	for rows.Next() {
		rows.Scan(&o_id, &incomimg_ids, &order_status, &i_time, &r_type)
		fmt.Println(strconv.Itoa(o_id) + ": incomimg_ids" + incomimg_ids + " Status " + order_status + " Time " + i_time + " Type: " + r_type)
		var order Order
		order.O_id = o_id
		order.Incoming_ids = incomimg_ids
		count := strings.Split(incomimg_ids, ",")
		order.Count = len(count)
		order.Status = order_status
		order.I_time = i_time
		order.R_type = r_type

		//	if !slices.Contains(order_ids, o_id) {
		//	order_ids = append(order_ids, o_id)
		order_list.Orders = append(order_list.Orders, order)
		//	}

	}
	return order_list

}

// Used in tree list and loadouts...
func LoadRecordById(item_id int) Record {
	fmt.Println("Loading Record By Id:" + strconv.Itoa(item_id))
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var record Record
	var (
		for_product_id    int
		product_label     string
		ia_comment        string
		amount            int
		out_amount        int
		ship_address      string
		res_out_message   string
		res_buyer_address string
		res_txid          string
		res_time_utc      string
	)
	err = db.QueryRow(
		"SELECT for_product_id,product_label, ia_comment, amount, out_amount, ship_address, responses.out_message AS res_out_message, responses.buyer_address AS res_buyer_address, responses.txid AS res_txid, responses.time_utc AS res_time_utc  FROM orders "+
			"JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%')) "+
			"INNER JOIN responses ON (orders.o_id = responses.order_id) "+
			"WHERE (responses.type = 'sale' OR responses.type = 'token_sale' OR responses.type = 'sc_sale') AND incoming.i_id = ?",
		item_id).Scan(&for_product_id, &product_label, &ia_comment, &amount, &out_amount, &ship_address, &res_out_message, &res_buyer_address, &res_txid, &res_time_utc)

	switch {
	case err != nil:
		if LOGGING {
			fmt.Println("error record")
		}
		log.Fatal(err)

	}

	//	fmt.Printf("\n\nShip Address: %v", ship_address)
	record.P_id = for_product_id
	record.Product_label = product_label
	record.Ia_comment = ia_comment
	record.Amount = amount
	record.Out_amount = out_amount
	record.Ship_address = ship_address
	record.Res_out_message = res_out_message
	record.Res_buyer_address = crypt.Decrypt(res_buyer_address)
	record.Res_txid = res_txid
	record.Res_time_utc = res_time_utc
	return record

}

// used for order loadout
func LoadOrderById(order_id int) Order {
	fmt.Println("Loading Order By Id")
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var order Order
	var (
		o_id         int
		incomimg_ids string
		order_status string
		i_time       string
		r_type       string
	)

	err = db.QueryRow(
		"SELECT o_id,incoming_ids,order_status,incoming.time_utc AS i_time,responses.type AS r_type FROM orders "+
			"JOIN incoming ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%'))  "+
			"INNER JOIN responses ON (orders.o_id = responses.order_id) "+
			"WHERE (responses.type = 'sale' OR responses.type = 'token_sale' OR responses.type = 'sc_sale') AND orders.o_id = ?",
		order_id).Scan(&o_id, &incomimg_ids, &order_status, &i_time, &r_type)

	switch {
	case err != nil:
		fmt.Println("error record")
		log.Fatal(err)

		//	return p_and_ia_ids
	}

	order.O_id = o_id
	order.Incoming_ids = incomimg_ids
	ids := strings.Split(incomimg_ids, ",")
	order.Count = len(ids)
	order.Status = order_status
	order.I_time = i_time
	order.R_type = r_type

	for _, id := range ids {
		int_id, _ := strconv.Atoi(id)
		order.Items = append(order.Items, LoadRecordById(int_id))
	}
	return order
}
