package settings

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

/*
	type Form struct {
		Form          *widget.Form
		FormContainer *fyne.Container
		FormElements  map[string]*widget.Entry
	}
*/
type Form struct {
	Form          *widget.Form
	FormContainer *fyne.Container
	FormElements  struct {
		Install_time_utc    *widget.Entry
		Daemon_api          *widget.Entry
		Wallet_api          *widget.Entry
		Wallet_api_user     *widget.Entry
		Wallet_api_pass     *widget.Entry
		Web_api             *widget.Entry
		Web_api_user        *widget.Entry
		Web_api_wallet      *widget.Entry
		Web_api_id          *widget.Entry
		Next_checkin_utc    *widget.Entry
		Start_block         *widget.Entry
		Last_synced_block   *widget.Entry
		Start_balance       *widget.Entry
		Last_synced_balance *widget.Entry
	}
}

type WalletConn struct {
	Api  string
	User string
	Pass string
}

type WebAPIConn struct {
	Api    string
	User   string
	Wallet string
	Api_id string
}

type Settings struct {
	Install_time_utc    string
	Daemon_api          string
	Wallet_api          string
	Wallet_api_user     string
	Wallet_api_pass     string
	Web_api             string
	Web_api_user        string
	Web_api_wallet      string
	Web_api_id          string
	Next_checkin_utc    string
	Start_block         int
	Last_synced_block   int
	Start_balance       int
	Last_synced_balance int
}

type NewTxSettings struct {
	Send_uuid  bool
	Send_ia_id bool
}
type AdvancedForm struct {
	Form          *widget.Form
	FormContainer *fyne.Container
	FormElements  struct {
		Send_uuid  *widget.Check
		Send_ia_id *widget.Check
	}
}

/*
	func getSettings() {
		//var connection WalletConnection
		//connection.Ip =

}

	func getWalletConnection() {
		//var connection WalletConnection
		//connection.Ip =
		//	//BadExpr
		return
	}
*/
func Update(sform Form) {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	v := reflect.ValueOf(sform.FormElements)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		//	name :=""
		if strings.ToLower(typeOfS.Field(i).Name) == "last_synced_block" || strings.ToLower(typeOfS.Field(i).Name) == "last_synced_balance" {
			continue
		}
		//fmt.Printf("Field: %s\tValue: %v\n", strings.ToLower(typeOfS.Field(i).Name), v.Field(i).Interface().(*widget.Entry).Text)
		//	if typeOfS.Field(i).Type.String() == "int"{
		//		name = strconv.Itoa( typeOfS.Field(i).Name )
		//	}else{
		name := strings.ToLower(typeOfS.Field(i).Name)
		//	}
		value := v.Field(i).Interface().(*widget.Entry).Text

		statement, err := db.Prepare("UPDATE settings SET value = ? WHERE name = ?;")
		if err != nil {
			log.Fatal(err)
		}
		result, _ := statement.Exec(
			value,
			name,
		)
		affected_rows, _ := result.RowsAffected()
		fmt.Println("Updated?:", affected_rows)
	}

}

func Load() Settings {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var settings Settings
	rows, _ := db.Query("SELECT name, value FROM settings")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		//	fmt.Printf("yay:%v\n\n", &id)
		/*	if len(image) > 100 {
			imgstr = image[0:100]
		}	*/

		//fmt.Println(name + ": " + value)

		switch name {
		case "install_time_utc":
			settings.Install_time_utc = value
		case "daemon_api":
			settings.Daemon_api = value
		case "wallet_api":
			settings.Wallet_api = value
		case "wallet_api_user":
			settings.Wallet_api_user = value
		case "wallet_api_pass":
			settings.Wallet_api_pass = value
		case "web_api":
			settings.Web_api = value
		case "web_api_user":
			settings.Web_api_user = value
		case "web_api_wallet":
			settings.Web_api_wallet = value
		case "web_api_id":
			settings.Web_api_id = value
		case "next_checkin_utc":
			settings.Next_checkin_utc = value
		case "start_block":
			settings.Start_block, _ = strconv.Atoi(value)
		case "last_synced_block":
			settings.Last_synced_block, _ = strconv.Atoi(value)
		case "start_balance":
			settings.Start_balance, _ = strconv.Atoi(value)
		case "last_synced_balance":
			settings.Last_synced_balance, _ = strconv.Atoi(value)

		}

	}
	return settings

}

func GetWalletConn() WalletConn {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var walletConn WalletConn
	rows, _ := db.Query("SELECT name, value FROM settings WHERE type = 'wallet'")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		//fmt.Println(name + ": " + value)

		switch name {
		case "wallet_api":
			walletConn.Api = value
		case "wallet_api_user":
			walletConn.User = value
		case "wallet_api_pass":
			walletConn.Pass = value
		}
	}
	return walletConn
}

func GetDaemonConn() string {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var Api string
	rows, _ := db.Query("SELECT name, value FROM settings WHERE name = 'daemon_api'")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		//fmt.Println(name + ": " + value)

		switch name {
		case "daemon_api":
			Api = value

		}
	}
	return Api
}

func GetWebConn() WebAPIConn {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var webConn WebAPIConn
	rows, _ := db.Query("SELECT name, value FROM settings WHERE type = 'web'")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		//	fmt.Println(name + ": " + value)

		switch name {
		case "web_api":
			webConn.Api = value
		case "web_api_user":
			webConn.User = value
		case "web_api_wallet":
			webConn.Wallet = value
		case "web_api_id":
			webConn.Api_id = value
		}
	}
	return webConn
}

func GetNewTxSettings() NewTxSettings {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var txSettings NewTxSettings
	rows, _ := db.Query("SELECT name, value FROM settings WHERE type = 'web'")
	var (
		name  string
		value string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&name, &value)

		fmt.Println(name + ": " + value)

		switch name {
		case "new_tx_send_uuid":
			txSettings.Send_uuid, _ = strconv.ParseBool(value)
		case "new_tx_send_ia_id":
			txSettings.Send_ia_id, _ = strconv.ParseBool(value)
		}
	}
	return txSettings
}

func UpdateSettingByName(name string, value string) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	statement, err := db.Prepare("UPDATE settings SET value = ? WHERE name = ?;")
	if err != nil {
		log.Fatal(err)
	}
	result, _ := statement.Exec(
		value,
		name,
	)
	affected_rows, _ := result.RowsAffected()
	fmt.Println("Updated?:", affected_rows)
	if affected_rows != 0 {
		return true
	}
	return false
}
