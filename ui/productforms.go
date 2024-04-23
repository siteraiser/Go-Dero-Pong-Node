package ui

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"

	helpers "node/helpers"
	products "node/models/products"
	webapi "node/models/webapi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var pform products.Form
var filling_up = false // check if needed

func disableElements(selected string) {
	if selected == "physical" {
		pform.FormElements.Out_message.Enable()
		pform.FormElements.Out_message_uuid.Enable()
		pform.FormElements.Out_message_uuid.Checked = true
		pform.FormElements.Api_url.Enable()
		pform.FormElements.Scid.Disable()
	} else if selected == "digital" {
		pform.FormElements.Out_message.Enable()
		pform.FormElements.Out_message_uuid.Enable()
		pform.FormElements.Api_url.Enable()
		pform.FormElements.Scid.Disable()
	} else if selected == "token" {
		pform.FormElements.Out_message.Disable()
		pform.FormElements.Out_message_uuid.Disable()
		pform.FormElements.Out_message_uuid.Checked = false
		pform.FormElements.Api_url.Disable()
		pform.FormElements.Api_url.Text = ""
		pform.FormElements.Scid.Enable()
	}

	if !reflect.ValueOf(pform.Form).IsZero() {
		pform.Form.Refresh()
	}
}

func updateElements(selected string) {
	if selected == "physical" {
		pform.FormElements.Out_message.Enable()

		pform.FormElements.Out_message_uuid.Enable()
		pform.FormElements.Out_message.SetPlaceHolder("Leave blank or UUID will be appended (Use UUID must be selected for after the fact shipping submissions)")
		pform.FormElements.Out_message_uuid.Checked = true
		pform.FormElements.Api_url.Enable()
		pform.FormElements.Scid.Disable()
	} else if selected == "digital" {

		pform.FormElements.Out_message.Enable()
		pform.FormElements.Out_message.SetPlaceHolder("Link to E-Goods (https://news.com/eg1) or (https://news.com/eg?id=UUID)")
		pform.FormElements.Out_message_uuid.Enable()
		pform.FormElements.Api_url.Enable()
		pform.FormElements.Scid.Disable()
	} else if selected == "token" {
		pform.FormElements.Out_message.Disable()
		pform.FormElements.Out_message_uuid.Disable()
		pform.FormElements.Out_message_uuid.Checked = false
		pform.FormElements.Api_url.Disable()
		pform.FormElements.Api_url.Text = ""
		pform.FormElements.Scid.Enable()
	}

	if !reflect.ValueOf(pform.Form).IsZero() {
		pform.Form.Refresh()
	}
}

func checkOutMessage() {
	if !filling_up {
		checked := pform.FormElements.Out_message_uuid.Checked
		input_len := len(pform.FormElements.Out_message.Text)
		if reflect.ValueOf(pform.Form).IsZero() {
			if (checked && input_len > 92) || (!checked && input_len > 128) {

				pform.Form.Disable()
			} else {
				pform.Form.Enable()
			}

			pform.Form.Refresh()
			//	log.Println("Check set to", checked)
		}
	}
}

func ResetPForm() {
	pform.FormElements.Label.SetText("")
	pform.FormElements.Details.SetText("")
	pform.FormElements.Inventory.SetText("")
	pform.FormElements.Out_message.SetText("")
	pform.FormElements.Api_url.SetText("")
	pform.FormElements.Scid.SetText("")
	pform.FormElements.Respond_amount.SetText(".00002")
	pform.Img_string = ""
	pform.Img = canvas.NewImageFromImage(helpers.DefaultImg())
	pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
	//pform.EditContainer = container.New(layout.NewGridLayoutWithRows(1), pform.Form)
	canvas.Refresh(pform.Img)
	//pform.Img.Hide()
	pform.Form.Refresh()

}

