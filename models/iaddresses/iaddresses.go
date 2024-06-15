package iaddress

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	//"math/rand"
	"strconv"

	helpers "node/helpers"

	"node/models/walletapi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const LOGGING = false

type List struct {
	Items []IAddress
}
type IAddress struct {
	Id                int
	Product_id        int
	Iaddress          string
	Comment           string
	Ask_amount        int
	Ia_respond_amount int
	Port              int
	Ia_scid           string
	Ia_inventory      int
	Status            bool
	Expiry            string
}
type Form struct {
	Form          *widget.Form
	FormContainer *fyne.Container
	FormElements  struct {
		Comment           *widget.Entry
		Ask_amount        *widget.Entry
		Ia_respond_amount *widget.Entry
		Port              *widget.Entry
		Ia_scid           *widget.Entry
		Ia_inventory      *widget.Entry
		Status            *widget.Check
		Expiry            *widget.Entry
	}
}

var has_active_expires = "uninitialized"

/*
* Integrated Addresses

func checkIA(){

}
*
*/
func Add(FormSubmission Form, pid int) (int, string) {
	active_ia_error := IsActiveIAddress(FormSubmission, 0)
	if active_ia_error != "" {
		return 0, active_ia_error
	}
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("INSERT INTO iaddresses (product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status, expiry) VALUES (?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	// string to int
	port_int, _ := strconv.Atoi(FormSubmission.FormElements.Port.Text)
	ia_inventory_int, _ := strconv.Atoi(FormSubmission.FormElements.Ia_inventory.Text)
	ask_amount := helpers.ConvertToAtomicUnits(FormSubmission.FormElements.Ask_amount.Text)
	ia_respond_amount := helpers.ConvertToAtomicUnits(FormSubmission.FormElements.Ia_respond_amount.Text)

	/*	*/
	var exp time.Time
	if FormSubmission.FormElements.Expiry.Text != "" {
		exp, err = time.Parse("2006-01-02 15:04:05", FormSubmission.FormElements.Expiry.Text)
		if err != nil && LOGGING {
			fmt.Println("Expiry:" + err.Error())
		}
	}

	integrated_address, err := walletapi.MakeIntegratedAddress(port_int, FormSubmission.FormElements.Comment.Text, ask_amount, exp.Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err.Error()
	}

	result, _ := statement.Exec(
		pid,
		integrated_address,
		FormSubmission.FormElements.Comment.Text,
		ask_amount,
		ia_respond_amount,
		port_int,
		FormSubmission.FormElements.Ia_scid.Text,
		ia_inventory_int,
		FormSubmission.FormElements.Status.Checked,
		FormSubmission.FormElements.Expiry.Text,
	)
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Inserted?:", affected_rows)
	}
	last_insert_id, _ := result.LastInsertId()
	//Set flag for if to check or not
	resetActiveExpires()
	return int(last_insert_id), ""

}

func LoadByProductId(pid int) List {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var iaddress_list List
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status, expiry FROM iaddresses WHERE product_id =?", pid)
	var (
		ia_id             int
		product_id        int
		iaddress          string
		comment           string
		ask_amount        int
		ia_respond_amount int
		port              int
		ia_scid           string
		ia_inventory      int
		status            bool
		expiry            string
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status, &expiry)

		var Iaddress IAddress
		Iaddress.Id = ia_id
		Iaddress.Product_id = product_id
		Iaddress.Iaddress = iaddress
		Iaddress.Comment = comment
		Iaddress.Ask_amount = ask_amount
		Iaddress.Ia_respond_amount = ia_respond_amount
		Iaddress.Port = port
		Iaddress.Ia_scid = ia_scid
		Iaddress.Ia_inventory = ia_inventory
		Iaddress.Status = status
		Iaddress.Expiry = expiry

		iaddress_list.Items = append(iaddress_list.Items, Iaddress)

	}
	return iaddress_list

}

func UpdateById(FormSubmission Form, ia_id int) string {

	active_ia_error := IsActiveIAddress(FormSubmission, ia_id)
	if active_ia_error != "" {
		return active_ia_error
	}

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE iaddresses SET ia_scid = ?, ia_respond_amount = ?, ia_inventory = ?, status = ? WHERE ia_id = ?")
	if err != nil {
		log.Fatal(err)
	}

	// string to int
	inventory_int, _ := strconv.Atoi(FormSubmission.FormElements.Ia_inventory.Text)
	ia_respond_amount_int := helpers.ConvertToAtomicUnits(FormSubmission.FormElements.Ia_respond_amount.Text)
	result, _ := statement.Exec(
		FormSubmission.FormElements.Ia_scid.Text,
		ia_respond_amount_int,
		inventory_int,
		FormSubmission.FormElements.Status.Checked,
		ia_id,
	)
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Inserted?:", affected_rows)
	}
	if affected_rows != 0 {
		return ""
	}
	return "Integrated address not updated"
}

