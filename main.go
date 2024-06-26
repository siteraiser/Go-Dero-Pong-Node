package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"

	//"math/rand"

	"node/crypt"
	"node/models/process"
	"node/models/walletapi"
	"node/ui"
)

const LOGGING = false

// var db *sql.DB
// var pform products.Form
// var wref *fyne.Window
var window fyne.Window

type LoginForm struct {
	Password *widget.Entry
}

var logged_in = false
var loginForm LoginForm

const CHECKSUM = "ok"

func login(message string) {

	if !accountCreated() {
		message = "Create password"
	}
	window.SetTitle("Login")

	loginForm.Password = widget.NewPasswordEntry()
	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: message, Widget: loginForm.Password},
		},
		OnSubmit: func() {

			crypt.MakeMD5Hash(loginForm.Password.Text)

			error_msg := dbInit()
			check := checkCheckSum()
			//fmt.Println("CHECKSUM\n" + check)
			if CHECKSUM != check {
				login("Wrong password")
			} else {
				logged_in = true
				ui.Begin(error_msg)
			}

		},
	}
	window.SetContent(container.New(layout.NewStackLayout(), loginForm))
}
func main() {
	//var window fyne.Window

	/*	*/

	ui.Init()

	//daemonapi.NameToAddress("WebGuy")
	window = ui.GetWindowReference()

	login("Enter Password")

	message := dialog.NewInformation("Msg Placeholder", strconv.Itoa(0), window)
	err_dialog := dialog.NewConfirm(
		"placeholder",
		"placeholder",
		func(b bool) {
			if b {
				process.ResetMessages()
			}
		}, window)

	//initialize the token array...
	process.ResetTokenBalances()
	go func() {

		for {
			time.Sleep(time.Second * 7)
			//See if scids have been edited, reset running balances if so.
			if ui.NeedTokenReset() {
				process.ResetTokenBalances()
			}
			//See if paused...
			if !ui.IsPaused() && logged_in {
				//process.newTxs()

				messages, errors_msgs := process.Transactions()
				//Output result of processing to dialogs
				if len(messages) != 0 {
					messages_str := strings.Join(messages[:], ",")
					message.Hide()
					message = dialog.NewInformation("New Info: click 'Yes' to clear list: \n", messages_str, window)
					message = dialog.NewConfirm(
						"New Info: \n",
						messages_str+"\nClick Yes to clear",
						func(b bool) {
							if b {
								process.ResetMessages()
							}
						}, window)
					message.Refresh()
					message.Show()
				}
				if len(errors_msgs) != 0 {
					err_str := strings.Join(errors_msgs[:], "\n")
					err_dialog.Hide()
					err_dialog = dialog.NewConfirm(
						"New Error (click Yes to clear list): \n",
						err_str+"\nClick Yes to clear",
						func(b bool) {
							if b {
								process.ResetErrorMessages()
							}
						}, window)
					err_dialog.Refresh()
					err_dialog.Show()

				}

			}

		}
	}()

	window.ShowAndRun()

}

/***            ***/
/***  Database  ***/
/***            ***/

/* db setup */

