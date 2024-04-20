package ui

import (
	"errors"
	"image/color"
	"log"
	"net/url"
	"reflect"
	"strconv"

	helpers "node/helpers"
	"node/loadout"
	iaddresses "node/models/iaddresses"
	products "node/models/products"
	settings "node/models/settings"
	"node/models/webapi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var window fyne.Window
var errors_msgs []string

var pauseButton *widget.Button
var paused = false

var need_token_reset = false

var topContainer *fyne.Container

func GetWindowReference() fyne.Window {
	return window
}

func showErrors() {
	if len(errors_msgs) > 0 {
		errors_str := ""
		for _, v := range errors_msgs {
			errors_str = errors_str + " " + v
		}
		//fmt.Println(errors_str)
		dialog.ShowError(errors.New(errors_str), window)
		errors_msgs = errors_msgs[:0]
	}
}

func Init() {
	myApp := app.New()
	window = myApp.NewWindow("Dero Pong Node")

	/*	*/
	pform.Img = canvas.NewImageFromImage(helpers.DefaultImg())
	canvas.Refresh(pform.Img)
	//pform.Img.Hide()
	//Go to settings until they have no errors.

	window.Resize(fyne.NewSize(460, 600))

}

func Begin(error_msg string) {

	var start string
	if error_msg == "" {
		start = "home"
	} else {
		start = "settings"
		errors_msgs = append(errors_msgs, error_msg)
	}
	DoLayout(start)

}

func setPauseButton() {
	//log.Printf("paused %v\n", paused)
	buttonTxt := "Pause"
	if paused {
		buttonTxt = "Paused"
	}
	pauseButton = widget.NewButton(buttonTxt, func() {
		if !paused {
			paused = true
			pauseButton.SetText("Paused")
			pauseButton.Refresh()
		} else {
			paused = false
			pauseButton.SetText("Pause")
			pauseButton.Refresh()
		}
	})
}
func IsPaused() bool {
	return paused
}

func NeedTokenReset() bool {
	if need_token_reset {
		need_token_reset = false
		return true
	} else {
		return false
	}
}

// Main Layouts
func DoLayout(route string) {

	var tabs *container.AppTabs

	switch route {
	case "home":
		window.SetTitle("Start")
		tabs = container.NewAppTabs(
			container.NewTabItem("Home", container.NewScroll(getHomeContent())),
			container.NewTabItem("Products", container.NewScroll(getProductTreeContent())),
			container.NewTabItem("Records", container.NewScroll(getOrdersTreeContent())),
			container.NewTabItem("Settings", container.NewScroll(getSettingsContent())),
		)
		tabs.SelectIndex(0)

	case "products":
		tabs = container.NewAppTabs(
			container.NewTabItem("Home", container.NewScroll(getHomeContent())),
			container.NewTabItem("List", container.NewScroll(getProductTreeContent())),
			container.NewTabItem("Add Products", container.NewScroll(getAddProductContent())),
		)
		tabs.SelectIndex(1)

	case "products/add":

		window.SetTitle("Add Products")
		tabs = container.NewAppTabs(
			container.NewTabItem("Home", container.NewScroll(getHomeContent())),
			container.NewTabItem("List", container.NewScroll(getProductTreeContent())),
			container.NewTabItem("Add Products", container.NewScroll(getAddProductContent())),
		)
		tabs.SelectIndex(2)

	case "records":
		window.SetTitle("Products")
		tabs = container.NewAppTabs(
			container.NewTabItem("Home", container.NewScroll(getHomeContent())),
			container.NewTabItem("Products", container.NewScroll(getProductTreeContent())),
			container.NewTabItem("Records", container.NewScroll(getOrdersTreeContent())),
			container.NewTabItem("Settings", container.NewScroll(getSettingsContent())),
		)
		tabs.SelectIndex(2)

	case "settings":
		//settings := settings.Load()
		//fillSettingsForm(settings)

		window.SetTitle("Settings")
		tabs = container.NewAppTabs(
			container.NewTabItem("Home", container.NewScroll(getHomeContent())),
			container.NewTabItem("Products", container.NewScroll(getProductTreeContent())),
			container.NewTabItem("Records", container.NewScroll(getOrdersTreeContent())),
			container.NewTabItem("Settings", container.NewScroll(getSettingsContent())),
		)
		tabs.SelectIndex(3)
		window.SetTitle("Settings")
	}
	/*	*/
	/*	*/

	//widget.NewLabel("World!")

	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Home":
			window.SetTitle("Home")
			DoLayout("home")
		case "Products":
			window.SetTitle("Products Overview")
			DoLayout("products")
		case "Add Products":
			window.SetTitle("Add Products")
			ResetPForm()
			DoLayout("products/add")
		case "Settings":
			window.SetTitle("Settings")
			DoLayout("settings")
		}
	}

	tabs.SetTabLocation(container.TabLocationTop) //TabLocationLeading

	//screen := container.NewScroll(content)
	setPauseButton()
	pauseButton.Resize(fyne.NewSize(150, 30))
	topContainer = container.New(&diagonal{}, pauseButton)

	window.SetContent(container.New(layout.NewStackLayout(), topContainer, tabs))

	showErrors()

}

