package ui

import (
	helpers "node/helpers"
	iaddresses "node/models/iaddresses"
	products "node/models/products"
	webapi "node/models/webapi"
	"strconv"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
)

var iaform iaddresses.Form

func resetIAForm() {
	iaform.FormElements.Comment.SetText("")
	iaform.FormElements.Ask_amount.SetText("")
	iaform.FormElements.Ia_respond_amount.SetText("")
	iaform.FormElements.Port.SetText("")
	iaform.FormElements.Ia_scid.SetText("")
	iaform.FormElements.Status.SetText("")
	iaform.FormElements.Ia_inventory.SetText("")
	iaform.Form.Refresh()
}

func createIAForm(product products.Product) {

	iaform.FormElements.Comment = widget.NewEntry()
	iaform.FormElements.Comment.SetPlaceHolder("Comment that appears in wallet")

	amt := binding.NewString()
	amt.Set("")
	iaform.FormElements.Ask_amount = widget.NewEntryWithData(amt)
	iaform.FormElements.Ask_amount.OnChanged = func(amount string) {
		changes, is_valid := helpers.ValidateAmount(amount) //, product.P_type
		if changes != "" {
			amt.Set(changes)
		}
		if !is_valid {
			iaform.Form.Disable()
		} else if iaform.Form.Disabled() {
			iaform.Form.Enable()
		}
	}
	iaform.FormElements.Ask_amount.SetPlaceHolder("Price shown in wallet")

	amt2 := binding.NewString()
	amt2.Set("")
	iaform.FormElements.Ia_respond_amount = widget.NewEntryWithData(amt2)
	iaform.FormElements.Ia_respond_amount.OnChanged = func(amount string) {
		changes, _ := helpers.ValidateAmount(amount) //, product.P_type
		if changes != "" {
			amt2.Set(changes)
		}
	}
	iaform.FormElements.Ia_respond_amount.SetPlaceHolder("Overrides Product Respond Amt. (if not 0)")

	iaform.FormElements.Port = widget.NewEntry()
	iaform.FormElements.Port.SetPlaceHolder("Any integer, shows in wallet (64bit), required for processing")

	iaform.FormElements.Ia_scid = widget.NewEntry()
	if product.P_type != "token" {
		iaform.FormElements.Ia_scid.Disable()
	}
	iaform.FormElements.Ia_scid.SetPlaceHolder("Overrides Product SCID")

	iaform.FormElements.Ia_inventory = widget.NewEntry()
	iaform.FormElements.Ia_inventory.Validator = validation.NewRegexp(`^(0|[1-9][0-9]*)$`, "Must be a number")
	iaform.FormElements.Port.Validator = validation.NewRegexp(`^(0|[1-9][0-9]*)$`, "Must be a number")
	iaform.FormElements.Status = widget.NewCheck(getStatusText(iaform.FormElements.Status.Checked), func(value bool) {
		iaform.FormElements.Status.SetText(getStatusText(value))
	})
	iaform.FormElements.Ia_inventory.SetPlaceHolder("Overrides Product Inventory")

	iaform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Comment", Widget: iaform.FormElements.Comment},
			{Text: "Ask Amount", Widget: iaform.FormElements.Ask_amount},
			{Text: "Respond Amount", Widget: iaform.FormElements.Ia_respond_amount},
			{Text: "Port", Widget: iaform.FormElements.Port},
			{Text: "SCID", Widget: iaform.FormElements.Ia_scid},
			{Text: "Inventory", Widget: iaform.FormElements.Ia_inventory},
			{Text: "Status", Widget: iaform.FormElements.Status},
		},
		OnSubmit: func() { // optional, handle iaform submission

			if iaform.FormElements.Comment.Text != "" {
				//Try to add new...
				id, error_msg := iaddresses.Add(iaform, product.Id)
				webapi.SubmitIAddress(iaddresses.LoadById(id))
				//product)

				if error_msg != "" {
					errors_msgs = append(errors_msgs, error_msg)
					showErrors()
				} else {
					//reset the iaform

					resetIAForm()
					products.LoadAll()
					doUpdateLayout(product, true)
				}
			}
		},
	}

	// we can also append items
	//iaform.Form.Append("Details", iaform.FormElements.Details)

}

