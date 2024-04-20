package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	helpers "node/helpers"
	loadout "node/loadout"
	iaddresses "node/models/iaddresses"
	products "node/models/products"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func getProductTreeContent() *widget.Tree { //*widget.List
	products_list := products.LoadAll()

	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return getBranchIds(products_list, id)

			/*	switch id {
				case "":
					return []widget.TreeNodeID{"a", "b", "c"}
				case "a":
					return []widget.TreeNodeID{"a1", "a2"}
				}
				return []string{}
			*/
		},
		func(id widget.TreeNodeID) bool {
			pids := getProductIdList(products_list)
			if id == "" {
				return true //[]widget.TreeNodeID{"a", "b", "c"}
			} else if slices.Contains(pids, id) {
				return true
			}

			return false
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return container.New(
					layout.NewHBoxLayout(),
					widget.NewButton("Edit", nil),
					widget.NewLabel("Will be replaced"),
					layout.NewSpacer(),
					widget.NewLabel("Will be replaced"),
				)
			}
			return container.New(
				layout.NewHBoxLayout(),
				widget.NewButton("Edit", nil),
				widget.NewLabel("Leaf template"),
				layout.NewSpacer(),
				widget.NewLabel("Will be replaced"),
			)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			text := id
			tpid := text
			pid, _ := strconv.Atoi(tpid)
			if branch || text[0:1] != "I" {

				//text += " (branch)"
				o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {
					doUpdateLayout(products_list.Items[getIndexFromId(products_list, pid)], true)
				}

				o.(*fyne.Container).Objects[1].(*widget.Label).SetText(products_list.Items[getIndexFromId(products_list, pid)].Label)
				o.(*fyne.Container).Objects[3].(*widget.Label).SetText(strconv.Itoa(products_list.Items[getIndexFromId(products_list, pid)].Inventory))
			} else {
				IAddr := getIAByWackId(text)

				o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {
					doAddIAUpdateLayout(IAddr)
				}
				o.(*fyne.Container).Objects[1].(*widget.Label).SetText(IAddr.Comment)

				o.(*fyne.Container).Objects[3].(*widget.Label).SetText(strconv.Itoa(IAddr.Ia_inventory))

			}

		},
	)
	tree.OpenAllBranches()
	return tree
}

func getIndexFromId(products_list products.List, id int) int {
	for i := 0; i < len(products_list.Items); i++ {
		if products_list.Items[i].Id == id {
			return i
		}

	}
	return 0
}

func getProductIdList(products_list products.List) []widget.TreeNodeID {
	var pids = []widget.TreeNodeID{}
	for i := 0; i < len(products_list.Items); i++ {
		pids = append(pids, strconv.Itoa(products_list.Items[i].Id))
	}
	return pids
}

func getBranchIds(products_list products.List, id string) []widget.TreeNodeID {
	pids := getProductIdList(products_list)

	if id == "" {
		return pids //[]widget.TreeNodeID{"a", "b", "c"}
	} else if slices.Contains(pids, id) {
		return getTreeLeafIds(id)
	}
	fmt.Println("IDS:", pids)
	return []string{}
}

func getTreeLeafIds(id string) []widget.TreeNodeID {
	//fmt.Println("IDS:", id)
	iids := []widget.TreeNodeID{}
	//Get Integrated Addresses...
	pid, _ := strconv.Atoi(id)
	IAList := iaddresses.LoadByProductId(pid)
	for i := 0; i < len(IAList.Items); i++ {
		iids = append(iids, "IA-"+strconv.Itoa(IAList.Items[i].Id))
	}
	return iids

}

func getIAByWackId(id string) iaddresses.IAddress {
	//fmt.Println("IDS:", id)
	iaid := id[3:len(id)]
	ia_id, _ := strconv.Atoi(iaid)
	return iaddresses.LoadById(ia_id)

}