// tables
var (
	create_table_incoming = "CREATE TABLE IF NOT EXISTS incoming (" +
		"i_id INTEGER PRIMARY KEY, " +
		"txid TEXT NOT NULL, " +
		"buyer_address TEXT NOT NULL, " +
		"amount UNSIGNED INTEGER, " +
		"port UNSIGNED INTEGER, " +
		"for_product_id UNSIGNED INTEGER, " +
		"for_ia_id UNSIGNED INTEGER, " +
		"ia_comment TEXT, " +
		"product_label TEXT, " +
		"successful UNSIGNED INTEGER, " +
		"processed UNSIGNED INTEGER, " +
		"block_height TEXT NOT NULL, " +
		"time_utc TEXT)"
	create_table_failed_incoming = "CREATE TABLE IF NOT EXISTS failed_incoming (" +
		"i_id INTEGER PRIMARY KEY, " +
		"txid TEXT NOT NULL, " +
		"buyer_address TEXT NOT NULL, " +
		"amount UNSIGNED INTEGER, " +
		"port UNSIGNED INTEGER, " +
		"for_product_id UNSIGNED INTEGER, " +
		"for_ia_id UNSIGNED INTEGER, " +
		"ia_comment TEXT, " +
		"product_label TEXT, " +
		"successful UNSIGNED INTEGER, " +
		"processed UNSIGNED INTEGER, " +
		"block_height TEXT NOT NULL, " +
		"time_utc TEXT)"
	create_table_shipping_address = "CREATE TABLE IF NOT EXISTS shipping_address (" +
		"sa_id INTEGER PRIMARY KEY, " +
		"shipping_address_txid TEXT, " +
		"wallet_address TEXT, " +
		"block_height TEXT)"
	create_table_orders = "CREATE TABLE IF NOT EXISTS orders (" +
		"o_id INTEGER PRIMARY KEY, " +
		"incoming_ids TEXT, " +
		"order_type TEXT, " +
		"order_status TEXT)"
	create_table_responses = "CREATE TABLE IF NOT EXISTS responses (" +
		"r_id INTEGER PRIMARY KEY, " +
		"order_id UNSIGNED INTEGER, " +
		"txid TEXT NOT NULL, " +
		"type TEXT, " +
		"buyer_address TEXT, " +
		"out_amount  UNSIGNED INTEGER, " +
		"port UNSIGNED INTEGER, " +
		"out_message TEXT, " +
		"out_message_uuid UNSIGNED INTEGER, " +
		"uuid TEXT, " +
		"api_url TEXT, " +
		"out_scid TEXT NULL, " +
		"crc32 TEXT, " +
		"ship_address TEXT, " +
		"confirmed  UNSIGNED INTEGER, " +
		"time_utc TEXT, " +
		"t_block_height  UNSIGNED INTEGER)"
	create_table_pending = "CREATE TABLE IF NOT EXISTS pending (" +
		"pend_id INTEGER PRIMARY KEY, " +
		"url TEXT NOT NULL, " +
		"json_text TEXT, " +
		"method TEXT, " +
		"aid TEXT, " +
		"error TEXT)"
	create_table_products = "CREATE TABLE IF NOT EXISTS products (" +
		"p_id INTEGER PRIMARY KEY, " +
		"p_type TEXT, " +
		"tags TEXT NULL, " +
		"label TEXT, " +
		"details TEXT, " +
		"shipping_policy TEXT NULL, " +
		"out_message TEXT, " +
		"out_message_uuid UNSIGNED INTEGER, " +
		"api_url TEXT, " +
		"scid TEXT NULL, " +
		"respond_amount UNSIGNED INTEGER, " +
		"inventory UNSIGNED INTEGER, " + //UNSIGNED NOT NULL,out_message respond_amount
		"image TEXT, " +
		"image_hash TEXT) "
	create_table_iaddresses = "CREATE TABLE IF NOT EXISTS iaddresses (" +
		"ia_id INTEGER PRIMARY KEY, " +
		"product_id INTEGER, " +
		"iaddress TEXT, " +
		"comment TEXT, " +
		"ask_amount UNSIGNED INTEGER, " +
		"ia_respond_amount UNSIGNED INTEGER, " +
		"port UNSIGNED INTEGER, " +
		"ia_scid TEXT NULL, " +
		"ia_inventory UNSIGNED INTEGER, " +
		"status UNSIGNED INTEGER, " +
		"expiry TEXT NOT NULL DEFAULT '')"
	create_table_settings = "CREATE TABLE IF NOT EXISTS settings (" +
		"s_id INTEGER PRIMARY KEY, " +
		"name  TEXT, " +
		"value  TEXT, " +
		"type TEXT)"
)

var table_queries []string = []string{
	//Stores incoming transactions
	create_table_incoming,
	//Stores failed incoming transactions since go with sqlite is seemingly broken
	create_table_failed_incoming,
	//Stores shipping address txids until response has been generated (for same block submissions type)
	create_table_shipping_address,
	//Stores combined orders (physical), digital are seperate since they may have different responses
	create_table_orders,
	//Responses sent back to buyer
	create_table_responses,
	//web api failed transactions
	create_table_pending,
	// products
	create_table_products,
	// integrated addresses
	create_table_iaddresses,
	// settings
	create_table_settings,
}

