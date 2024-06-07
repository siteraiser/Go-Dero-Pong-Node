package ui

import (
	"image/color"
	"node/helpers"
	"node/loadout"
	"node/models/process"
	"node/models/walletapi"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func orderLayout(order loadout.Order) []fyne.CanvasObject {
	// this is a pattern
	// we are creating buckets
	var layout []fyne.CanvasObject
	var (
		item_number        = 0
		order_total        = 0
		shipping_submitted = 0
	)

	addOrderDetails(
		&layout, // we then take the layout bucket
		order,   // and add order details
	)

	for _, item := range order.Items { // now lets add all of the order items to the view
		item_number++              // let's be sure to increment item number
		order_total += item.Amount // and also increment order_total

		addItemDetails(
			&layout,     // then let's take our layout
			item,        // and include and item details
			item_number, // along with its item increment
		)

		if item.Ship_address != "" { // and if we have some ship_address
			shipping_submitted = addShippingDetails(&layout, item)
		}

		addTXID(&layout, item)
	}

	addOrder_totals(
		&layout,
		order_total,
		shipping_submitted,
	)

	return layout
}

func addShippingDetails(layout *[]fyne.CanvasObject, item loadout.Record) (shipping_submitted int) {
	// let's get some details
	shipping_text, submitted := getShippingDetails(item.Ship_address)
	shipping_submitted = submitted

	addToMultiLineEntry(
		layout,        // and now lets take our layout
		shipping_text, // and include shipping information
	)

	return
}

func addTXID(layout *[]fyne.CanvasObject, item loadout.Record) {

	addToText(
		layout,
		"Response TXID: ",
	)

	addToEntry(
		layout,
		item.Res_txid,
	)
}

func addOrderDetails(layout *[]fyne.CanvasObject, order loadout.Order) {
	details := []any{
		order.O_id,
		"Type: " + order.R_type,
		"Time: " + order.I_time,
		"Number of items: " + strconv.Itoa(order.Count),
	}

	for _, detail := range details {
		addToText(layout, detail)
	}
}

func addItemDetails(layout *[]fyne.CanvasObject, item loadout.Record, item_number int) {
	details := []any{
		"******* ITEMS *******",
		"******* ITEM " + strconv.Itoa(item_number) + " *******",
		"Product ID: ",
		strconv.Itoa(item.P_id),
		"Product Label: ",
		item.Product_label,
		"Integrated Address Comment: ",
		item.Ia_comment,
		"Amount: " + helpers.ConvertToDeroUnits(item.Amount),
		"Response Out Amount: " + helpers.ConvertToDeroUnits(item.Out_amount),
		"Response out message: ",
	}

	for _, detail := range details {
		addToText(layout, detail)
	}

	addToEntry(layout, item.Res_out_message)
	addToText(layout, "Buyer Wallet Address: ")
	addToEntry(layout, item.Res_buyer_address)
}

func getShippingDetails(shipAddress string) (string, int) {
	ship, _ := walletapi.GetTransferByTXID(shipAddress)
	shipping_submitted := int(ship.Entry.Amount)
	address_map := process.GetAddressMapFromEntry(ship.Entry)

	if len(address_map) > 8 {
		shippingAddress := process.GetAddressSubmission(address_map)
		addressParts := []string{
			shippingAddress.Name,
			shippingAddress.Level1,
			shippingAddress.Level2,
			shippingAddress.City,
			shippingAddress.State,
			shippingAddress.Zip,
			shippingAddress.Country,
		}
		return strings.Join(addressParts, "\n"), shipping_submitted
	}

	return "", shipping_submitted
}

func addOrder_totals(layout *[]fyne.CanvasObject, order_total, shipping_submitted int) {
	addToText(layout, "Total: "+helpers.ConvertToDeroUnits(order_total))
	if shipping_submitted != 0 {
		addToText(layout, "Shipping Amount Received: "+helpers.ConvertToDeroUnits(shipping_submitted))
		addToText(layout, "Grand Total: "+helpers.ConvertToDeroUnits(order_total+shipping_submitted))
	}
}

func addToMultiLineEntry(layout *[]fyne.CanvasObject, value any) {
	w := widget.NewMultiLineEntry()
	w.SetText(getString(value))
	*layout = append(*layout, container.NewVBox(w))
}

func addToEntry(layout *[]fyne.CanvasObject, value any) {
	w := widget.NewEntry()
	// we don't want them to be able to edit this
	w.Disable() // but it does make sense they would want to copypasta the values
	w.SetText(getString(value))
	*layout = append(*layout, container.NewVBox(w))
}

func addToText(layout *[]fyne.CanvasObject, value any) {
	t := canvas.NewText(getString(value), color.RGBA{100, 200, 100, 0xff})
	*layout = append(*layout, container.NewVBox(t))
}

func getString(value any) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	default:
		return ""
	}
}
