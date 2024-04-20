package ui

import (
	"log"
	"node/models/daemonapi"
	settings "node/models/settings"
	walletapi "node/models/walletapi"
	"strconv"

	"fyne.io/fyne/v2/widget"
)

var sform settings.Form

func fillSettingsForm(settings settings.Settings) {
	sform.FormElements.Install_time_utc.SetText(settings.Install_time_utc)
	sform.FormElements.Daemon_api.SetText(settings.Daemon_api)
	sform.FormElements.Wallet_api.SetText(settings.Wallet_api)
	sform.FormElements.Wallet_api_user.SetText(settings.Wallet_api_user)
	sform.FormElements.Wallet_api_pass.SetText(settings.Wallet_api_pass)
	sform.FormElements.Web_api.SetText(settings.Web_api)
	sform.FormElements.Web_api_user.SetText(settings.Web_api_user)
	if settings.Web_api_wallet == "Wallet Address" || settings.Web_api_wallet == "" {
		settings.Web_api_wallet = walletapi.GetAddress()
	}
	sform.FormElements.Web_api_wallet.SetText(settings.Web_api_wallet)
	sform.FormElements.Web_api_id.SetText(settings.Web_api_id)
	sform.FormElements.Next_checkin_utc.SetText(settings.Next_checkin_utc)
	sform.FormElements.Start_block.SetText(strconv.Itoa(settings.Start_block))
	sform.FormElements.Start_block.Disable()
	sform.FormElements.Last_synced_block.SetText(strconv.Itoa(settings.Last_synced_block))
	sform.FormElements.Last_synced_block.Disable()
	sform.FormElements.Start_balance.SetText(strconv.Itoa(settings.Start_balance))
	sform.FormElements.Start_balance.Disable()
	sform.FormElements.Last_synced_balance.SetText(strconv.Itoa(settings.Last_synced_balance))
	sform.FormElements.Last_synced_balance.Disable()

	sform.Form.Refresh()

}
func createSettingsForm() {

	sform.FormElements.Install_time_utc = widget.NewEntry()
	sform.FormElements.Daemon_api = widget.NewEntry()
	sform.FormElements.Wallet_api = widget.NewEntry()
	sform.FormElements.Wallet_api_user = widget.NewEntry()
	sform.FormElements.Wallet_api_pass = widget.NewEntry()
	sform.FormElements.Web_api = widget.NewEntry()
	sform.FormElements.Web_api_user = widget.NewEntry()
	sform.FormElements.Web_api_wallet = widget.NewEntry()
	sform.FormElements.Web_api_id = widget.NewEntry()
	sform.FormElements.Next_checkin_utc = widget.NewEntry()
	sform.FormElements.Start_block = widget.NewEntry()
	sform.FormElements.Last_synced_block = widget.NewEntry()
	sform.FormElements.Start_balance = widget.NewEntry()
	sform.FormElements.Last_synced_balance = widget.NewEntry()

	sform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Install_time_utc", Widget: sform.FormElements.Install_time_utc},
			{Text: "Daemon API", Widget: sform.FormElements.Daemon_api},
			{Text: "Wallet API", Widget: sform.FormElements.Wallet_api},
			{Text: "Wallet API User", Widget: sform.FormElements.Wallet_api_user},
			{Text: "Wallet API Pass", Widget: sform.FormElements.Wallet_api_pass},
			{Text: "Web API URL", Widget: sform.FormElements.Web_api},
			{Text: "Web API User", Widget: sform.FormElements.Web_api_user},
			{Text: "Web API Wallet", Widget: sform.FormElements.Web_api_wallet},
			{Text: "Web API Id", Widget: sform.FormElements.Web_api_id},
			{Text: "Next_checkin_utc", Widget: sform.FormElements.Next_checkin_utc},
			{Text: "Start_block", Widget: sform.FormElements.Start_block},
			{Text: "Last_synced_block", Widget: sform.FormElements.Last_synced_block},
			{Text: "Start_balance", Widget: sform.FormElements.Start_balance},
			{Text: "Last_synced_balance", Widget: sform.FormElements.Last_synced_balance},
		},
		OnSubmit: func() { // optional, handle iaform submission

			if sform.FormElements.Install_time_utc.Text != "" {

				if wallet, err := daemonapi.NameToAddress(sform.FormElements.Web_api_user.Text); err != nil || wallet != sform.FormElements.Web_api_wallet.Text {
					sform.FormElements.Web_api_user.Text = ""
				}
				//need product id
				settings.Update(sform)
				//reset token balance records in next processing round... (maybe should check if old scid is the same...)
				need_token_reset = true
				//	new_settings := settings.LoadSettings()
				DoLayout("settings")
			}
		},
	}

	// we can also append items
	//iaform.Form.Append("Details", iaform.FormElements.Details)

}

var asform settings.AdvancedForm

func fillAdvncedSettingsForm(settings settings.NewTxSettings) {
	asform.FormElements.Send_uuid.SetChecked(settings.Send_uuid)
	asform.FormElements.Send_ia_id.SetChecked(settings.Send_ia_id)
	asform.Form.Refresh()

}

func createAdvancedSettingsForm() {

	asform.FormElements.Send_uuid = widget.NewCheck("Send UUID (generated for response)", func(value bool) {
		log.Println("Check set to", value)
	})
	asform.FormElements.Send_ia_id = widget.NewCheck("Send Integrated Address Id", func(value bool) {
		log.Println("Check set to", value)
	})
	asform.Form = &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Web API New TX", Widget: asform.FormElements.Send_uuid},
			{Text: "Web API New TX", Widget: asform.FormElements.Send_ia_id},
		},
		OnSubmit: func() { // optional, handle iaform submission

			//need product id
			//	settings.UpdateAdvanced(asform)
			new_tx_send_uuid := "0"
			if asform.FormElements.Send_uuid.Checked {
				new_tx_send_uuid = "1"
			}
			new_tx_send_ia_id := "0"
			if asform.FormElements.Send_ia_id.Checked {
				new_tx_send_ia_id = "1"
			}
			settings.UpdateSettingByName("new_tx_send_uuid", new_tx_send_uuid)
			settings.UpdateSettingByName("new_tx_send_ia_id", new_tx_send_ia_id)
			//reset token balance records in next processing round... (maybe should check if old scid is the same...)
			//need_token_reset = true
			//	new_settings := settings.LoadSettings()
			doAdvancedSettingsLayout()

		},
	}

	// we can also append items
	//iaform.Form.Append("Details", iaform.FormElements.Details)

}
