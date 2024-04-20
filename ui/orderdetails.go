package ui

import (
	"image/color"
	loadout "node/loadout"
	walletapi "node/models/walletapi"
	"reflect"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/deroproject/derohe/rpc"
)

var loadoutlayout []fyne.CanvasObject

func orderLayout(order loadout.Order) []fyne.CanvasObject {
	//order.O_id

	addToText(order.O_id)
	addToText("Type: " + order.R_type)
	addToText("Time: " + order.I_time)
	addToText("Number of items: " + strconv.Itoa(order.Count))
	addToText("")
	for _, item := range order.Items {
		addToText("Product Label: ")
		addToText(item.Product_label)
		addToText("")
		addToText("Integrated Address Comment: ")
		addToText(item.Ia_comment)
		addToText("")
		addToText("Amount: " + strconv.Itoa(item.Amount))
		addToText("Response Out Amount: " + strconv.Itoa(item.Out_amount))
		addToText("Resonse out message: ")
		addToText(item.Res_out_message)
		addToText("")
		addToText("Buyer Wallet Address: ")
		addToEntry(item.Res_buyer_address)
		if item.Ship_address != "" {
			addToText("Buyer Shipping Address: ")
			ship, _ := walletapi.GetTransferByTXID(item.Ship_address)
			addToMultiLineEntry(ship.Entry.Payload_RPC.Value(rpc.RPC_COMMENT, rpc.DataString).(string))
		}
		addToText("Response TXID: ")
		addToEntry(item.Res_txid)
		addToText("")
	}

	return loadoutlayout
}
func addToMultiLineEntry(value any) {

	w := widget.NewMultiLineEntry()
	w.SetText(getString(value))
	loadoutlayout = append(loadoutlayout, container.New(layout.NewVBoxLayout(), w))

}
func addToEntry(value any) {

	w := widget.NewEntry()
	w.SetText(getString(value))
	loadoutlayout = append(loadoutlayout, container.New(layout.NewVBoxLayout(), w))

}
func addToText(value any) {

	t := canvas.NewText(getString(value), color.Black)
	loadoutlayout = append(loadoutlayout, container.New(layout.NewVBoxLayout(), t))

}
func getString(value any) string {
	str := ""
	if reflect.TypeOf(value).String() == "int" {
		str = strconv.Itoa(value.(int))
	} else if reflect.TypeOf(value).String() == "string" {
		str = value.(string)

	}
	return str
}