func dbInit() string {
	//var err string
	var q string = ""

	//Open db
	db, err := sql.Open("sqlite3", "./pong.db")
	handleError(err)

	defer db.Close()

	// Execute table creation queries
	for _, q := range table_queries {
		executeQuery(db, q)
	}

	//Finally try to add new columns
	q = "ALTER TABLE iaddresses ADD COLUMN expiry TEXT NOT NULL DEFAULT ''"
	statement, err := db.Prepare(q)
	if err == nil {
		statement.Exec()
	}

	/*
		q = "ALTER TABLE products ADD COLUMN shipping_policy TEXT NULL"
		statement, err = db.Prepare(q)
		if err == nil {
			statement.Exec()
		}
	*/
	/* Add Starting settings */

	var count int

	err = db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
	switch {
	case err != nil && LOGGING:
		log.Fatal(err)
	default:
		if count == 0 {
			now := time.Now().UTC()
			addtime := 5
			then := now.Add(
				time.Duration(addtime) * time.Minute,
			)
			init_settings := []struct {
				s_type string
				key    string
				value  string
			}{
				// system
				{"system", "checksum", crypt.Encrypt(CHECKSUM)},
				{"system", "install_time_utc", now.Format("2006-01-02 15:04:05")},
				// daemon
				{"daemon", "daemon_api", "node.derofoundation.org:11012"},
				// wallet
				{"wallet", "wallet_api", "127.0.0.1:10103"},
				{"wallet", "wallet_api_user", "secret"},
				{"wallet", "wallet_api_pass", "pass"},
				// web
				{"web", "web_api", "https://derolist.com/papi"},
				{"web", "web_api_user", "Dero User Name"},
				{"web", "web_api_wallet", "Wallet Address"},
				{"web", "web_api_id", ""},
				{"web", "new_tx_send_uuid", "0"},
				{"web", "new_tx_send_ia_id", "0"},
				{"web", "next_checkin_utc", then.Format("2006-01-02 15:04:05")},
			}
			for _, setting := range init_settings {
				insertSetting(db, setting.key, setting.value, setting.s_type)
			}
		}
	}

	/* Add syncronization settings */
	err = db.QueryRow("SELECT COUNT(*) FROM settings WHERE name = 'start_balance'").Scan(&count)
	switch {
	case err != nil && LOGGING:
		log.Fatal(err)
	default:
		if count == 0 {
			block_height := walletapi.GetHeight()
			balance := walletapi.GetBalance()
			if balance > 0 && block_height > 0 {
				sync_settings := []struct {
					s_type string
					key    string
					value  string
				}{
					{"system", "start_block", strconv.Itoa(block_height)},
					{"system", "last_synced_block", strconv.Itoa(block_height)},
					{"system", "start_balance", strconv.Itoa(balance)},
					{"system", "last_synced_balance", strconv.Itoa(balance)},
				}
				for _, setting := range sync_settings {
					insertSetting(db, setting.key, setting.value, setting.s_type)
				}

			} else {
				//Wallet error
				return "Error connecting to wallet or not enough funds."
			}
		}
		//fmt.Printf("Number of rows are %v\n", count)
	}

	return ""
}

func insertSetting(db *sql.DB, name string, value string, s_type string) {
	q := "INSERT INTO settings (name,value,type) VALUES('" + name + "','" + value + "','" + s_type + "');"
	statement, err := db.Prepare(q)
	handleError(err)
	statement.Exec()
}

func checkCheckSum() string {
	db, err := sql.Open("sqlite3", "./pong.db")
	handleError(err)
	defer db.Close()
	checksum := ""
	rows, _ := db.Query("SELECT name, value FROM settings WHERE name = 'checksum'")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		//fmt.Println(name + ":......... " + value)

		switch name {
		case "checksum":
			checksum = crypt.Decrypt(value)

		}
	}

	return checksum
}

// executeQuery prepares and executes a SQL query.
func executeQuery(db *sql.DB, query string) {
	statement, err := db.Prepare(query)
	handleError(err)
	statement.Exec()
}

// handleError logs the error if LOGGING is enabled and then exits the program.
func handleError(err error) {
	if err != nil && LOGGING {
		log.Fatal(err)
	}
}

func accountCreated() bool {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		cks string
	)

	db.QueryRow("SELECT value FROM settings WHERE name = 'checksum' ").Scan(&cks)
	return cks != ""
}
