package iaddress

import (
	_ "github.com/mattn/go-sqlite3"

	//"math/rand"
	"strconv"

	helpers "node/helpers"
)

func IsActiveIAddress(iaform Form, ia_id int) string {

	port_int, _ := strconv.Atoi(iaform.FormElements.Port.Text)
	ask_amount := helpers.ConvertToAtomicUnits(iaform.FormElements.Ask_amount.Text)

	iaddress := GetOtherActiveIA(ask_amount, port_int, ia_id)

	if iaddress.Status {
		return "Active integrated address already exists\n with port: " + strconv.Itoa(iaddress.Port) +
			" and ask amount: " + strconv.Itoa(iaddress.Ask_amount) +
			",\n change port, amount or deactivate " + iaddress.Comment +
			" \n with integrated address:\n" + iaddress.Iaddress

	}
	return ""
}