/* generate integrated address table for viewing while updating products */
func getIAddressTable(product products.Product) *widget.Table {
	iaddress_list := iaddresses.LoadByProductId(product.Id)

	headers := []string{"Edit", "Id", "Comment", "Ask Amt.", "Inv.", "Status", "Port", "Integrated Address"}

	table := widget.NewTable(

		func() (int, int) {
			rows := len(iaddress_list.Items) + 1 // data rows with header
			cols := 8
			return rows, cols
		},

		// callback fn for Create each cell.
		func() fyne.CanvasObject {
			l := widget.NewLabel("placeholder")
			l.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			c := container.New(
				layout.NewStackLayout(),
				l,
				widget.NewButton("Edit", nil),
			)

			return c
		},

		// callback fn for Update each cell.
		// This may trigger on initial rendering process.
		// override result of second param in NewTable()
		func(id widget.TableCellID, c fyne.CanvasObject) {

			label := c.(*fyne.Container).Objects[0].(*widget.Label)
			col, row := id.Col, id.Row

			if row == 0 { // Header Row
				label.Alignment = fyne.TextAlignCenter
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(headers[col])
				c.(*fyne.Container).Objects[1].(*widget.Button).Hide()
				return
			}
			// Data row
			ia := iaddress_list.Items[row-1]
			if col == 0 {
				c.(*fyne.Container).Objects[1].(*widget.Button).OnTapped = func() {
					doAddIAUpdateLayout(ia)
				}
				c.(*fyne.Container).Objects[1].(*widget.Button).Resize(fyne.NewSize(50, 30))
				c.(*fyne.Container).Objects[1].(*widget.Button).Show()
			} else {
				c.(*fyne.Container).Objects[1].(*widget.Button).Hide()
				label.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			}
			var text string
			switch col {
			case 0:
				text = "Edit"
			case 1:
				text = strconv.Itoa(ia.Id)
			case 2:
				text = ia.Comment
			case 3:
				text = helpers.ConvertToDeroUnits(ia.Ask_amount)
			case 4:
				text = strconv.Itoa(ia.Ia_inventory)
			case 5:
				text = "Inactive"
				if ia.Status {
					text = "Active"
				}
			case 6:
				text = strconv.Itoa(ia.Port)
			case 7:
				text = ia.Iaddress
			default:
				text = "-"
			}
			label.SetText(text)

		})

	// NOTE: Set width for each columns...
	//
	// Columns for widget.Table is automatically determined from the template object
	// specified in CreateCell (second arg of function NewTable) by default.
	// Here, the size of each column is determined separately from a sample of the data.
	//sample := iaddress_list.Items[0]
	table.SetColumnWidth(0, widget.NewLabel(strconv.Itoa(12345)).MinSize().Width)
	table.SetColumnWidth(1, widget.NewLabel(strconv.Itoa(12345)).MinSize().Width)
	table.SetColumnWidth(2, widget.NewLabel("Sample Comment").MinSize().Width)
	table.SetColumnWidth(3, widget.NewLabel("ask amount 11").MinSize().Width)
	table.SetColumnWidth(4, widget.NewLabel("100inv").MinSize().Width)
	table.SetColumnWidth(5, widget.NewLabel("Inactive").MinSize().Width)
	table.SetColumnWidth(6, widget.NewLabel("Port Num").MinSize().Width)
	table.SetColumnWidth(7, widget.NewLabel("Integrated Address").MinSize().Width)
	//table.SetRowHeight(3, widget.NewLabel(strconv.Itoa(sample.Inventory)).MinSize().Height)

	return table
}

/*
func getOrdersTreeContent() {

	loadout.LoadOrders()
}
*/