// Pause button placement
type diagonal struct {
}

func (d *diagonal) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()

		w += childSize.Width
		h += childSize.Height
	}
	return fyne.NewSize(w, h)
}
func (d *diagonal) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {

	o := objects[0]
	pos := fyne.NewPos(window.Canvas().Size().Width-o.MinSize().Width, 0)
	o.Resize(o.MinSize())
	o.Move(pos)
}

func getHomeContent() fyne.CanvasObject {

	//allow retries if they're needed
	retryPending := widget.NewButton("Retry Lost WebAPI Calls", func() {
		webapi.TryPending()
		getHomeContent()
	})
	retryPending.Resize(fyne.NewSize(150, 30))
	if !webapi.CheckPending() {
		retryPending.Disable()
	} else {
		retryPending.Enable()
	}

	t := canvas.NewText("CLI wallet launch command", color.Black)
	launchText := container.New(layout.NewVBoxLayout(), t)
	w := widget.NewEntry()
	w.SetText("dero-wallet-cli-windows-amd64 --daemon-address=node.derofoundation.org:11012 --rpc-server --rpc-bind=127.0.0.1:10103")
	launchEntry := container.New(layout.NewVBoxLayout(), w)

	text1 := canvas.NewText("Test Website", color.Black)
	websiteText := container.New(layout.NewVBoxLayout(), text1)
	w2 := widget.NewEntry()
	w2.SetText("https://www.siteraiser.com/dero-pong-store")
	websiteEntry := container.New(layout.NewVBoxLayout(), w2)

	t2 := canvas.NewText("Dero Donations:", color.Black)
	donateText := container.New(layout.NewVBoxLayout(), t2)
	w3 := widget.NewEntry()
	w3.SetText("WebGuy")
	donateEntry := container.New(layout.NewVBoxLayout(), w3)

	t3 := canvas.NewText("Setup Instructions:", color.Black)

	t4 := canvas.NewText("Make sure you have opened and started", color.Black)
	t5 := canvas.NewText("the recommended Cli-Wallet using a full node.", color.Black)
	t6 := canvas.NewText("Go to settings and set your Dero username if ", color.Black)
	t7 := canvas.NewText("you have one or set to blank. ", color.Black)
	t8 := canvas.NewText("Click update and then register with the webapi", color.Black)
	t9 := canvas.NewText("to get your key (one-time per wallet).", color.Black)
	t10 := canvas.NewText("Start creating products and adding I. Addresses.", color.Black)
	t11 := canvas.NewText("You can pause the processing (and processing error messages)", color.Black)
	t12 := canvas.NewText("which runs every 7 seconds.", color.Black)
	instructionText := container.New(layout.NewVBoxLayout(), t3, t4, t5, t6, t7, t8, t9, t10, t11, t12)

	settingsContainer := container.New(
		layout.NewVBoxLayout(),
		retryPending,
		launchText,
		launchEntry,
		websiteText,
		websiteEntry,
		donateText,
		donateEntry,
		instructionText,
	)
	settingsContainer.Resize(fyne.NewSize(settingsContainer.MinSize().Height, settingsContainer.MinSize().Width))

	//	registerButtonContainer.Resize(fyne.NewSize(150, 30))
	homeContainer := container.New(layout.NewStackLayout(), settingsContainer)

	//treeContain := container.New(layout.NewBorderLayout(topContainer, topContainer, topContainer, topContainer), tree)

	return homeContainer

}