func createPForm() {

	pform.OpenButton = widget.NewButton("Choose product image", func() {
		openfileDialog := dialog.NewFileOpen(
			func(r fyne.URIReadCloser, _ error) {
				if r == nil {
					pform.Img_string = ""
					pform.Img = canvas.NewImageFromImage(helpers.DefaultImg())
					pform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), pform.Form)
					pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
					//	pform.Img.Hide()
				} else {
					data, _ := io.ReadAll(r)
					result := fyne.NewStaticResource("name", data)
					fmt.Println(result.StaticName + r.URI().Path())
					//window.SetTitle(filepath)
					imgOb := helpers.ResizeImg(r)
					pform.Img_string = helpers.GetImgBase64(imgOb)
					pform.Img = canvas.NewImageFromImage(imgOb)  //
					pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
					pform.FormContainer = container.New(layout.NewGridLayoutWithRows(2), pform.Img, pform.Form)
					//	pform.Img.Show()
				}
				canvas.Refresh(pform.Img)
				//redo layout on close of explorer...
				DoLayout("products/add")

			}, window)
		openfileDialog.SetFilter(
			storage.NewExtensionFileFilter([]string{".png", ".PNG", ".jpg", ".JPG", ".jpeg", ".JPEG"}))
		openfileDialog.Show()
	})

	pform.FormElements.Label = widget.NewEntry()
	//label.SetText("Lorem ipsum ...")
	pform.FormElements.Details = widget.NewMultiLineEntry()

	pform.FormElements.Out_message = widget.NewEntry()
	pform.FormElements.Out_message.Validator = validation.NewRegexp(`^.{0,128}$`, "Max total lngth 128 chars")
	pform.FormElements.Out_message.OnChanged = func(amount string) {
		checkOutMessage()
	}

	pform.FormElements.Out_message_uuid = widget.NewCheck("Use UUID?", func(value bool) {
		checkOutMessage()
	})

	pform.FormElements.Api_url = widget.NewEntry()
	pform.FormElements.Api_url.SetPlaceHolder("API URL to send UUID to if Use UUID is checked")
	pform.FormElements.Scid = widget.NewEntry()
	//pform.FormElements.Respond_amount = widget.NewEntry()
	//pform.FormElements.Respond_amount.Validator = validation.NewRegexp(`^([1-9][0-9]*)$`, "Must be a number greater than 1")

	//Add the select down here (but before a form related disable) to give it access to all of the elements that have been activated above (maybe could add the func on its own)
	pform.FormElements.Selections = []string{"physical", "digital", "token"}
	pform.FormElements.P_type = widget.NewSelect(pform.FormElements.Selections, func(value string) {
		updateElements(value)

		//	log.Println("Select set to", value)
	})
	pform.FormElements.P_type.SetSelectedIndex(0)

	amt := binding.NewString()
	amt.Set(".00002")
	pform.FormElements.Respond_amount = widget.NewEntryWithData(amt)
	pform.FormElements.Respond_amount.Validator = nil
	pform.FormElements.Respond_amount.OnChanged = func(amount string) {
		changes, is_valid := helpers.ValidateAmount(amount) //, pform.FormElements.P_type.Selected
		if changes != "" {
			amt.Set(changes)
		}
		if !is_valid {
			if !reflect.ValueOf(pform.Form).IsZero() {
				pform.Form.Disable()
			}
		} else if pform.Form.Disabled() {
			pform.Form.Enable()
		}
	}

	pform.FormElements.Inventory = widget.NewEntry()
	pform.FormElements.Inventory.Validator = validation.NewRegexp(`^(0|[1-9][0-9]*)$`, "Must be a number")

	pform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor

			{Text: "Image", Widget: pform.OpenButton},
			{Text: "Product Type", Widget: pform.FormElements.P_type},
			{Text: "Label", Widget: pform.FormElements.Label},
			{Text: "Details", Widget: pform.FormElements.Details},
			{Text: "Out Message", Widget: pform.FormElements.Out_message},
			{Text: "Use UUID?", Widget: pform.FormElements.Out_message_uuid},
			{Text: "API URL", Widget: pform.FormElements.Api_url},
			{Text: "SCID", Widget: pform.FormElements.Scid},
			{Text: "Respond Amt.", Widget: pform.FormElements.Respond_amount},
			{Text: "Inv.", Widget: pform.FormElements.Inventory}},
		OnSubmit: func() { // optional, handle pform submission

			if pform.FormElements.Label.Text != "" && pform.FormElements.Details.Text != "" && pform.FormElements.Inventory.Text != "" {

				product := products.LoadById(products.Add(pform))
				//submit with new image
				api_error := webapi.SubmitProduct(product, true)

				if api_error != "" {
					var apiError ApiError
					apiError.Error = api_error
					apiError.Type = "product"
					apiError.Id = product.Id
					apiError.Product = product
					apiErrors.Errors = append(apiErrors.Errors, apiError)
					showApiErrors()
				}
				//reset the pform
				/* ResetPForm()
				//DoLayout("products/list")
				*/

				doUpdateLayout(product, true)
			}
		},
	}

	// we can also append items
	//pform.Form.Append("Details", pform.FormElements.Details)

}

