package iaddress

import (
	"database/sql"
	"fmt"
	"log"

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
	}
}

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
	statement, err := db.Prepare("INSERT INTO iaddresses (product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status) VALUES (?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	// string to int
	port_int, _ := strconv.Atoi(FormSubmission.FormElements.Port.Text)
	ia_inventory_int, _ := strconv.Atoi(FormSubmission.FormElements.Ia_inventory.Text)
	ask_amount := helpers.ConvertToAtomicUnits(FormSubmission.FormElements.Ask_amount.Text)
	ia_respond_amount := helpers.ConvertToAtomicUnits(FormSubmission.FormElements.Ia_respond_amount.Text)
	/*	*/

	integrated_address, err := walletapi.MakeIntegratedAddress(port_int, FormSubmission.FormElements.Comment.Text, ask_amount)
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
	)
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Inserted?:", affected_rows)
	}
	last_insert_id, _ := result.LastInsertId()
	return int(last_insert_id), ""

}

func LoadByProductId(pid int) List {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var iaddress_list List
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status FROM iaddresses WHERE product_id =?", pid)
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
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status)

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
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status FROM iaddresses WHERE ia_id =?", iaid)
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
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status)

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
	rows, _ := db.Query("SELECT ia_id, product_id, iaddress, comment, ask_amount, ia_respond_amount, port, ia_scid, ia_inventory, status FROM iaddresses WHERE (ask_amount = ? AND port = ? AND status = '1' AND NOT(ia_id = ?))", amount, portno, iaid)
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
	)

	for rows.Next() {
		rows.Scan(&ia_id, &product_id, &iaddress, &comment, &ask_amount, &ia_respond_amount, &port, &ia_scid, &ia_inventory, &status)

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
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(
		"DELETE FROM iaddresses WHERE ia_id = ?",
		iaid)
	return err == nil

}