/***                                  ***/
/*** Products layout / form functions ***/
/***                                  ***/
// Called when product is updated
func doUpdateLayout(product products.Product, load_new bool) {

	window.SetTitle("Update Products")

	if load_new {
		//Reload the product to reset image string
		product = products.LoadById(product.Id)
		createUpdatePForm(product)
		fillUpdatePForm(product)
	}

	backButton, deleteButton, addIAddressButton := createUpdateExtraButtons(product)
	deleteButtonContainer, topContainer := prepareButtons(backButton, deleteButton)

	pform.FormContainer = container.New(layout.NewGridLayoutWithRows(2), pform.Img, pform.Form)

	updateContainer := container.New(
		layout.NewVBoxLayout(),
		topContainer,
		pform.FormContainer,
		deleteButtonContainer,
		addIAddressButton,
	)
	formAndIATableContainer := container.New(layout.NewGridLayoutWithRows(2), updateContainer, getIAddressTable(product))
	window.SetContent(container.New(layout.NewStackLayout(), container.NewScroll(formAndIATableContainer)))
}

// Should already have a form, just laying things out here
func prepareButtons(backButton *widget.Button, deleteButton *widget.Button) (*fyne.Container, *fyne.Container) {
	//product := loadProduct()

	setPauseButton()
	topContainer = container.New(
		layout.NewHBoxLayout(),
		backButton,
		layout.NewSpacer(),
		pauseButton,
	)
	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	deleteButtonContainer := container.New(
		layout.NewVBoxLayout(),
		deleteButton,
	)

	return deleteButtonContainer, topContainer
}

// Create a form
func getAddProductContent() *fyne.Container {
	// If no form or coming from product updates and need to make a new one.
	if reflect.ValueOf(pform.Form).IsZero() || len(pform.FormElements.Selections) == 1 {
		//Called when "Add Products" tab is selected and when image is updated (image is done separately from the form)
		createPForm()
	}
	pform.FormContainer = container.New(layout.NewGridLayoutWithRows(2), pform.Img, pform.Form)

	return pform.FormContainer
}

// Delete Product and return to products list
func deleteProduct(pid int) {
	//confirmation.Hide()

	if products.DeleteById(pid) {
		webapi.DeleteProduct(pid)
		log.Printf("Deleted %v\n", pid)
	} else {
		log.Printf("Error Deleting %v\n", pid)
	}
	DoLayout("products")
}

/***                                             ***/
/*** Integrated Address Layouts / Form functions ***/
/***                                             ***/
// Called when product is updated
func doAddIALayout(product products.Product) {

	window.SetTitle("Add Integrated Addresses")

	createIAForm(product)
	backButton := widget.NewButton("Back", func() {
		doUpdateLayout(product, true)
	})
	setPauseButton()
	topContainer = container.New(
		layout.NewHBoxLayout(),
		backButton,
		layout.NewSpacer(),
		pauseButton,
	)
	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	iaform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), iaform.Form)

	IAddressContainer := container.New(
		layout.NewVBoxLayout(),
		topContainer,
		iaform.FormContainer,
	)

	window.SetContent(container.New(layout.NewStackLayout(), container.NewScroll(IAddressContainer)))
}