func fillUpdatePForm(product products.Product) {
	//pform.FormElements.Selections = []string{product.P_type}
	pform.FormElements.Label.SetText(product.Label)
	pform.FormElements.Details.SetText(product.Details)
	pform.FormElements.Inventory.SetText(strconv.Itoa(product.Inventory))

	pform.FormElements.Out_message.SetText(product.Out_message)
	filling_up = true
	//fmt.Printf("UUID SELECTED? %v\n", product.Out_message_uuid)
	pform.FormElements.Out_message_uuid.SetChecked(product.Out_message_uuid) //.SetText(product.Out_message)
	filling_up = false
	pform.FormElements.Api_url.SetText(product.Api_url)
	pform.FormElements.Scid.SetText(product.Scid)
	disableElements(product.P_type)
	pform.FormElements.Respond_amount.SetText(helpers.ConvertToDeroUnits(product.Respond_amount))
	pform.Img_string = product.Image
	pform.Img = canvas.NewImageFromImage(helpers.GetImgFromBase64(product.Image))
	pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
	//pform.EditContainer = container.New(layout.NewGridLayoutWithRows(1), pform.Form)
	canvas.Refresh(pform.Img)
	pform.Form.Refresh()
	/*if product.Image == "" {
		//pform.Img.Hide()

	}*/

}