func getOrdersTreeContent() *widget.Tree { //*widget.List
	order_list := loadout.LoadOrders()

	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return getOrderBranchIds(order_list, id)

		},
		func(id widget.TreeNodeID) bool {
			pids := getOrderIdList(order_list)
			if id == "" {
				return true //[]widget.TreeNodeID{"a", "b", "c"}
			} else if slices.Contains(pids, id) {
				return true
			}

			return false
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return container.New(
					layout.NewHBoxLayout(),
					widget.NewButton("Details", nil),
					widget.NewLabel("Will be replaced"),
					layout.NewSpacer(),
					widget.NewLabel("Will be replaced"),
				)
			}
			return container.New(
				layout.NewHBoxLayout(),
				widget.NewLabel("Will be replaced"),
			)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			text := id
			toid := text
			oid, _ := strconv.Atoi(toid)

			if branch || text[0:4] != "ORDa" {

				//text += " (branch)"
				o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {
					doFullOrderLayout(oid)
				}
				index := getIndexFromOrderId(order_list, oid)
				count := order_list.Orders[index].Count
				rtype := order_list.Orders[index].R_type
				desc := rtype
				if rtype == "sale" {
					s := ""
					if count > 1 {
						s = "s"
					}
					desc = rtype + ", " + strconv.Itoa(count) + " Item" + s
				}

				o.(*fyne.Container).Objects[1].(*widget.Label).SetText(desc)
				o.(*fyne.Container).Objects[3].(*widget.Label).SetText(order_list.Orders[index].I_time)
			} else { //leaf
				record := getOrderItemByWackId(text)
				rtext := record.Product_label + " / " + record.Ia_comment
				if len(record.Product_label) > 40 {
					rtext = record.Product_label[0:40] + "...|" + record.Ia_comment
				}

				o.(*fyne.Container).Objects[0].(*widget.Label).SetText(rtext)

			}

		},
	)
	//	tree.OpenAllBranches()
	return tree
}
func getIndexFromOrderId(order_list loadout.OrderList, id int) int {
	for i := 0; i < len(order_list.Orders); i++ {
		if order_list.Orders[i].O_id == id {
			return i
		}
	}
	return 0
}
func getOrderIdList(order_list loadout.OrderList) []widget.TreeNodeID {
	var oids = []widget.TreeNodeID{}
	for i := 0; i < len(order_list.Orders); i++ {
		oids = append(oids, strconv.Itoa(order_list.Orders[i].O_id))
	}
	return oids
}

func getOrderBranchIds(order_list loadout.OrderList, id string) []widget.TreeNodeID {
	oids := getOrderIdList(order_list)

	if id == "" {
		return oids //[]widget.TreeNodeID{"a", "b", "c"}
	} else if slices.Contains(oids, id) {
		return getOrderTreeLeafIds(order_list, id)
	}
	fmt.Println("IDS:", oids)
	return []string{}
}

func getOrderTreeLeafIds(order_list loadout.OrderList, id string) []widget.TreeNodeID {
	//fmt.Println("IDS:", id)
	oids := []widget.TreeNodeID{}
	//Get Integrated Addresses...
	oid, _ := strconv.Atoi(id)
	inc_ids := []string{}
	for i := 0; i < len(order_list.Orders); i++ {
		if order_list.Orders[i].O_id == oid {
			inc_ids = strings.Split(order_list.Orders[i].Incoming_ids, ",")
			break
		}

	}

	for i := 0; i < len(inc_ids); i++ {
		oids = append(oids, "ORDa-"+inc_ids[i])
	}
	return oids

}
func getOrderItemByWackId(id string) loadout.Record {
	//fmt.Println("IDS:", id)
	oid := id[5:len(id)]
	o_id, _ := strconv.Atoi(oid)
	return loadout.LoadRecordById(o_id)

}

/* Detailed list*/