func doAddIAUpdateLayout(iaddress iaddresses.IAddress) {

	window.SetTitle("Update Integrated Addresses")

	createUpdateIAForm(iaddress)
	fillUpdateIAForm(iaddress)
	product := products.LoadById(iaddress.Product_id)

	integratedAddress := widget.NewEntry()
	integratedAddress.SetText(iaddress.Iaddress)
	integratedAddress.Disable()

	setPauseButton()

	backButton := widget.NewButton("Back", func() {
		doUpdateLayout(product, true)
	})
	topContainer = container.New(
		layout.NewHBoxLayout(),
		backButton,
		layout.NewSpacer(),
		pauseButton,
	)
	pauseButton.Resize(fyne.NewSize(150, 30))
	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	deleteButton := widget.NewButton("Delete", func() {
		dialog.ShowConfirm(
			"Are you sure yo want to delete this integrated Address?",
			iaddress.Comment+" will be lost forever.",
			func(b bool) {
				if b {
					deleteIAddress(iaddress.Id, product)
				}
			}, window)
	})

	iaform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), iaform.Form)

	IAddressContainer := container.New(
		layout.NewVBoxLayout(),
		topContainer,
		integratedAddress,
		iaform.FormContainer,
		deleteButton,
	)

	window.SetContent(container.New(layout.NewStackLayout(), container.NewScroll(IAddressContainer)))
}
func deleteIAddress(iaid int, product products.Product) {
	//confirmation.Hide()

	if iaddresses.DeleteById(iaid) {
		webapi.DeleteIAddress(iaid)
		log.Printf("Deleted %v\n", iaid)
	} else {
		log.Printf("Error Deleting %v\n", iaid)
	}
	doUpdateLayout(product, true)
}

/**** Settings Window Assembly ****/

func getSettingsContent() *fyne.Container {
	createSettingsForm()
	settings := settings.Load()
	fillSettingsForm(settings)

	registerButton := widget.NewButton("Register", func() {
		reg_error := webapi.Register()
		if reg_error == "" {
			DoLayout("settings")
		} else {
			errors_msgs = append(errors_msgs, reg_error)
			showErrors()
		}
	})
	registerButton.Resize(fyne.NewSize(150, 30))

	moreButton := widget.NewButton("Advanced Settings", func() {
		doAdvancedSettingsLayout()
	})
	moreButton.Resize(fyne.NewSize(150, 30))

	_, err := url.ParseRequestURI(settings.Web_api)
	//log.Printf("Prsed... %v\n", err)
	if settings.Web_api_wallet == "Wallet Address" || settings.Web_api_wallet == "" || err != nil {
		registerButton.Disable()
	}

	settingsContainer := container.NewVBox(
		sform.Form,
		registerButton,
		moreButton,
	)

	//	registerButtonContainer.Resize(fyne.NewSize(150, 30))
	sform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), settingsContainer)

	return sform.FormContainer
}

/***                   ***/
/*** ADVANCED SETTINGS ***/
/***                   ***/

func doAdvancedSettingsLayout() {

	window.SetTitle("Advanced Settings")

	backButton := widget.NewButton("Back", func() {
		DoLayout("settings")
	})
	setPauseButton()
	topContainer = container.New(
		layout.NewHBoxLayout(),
		backButton,
		layout.NewSpacer(),
		pauseButton,
	)
	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	createAdvancedSettingsForm()

	fillAdvncedSettingsForm(settings.GetNewTxSettings())

	asform.FormContainer = container.New(layout.NewGridLayoutWithRows(1), asform.Form)

	advancedSettingsContainer := container.New(
		layout.NewVBoxLayout(),
		topContainer,
		asform.FormContainer,
	)

	window.SetContent(container.New(layout.NewStackLayout(), container.NewScroll(advancedSettingsContainer)))
}

/***                ***/
/*** ORDER LOADOUTS ***/
/***                ***/

func doFullOrderLayout(order_id int) {

	window.SetTitle("Order " + strconv.Itoa(order_id) + " Details")

	backButton := widget.NewButton("Back", func() {
		DoLayout("records")
	})
	setPauseButton()
	topContainer = container.New(
		layout.NewHBoxLayout(),
		backButton,
		layout.NewSpacer(),
		pauseButton,
	)
	backButton.Resize(fyne.NewSize(150, 30))
	topContainer.Resize(fyne.NewSize(150, 30))

	order := loadout.LoadOrderById(order_id)
	details := orderLayout(order)
	detailsContainer := container.New(
		layout.NewVBoxLayout(),
		details...,
	)
	fullContainer := container.New(
		layout.NewVBoxLayout(),
		topContainer,
		detailsContainer,
	)

	window.SetContent(container.New(layout.NewGridLayoutWithRows(1), container.NewScroll(fullContainer)))
}