func createUpdatePForm(product products.Product) {

	pform.OpenButton = widget.NewButton("Choose product image", func() {
		openfileDialog := dialog.NewFileOpen(
			func(r fyne.URIReadCloser, _ error) {
				if r == nil {
					pform.Img_string = ""
					//product.Image = "" just reload the product before creating the form
					pform.Img = canvas.NewImageFromImage(helpers.DefaultImg())
					pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
					pform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), pform.Form)
					//pform.Img.Hide()

				} else {
					data, _ := io.ReadAll(r)
					result := fyne.NewStaticResource("name", data)
					fmt.Println(result.StaticName + r.URI().Path())
					//window.SetTitle(filepath)
					imgOb := helpers.ResizeImg(r)
					pform.Img_string = helpers.GetImgBase64(imgOb)
					//	product.Image = pform.Img_string
					pform.Img = canvas.NewImageFromImage(imgOb)
					pform.Img.FillMode = canvas.ImageFillContain //ImageFillOriginal
					pform.FormContainer = container.New(layout.NewGridLayoutWithRows(2), pform.Img, pform.Form)
					//	pform.Img.Show()

				}
				//img = canvas.NewImageFromURI(r.URI())
				canvas.Refresh(pform.Img)
				doUpdateLayout(product, false)

			}, window)
		openfileDialog.SetFilter(
			storage.NewExtensionFileFilter([]string{".png", ".PNG", ".jpg", ".JPG", ".jpeg", ".JPEG"}))
		openfileDialog.Show()
	})

	pform.FormElements.Selections = []string{product.P_type}
	//fix me lol
	pform.FormElements.P_type = widget.NewSelect(pform.FormElements.Selections, func(value string) {
		log.Println("Select set to", value)
	}) /**/
	pform.FormElements.P_type.SetSelectedIndex(0)
	pform.FormElements.Label = widget.NewEntry()
	//label.SetText("Lorem ipsum ...")
	pform.FormElements.Details = widget.NewMultiLineEntry()
	pform.FormElements.Out_message = widget.NewEntry()
	pform.FormElements.Out_message.Validator = validation.NewRegexp(`^.{0,128}$`, "Max total lngth 128 chars")
	pform.FormElements.Out_message.OnChanged = func(amount string) {
		checkOutMessage()
	}
	pform.FormElements.Out_message_uuid = widget.NewCheck("Use UUID?", func(value bool) {

		checkOutMessage()

	})
	pform.FormElements.Api_url = widget.NewEntry()
	pform.FormElements.Scid = widget.NewEntry()

	amt := binding.NewString()
	amt.Set("")
	pform.FormElements.Respond_amount = widget.NewEntryWithData(amt)
	pform.FormElements.Respond_amount.Validator = nil
	pform.FormElements.Respond_amount.OnChanged = func(amount string) {
		changes, is_valid := helpers.ValidateAmount(amount) //, pform.FormElements.P_type.Selected
		if changes != "" {
			amt.Set(changes)
		}
		if !is_valid {

			if !reflect.ValueOf(pform.Form).IsZero() {
				pform.Form.Disable()
			}
			//pform.FormElements.Respond_amount.SetValidationError(errors.New("Fooey"))
		} else if pform.Form.Disabled() {
			pform.Form.Enable()
		}
	}

	pform.FormElements.Inventory = widget.NewEntry()
	pform.FormElements.Inventory.Validator = validation.NewRegexp(`^(0|[1-9][0-9]*)$`, "Must be a number")

	pform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor

			{Text: "Image", Widget: pform.OpenButton},
			{Text: "Product Type", Widget: pform.FormElements.P_type},
			{Text: "Label", Widget: pform.FormElements.Label},
			{Text: "Details", Widget: pform.FormElements.Details},
			{Text: "Out Message", Widget: pform.FormElements.Out_message},
			{Text: "Use UUID?", Widget: pform.FormElements.Out_message_uuid},
			{Text: "API URL", Widget: pform.FormElements.Api_url},
			{Text: "SCID", Widget: pform.FormElements.Scid},
			{Text: "Respond Amt.", Widget: pform.FormElements.Respond_amount},
			{Text: "Inv.", Widget: pform.FormElements.Inventory}},
		OnSubmit: func() { // optional, handle pform submission

			if pform.FormElements.Label.Text != "" && pform.FormElements.Details.Text != "" && pform.FormElements.Inventory.Text != "" {
				fmt.Println("UPDATE IT 1")

				//Should retun if it is a new image for product submissions
				new_image := products.Update(pform, product.Id)
				//reset token balance records in next processing round... (maybe should check if old scid is the same...)
				need_token_reset = true
				//pform.EditContainer = container.New(layout.NewGridLayoutWithRows(2), pform.Img, pform.Form)

				//resetForm()
				//Reload the product I would think...
				product = products.LoadById(product.Id)
				api_error := webapi.SubmitProduct(product, new_image)

				if api_error != "" {
					var apiError ApiError
					apiError.Error = api_error
					apiError.Type = "product"
					apiError.Id = product.Id
					apiError.Product = product
					apiErrors.Errors = append(apiErrors.Errors, apiError)
					showApiErrors()
				}

				//Set True to reload the product to reset image string
				doUpdateLayout(product, true)
				//	img.Refresh()
			}
			//window.Close()
		},
	}

	// we can also append items
	//pform.Form.Append("Details", pform.FormElements.Details)
	return
}

func createUpdateExtraButtons(product products.Product) (*widget.Button, *widget.Button, *widget.Button) {
	backButton := widget.NewButton("Back", func() {
		DoLayout("products")
	})

	addIAddressButton := widget.NewButton("Add Integrated Address", func() {
		doAddIALayout(product)
	})

	deleteButton := widget.NewButton("Delete Product", func() {
		dialog.ShowConfirm(
			"Are you sure yo want to delete this product?",
			product.Label+" will be lost forever.",
			func(b bool) {
				if b {
					deleteProduct(product.Id)
				}
			}, window)
		//log.Printf("tapped%v\n", product.Id)
	})
	return backButton, deleteButton, addIAddressButton
}
