package ui

import (
	"image/color"
	"node/helpers"
	loadout "node/loadout"
	"node/models/process"
	walletapi "node/models/walletapi"
	"reflect"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var loadoutlayout []fyne.CanvasObject

func orderLayout(order loadout.Order) []fyne.CanvasObject {
	//order.O_id
	loadoutlayout = []fyne.CanvasObject{}
	addToText(order.O_id)
	addToText("Type: " + order.R_type)
	addToText("Time: " + order.I_time)
	addToText("Number of items: " + strconv.Itoa(order.Count))
	addToText("******* ITEMS *******")
	item_number := 0
	for _, item := range order.Items {
		item_number++
		addToText("******* ITEM " + strconv.Itoa(item_number) + " *******")
		addToText("Product ID: ")
		addToText(strconv.Itoa(item.P_id))
		addToText("")
		addToText("Product Label: ")
		addToText(item.Product_label)
		addToText("")
		addToText("Integrated Address Comment: ")
		addToText(item.Ia_comment)
		addToText("")
		addToText("Amount: " + helpers.ConvertToDeroUnits(item.Amount))
		addToText("Response Out Amount: " + helpers.ConvertToDeroUnits(item.Out_amount))
		addToText("Resonse out message: ")
		addToEntry(item.Res_out_message)
		addToText("")
		addToText("Buyer Wallet Address: ")
		addToEntry(item.Res_buyer_address)
		if item.Ship_address != "" {
			addToText("Buyer Shipping Address: ")
			ship, _ := walletapi.GetTransferByTXID(item.Ship_address)
			address_array := process.GetAddressArray(ship.Entry)
			shipping_text := ""
			if len(address_array) > 8 {
				shippingAddress := process.GetAddressSubmission(address_array)
				shipping_text = shippingAddress.Name + "\n"
				shipping_text += shippingAddress.Level1 + "\n"
				shipping_text += shippingAddress.Level2 + "\n"
				shipping_text += shippingAddress.City + "\n"
				shipping_text += shippingAddress.State + "\n"
				shipping_text += shippingAddress.Zip + "\n"
				shipping_text += shippingAddress.Country + "\n"
			}
			addToMultiLineEntry(shipping_text)
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

	t := canvas.NewText(getString(value), color.RGBA{100, 200, 100, 0xff})
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