/*

func getDetailedOrderListContent() *widget.List { //

	products_list := loadProducts()

	list := widget.NewList(
		func() int {
			return len(products_list.Items)
		},
		func() fyne.CanvasObject {
			return container.New(
				layout.NewHBoxLayout(),
				widget.NewButton("Edit", nil),
				widget.NewLabel("Will be replaced"),
				layout.NewSpacer(),
				widget.NewLabel("Will be replaced"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[0].(*widget.Button).Resize(fyne.NewSize(50, 30))
			o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {

				fmt.Println("I am button " + strconv.Itoa(products_list.Items[i].Id))
			}
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(products_list.Items[i].Label)
			o.(*fyne.Container).Objects[3].(*widget.Label).SetText(strconv.Itoa(products_list.Items[i].Inventory))

		})
	return list
}











func getProductListContent() *widget.List { //

		products_list := loadProducts()

		list := widget.NewList(
			func() int {
				return len(products_list.Items)
			},
			func() fyne.CanvasObject {
				return container.New(
					layout.NewHBoxLayout(),
					widget.NewButton("Edit", nil),
					widget.NewLabel("Will be replaced"),
					layout.NewSpacer(),
					widget.NewLabel("Will be replaced"),
				)
			},
			func(i widget.ListItemID, o fyne.CanvasObject) {
				o.(*fyne.Container).Objects[0].(*widget.Button).Resize(fyne.NewSize(50, 30))
				o.(*fyne.Container).Objects[0].(*widget.Button).OnTapped = func() {

					fmt.Println("I am button " + strconv.Itoa(products_list.Items[i].Id))
				}
				o.(*fyne.Container).Objects[1].(*widget.Label).SetText(products_list.Items[i].Label)
				o.(*fyne.Container).Objects[3].(*widget.Label).SetText(strconv.Itoa(products_list.Items[i].Inventory))

			})
		return list
	}

	func getTable() *widget.Table {
		products_list := loadProducts()
		headers := []string{"id", "Type", "Label", "Details", "Inventory"}

		table := widget.NewTable(

			func() (int, int) {
				rows := len(products_list.Items) + 1 // data rows with header
				cols := 5
				return rows, cols
			},

			// callback fn for Create each cell.
			func() fyne.CanvasObject {
				l := widget.NewLabel("placeholder")
				l.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
				return l
			},

			// callback fn for Update each cell.
			// This may trigger on initial rendering process.
			// override result of second param in NewTable()
			func(id widget.TableCellID, c fyne.CanvasObject) {
				label := c.(*widget.Label)
				col, row := id.Col, id.Row
				if row == 0 { // Header Row
					label.Alignment = fyne.TextAlignCenter
					label.TextStyle = fyne.TextStyle{Bold: true}
					label.SetText(headers[col])
					return
				}
				// Data row
				pro := products_list.Items[row-1]
				var text string
				switch col {
				case 0:
					text = strconv.Itoa(pro.Id)
				case 1:
					text = pro.P_type
				case 2:
					text = pro.Label
				case 3:
					text = pro.Details
				case 4:
					text = strconv.Itoa(pro.Inventory)

				default:
					text = "-"
				}
				label.SetText(text)

			})

		// NOTE: Set width for each columns...
		//
		// Columns for widget.Table is automatically determined from the template object
		// specified in CreateCell (second arg of function NewTable) by default.
		// Here, the size of each column is determined separately from a sample of the data.
		sample := products_list.Items[0]
		table.SetColumnWidth(0, widget.NewLabel(strconv.Itoa(sample.Id)).MinSize().Width)
		table.SetColumnWidth(1, widget.NewLabel(sample.P_type).MinSize().Width)
		table.SetColumnWidth(2, widget.NewLabel(sample.Label).MinSize().Width)
		table.SetColumnWidth(3, widget.NewLabel(sample.Details).MinSize().Width)
		table.SetColumnWidth(4, widget.NewLabel(strconv.Itoa(sample.Inventory)).MinSize().Width)

		//table.SetRowHeight(3, widget.NewLabel(strconv.Itoa(sample.Inventory)).MinSize().Height)

		return table
	}
*/