func fillUpdateIAForm(iaddress iaddresses.IAddress) {
	iaform.FormElements.Comment.SetText(iaddress.Comment)
	iaform.FormElements.Comment.Disable()
	iaform.FormElements.Ask_amount.SetText(helpers.ConvertToDeroUnits(iaddress.Ask_amount))
	iaform.FormElements.Ask_amount.Disable()
	iaform.FormElements.Ia_respond_amount.SetText(helpers.ConvertToDeroUnits(iaddress.Ia_respond_amount))
	iaform.FormElements.Port.SetText(strconv.Itoa(iaddress.Port))
	iaform.FormElements.Port.Disable()
	iaform.FormElements.Ia_scid.SetText(iaddress.Ia_scid)
	iaform.FormElements.Ia_inventory.SetText(strconv.Itoa(iaddress.Ia_inventory))
	iaform.FormElements.Status.SetChecked(iaddress.Status) //.SetText(product.Out_message)
	iaform.Form.Refresh()

}

func createUpdateIAForm(iaddress iaddresses.IAddress) {
	p_type := products.LoadById(iaddresses.GetProductId(iaddress.Id)).P_type //disableElements(value)
	iaform.FormElements.Comment = widget.NewEntry()

	iaform.FormElements.Ask_amount = widget.NewEntry()

	amt := binding.NewString()
	amt.Set("")
	iaform.FormElements.Ia_respond_amount = widget.NewEntryWithData(amt)
	iaform.FormElements.Ia_respond_amount.OnChanged = func(amount string) {
		changes, _ := helpers.ValidateAmount(amount)
		if changes != "" {
			amt.Set(changes)
		}

	}
	iaform.FormElements.Ia_respond_amount.Validator = validation.NewRegexp(`^([0-9]+([.][0-9]*)?|[.][0-9]+)$`, "Must be greater than .00001")
	iaform.FormElements.Ia_respond_amount.SetPlaceHolder("Overrides Product Respond Amt. (if not 0)")

	iaform.FormElements.Port = widget.NewEntry()
	iaform.FormElements.Ia_scid = widget.NewEntry()
	if p_type != "token" {
		iaform.FormElements.Ia_scid.Disable()
	}
	iaform.FormElements.Ia_scid.SetPlaceHolder("Overrides Product SCID")

	iaform.FormElements.Ia_inventory = widget.NewEntry()
	iaform.FormElements.Ia_inventory.Validator = validation.NewRegexp(`^(0|[1-9][0-9]*)$`, "Must be a number")
	iaform.FormElements.Ia_inventory.SetPlaceHolder("Overrides Product Inventory")

	iaform.FormElements.Status = widget.NewCheck(getStatusText(iaddress.Status), func(value bool) {
		iaform.FormElements.Status.SetText(getStatusText(value))
	})

	iaform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Comment", Widget: iaform.FormElements.Comment},
			{Text: "Ask Amount", Widget: iaform.FormElements.Ask_amount},
			{Text: "Respond Amount", Widget: iaform.FormElements.Ia_respond_amount},
			{Text: "Port", Widget: iaform.FormElements.Port},
			{Text: "SCID", Widget: iaform.FormElements.Ia_scid},
			{Text: "Inventory", Widget: iaform.FormElements.Ia_inventory},
			{Text: "Status", Widget: iaform.FormElements.Status},
		},
		OnSubmit: func() { // optional, handle iaform submission
			//Update status and inventory
			if iaform.FormElements.Comment.Text != "" {

				error_msg := iaddresses.UpdateById(iaform, iaddress.Id)
				webapi.SubmitIAddress(iaddresses.LoadById(iaddress.Id))

				if error_msg != "" {
					errors_msgs = append(errors_msgs, error_msg)
					showErrors()
				} else {
					ia := iaddresses.LoadById(iaddress.Id)
					doAddIAUpdateLayout(ia)
				}

			}
		},
	}

}

func getStatusText(status bool) string {
	if status {
		return "active"
	}
	return "inactive"
}