func LoadById(iaid int) IAddress {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var Iaddress IAddress
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status, expiry FROM iaddresses WHERE ia_id =?", iaid)
	var (
		ia_id             int
		product_id        int
		iaddress          string
		comment           string
		ask_amount        int
		ia_respond_amount int
		port              int
		ia_scid           string
		ia_inventory      int
		status            bool
		expiry            string
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status, &expiry)

		Iaddress.Id = ia_id
		Iaddress.Product_id = product_id
		Iaddress.Iaddress = iaddress
		Iaddress.Comment = comment
		Iaddress.Ask_amount = ask_amount
		Iaddress.Ia_respond_amount = ia_respond_amount
		Iaddress.Port = port
		Iaddress.Ia_scid = ia_scid
		Iaddress.Ia_inventory = ia_inventory
		Iaddress.Status = status
		Iaddress.Expiry = expiry

	}
	return Iaddress

}

func GetOtherActiveIA(amount int, portno int, iaid int) IAddress {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var Iaddress IAddress
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status, expiry FROM iaddresses WHERE (ask_amount = ? AND port = ? AND status = '1' AND NOT(ia_id = ?))", amount, portno, iaid)
	var (
		ia_id             int
		product_id        int
		iaddress          string
		comment           string
		ask_amount        int
		ia_respond_amount int
		port              int
		ia_scid           string
		ia_inventory      int
		status            bool
		expiry            string
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status, &expiry)

		Iaddress.Id = ia_id
		Iaddress.Product_id = product_id
		Iaddress.Iaddress = iaddress
		Iaddress.Comment = comment
		Iaddress.Ask_amount = ask_amount
		Iaddress.Ia_respond_amount = ia_respond_amount
		Iaddress.Port = port
		Iaddress.Ia_scid = ia_scid
		Iaddress.Ia_inventory = ia_inventory
		Iaddress.Status = status
		Iaddress.Expiry = expiry

	}
	return Iaddress

}

func GetProductId(ia_id int) int {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		product_id int
	)
	db.QueryRow("SELECT product_id FROM iaddresses WHERE ia_id =?", ia_id).Scan(&product_id)
	/*defer rows.Close()
	for rows.Next() {
		rows.Scan(&image_hash)
		//	fmt.Printf("yay:%v\n\n", &image_hash)

		return image_hash //returns without closing the rows, so use defer rows.Close()

	}*/
	return product_id
}
func DeleteById(iaid int) bool {
	if iaddressIsProcessing(iaid) {
		return false
	}

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(
		"DELETE FROM iaddresses WHERE ia_id = ?",
		iaid)

	//Set flag for if to check or not
	resetActiveExpires()
	return err == nil

}

func iaddressIsProcessing(iaid int) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		count string
	)
	_ = db.QueryRow("SELECT COUNT(*) FROM iaddresses "+
		"JOIN incoming ON (iaddresses.ia_id = incoming.for_ia_id OR incoming.for_ia_id = ifnull(incoming.for_ia_id,''))  "+
		"JOIN orders ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE (incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id))  "+
		"JOIN responses ON (orders.o_id = responses.order_id) "+
		"WHERE iaddresses.ia_id = ? AND (orders.order_status != 'confirmed' OR responses.confirmed = '0' OR incoming.processed = '0')", iaid).Scan(&count)

	return count != "0"
}

//Time / date expiry stuff

func HasActiveExpires() string {
	if has_active_expires == "uninitialized" {
		//fmt.Println(has_active_expires)
		has_active_expires = checkDBForActiveExpires()
		//check once in case seller was offline during expiration
		return "true"
	}
	//fmt.Println(has_active_expires)
	return has_active_expires
}
func resetActiveExpires() {
	has_active_expires = checkDBForActiveExpires()
}
func checkDBForActiveExpires() string {
	now := time.Now().UTC()
	now = now.Add(
		-time.Duration(60) * time.Minute,
	)
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		count string
	)
	//not a perfect solution but keep flag on during the whole day as not to set too early...
	db.QueryRow("SELECT COUNT(*) FROM iaddresses WHERE expiry >= ?", now.Format("2006-01-02 15:04:05")).Scan(&count)
	if count != "0" {
		return "true"
	}
	return "false"
}

func IsExpired(ia_id int) bool {
	now := time.Now().UTC()
	//fmt.Println(now.Format("2006-01-02 15:04:05"))

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		count string
	)
	db.QueryRow("SELECT COUNT(*) FROM iaddresses WHERE ia_id = ? AND expiry != '' AND expiry <= ?", ia_id, now.Format("2006-01-02 15:04:05")).Scan(&count)

	return count != "0"
}

// Set status to 0 if updated, returns true
func SetExpiredIAById(ia_id int) bool {
	//now := time.Now().UTC()
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE iaddresses SET status = '0' WHERE ia_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	result, _ := statement.Exec(ia_id)
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Updated?:", affected_rows)
	}
	if affected_rows != 0 {
		return true
	}
	return false
}

func GetActiveExpired() []int {
	now := time.Now().UTC()
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var expired_ids []int
	rows, _ := db.Query("SELECT ia_id FROM iaddresses WHERE status = '1' AND expiry != '' AND expiry <= ?", now.Format("2006-01-02 15:04:05"))
	var (
		ia_id int
	)

	for rows.Next() {
		rows.Scan(&ia_id)
		expired_ids = append(expired_ids, ia_id)
	}
	return expired_ids

}
