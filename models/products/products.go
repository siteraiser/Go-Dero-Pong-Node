package product

import (
	"database/sql"
	"fmt"
	"hash/crc32"
	"log"

	_ "github.com/mattn/go-sqlite3"

	//"math/rand"
	"strconv"

	helpers "node/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const LOGGING = false

type List struct {
	Items []Product
}
type Product struct {
	Id               int
	P_type           string
	Tags             string
	Label            string
	Details          string
	Shipping_policy  sql.NullString
	Out_message      string
	Out_message_uuid bool
	Api_url          string
	Scid             string
	Respond_amount   int
	Inventory        int
	Image            string
}
type Form struct {
	Form          *widget.Form
	OpenButton    *widget.Button
	FormContainer *fyne.Container
	Img           *canvas.Image
	Img_string    string
	FormElements  struct {
		Id               *widget.Entry
		Selections       []string
		P_type           *widget.Select
		Tags             *widget.Entry
		Label            *widget.Entry
		Details          *widget.Entry
		Shipping_policy  *widget.Entry
		Out_message      *widget.Entry
		Out_message_uuid *widget.Check
		Api_url          *widget.Entry
		Scid             *widget.Entry
		Respond_amount   *widget.Entry
		Inventory        *widget.Entry
		Image            string
	}
}

/* Products */
func Add(pform Form) int {
	table := crc32.MakeTable(crc32.IEEE)
	checksum := crc32.Checksum([]byte(pform.Img_string), table)
	image_hash := strconv.FormatInt(int64(checksum), 10)

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("INSERT INTO products (p_type, tags, label, details, shipping_policy, out_message, out_message_uuid, api_url, scid, respond_amount, inventory, image, image_hash) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	// string to int
	inventory_int, _ := strconv.Atoi(pform.FormElements.Inventory.Text)
	//respond_amount_int, _ := strconv.Atoi(FormSubmission.FormElements.Respond_amount.Text)
	/*	*/
	result, _ := statement.Exec(
		pform.FormElements.Selections[pform.FormElements.P_type.SelectedIndex()],
		pform.FormElements.Tags.Text,
		pform.FormElements.Label.Text,
		pform.FormElements.Details.Text,
		pform.FormElements.Shipping_policy.Text,
		pform.FormElements.Out_message.Text,
		pform.FormElements.Out_message_uuid.Checked,
		pform.FormElements.Api_url.Text,
		pform.FormElements.Scid.Text,
		helpers.ConvertToAtomicUnits(pform.FormElements.Respond_amount.Text),
		inventory_int,
		pform.Img_string,
		image_hash,
	)
	affected_rows, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("Inserted?:", affected_rows)
	}
	last_insert_id, _ := result.LastInsertId()
	return int(last_insert_id)
}

func LoadAll() List {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var product_list List
	rows, err := db.Query(
		`
		SELECT p_id, 
		p_type, 
		tags, 
		label, 
		details, 
		shipping_policy, 
		out_message, 
		out_message_uuid, 
		api_url, 
		scid, 
		respond_amount, 
		inventory, 
		image 
		FROM products
		`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var (
		p_id             int
		p_type           string
		tags             string
		label            string
		details          string
		shipping_policy  sql.NullString
		out_message      string
		out_message_uuid bool
		api_url          string
		scid             string
		respond_amount   int
		inventory        int
		image            string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		err = rows.Scan(
			&p_id,
			&p_type,
			&tags,
			&label,
			&details,
			&shipping_policy,
			&out_message,
			&out_message_uuid,
			&api_url,
			&scid,
			&respond_amount,
			&inventory,
			&image,
		)
		if err != nil {
			log.Fatal(err)
		}

		var product Product = Product{
			Id:               p_id,
			P_type:           p_type,
			Tags:             tags,
			Label:            label,
			Details:          details,
			Shipping_policy:  shipping_policy,
			Out_message:      out_message,
			Out_message_uuid: out_message_uuid,
			Api_url:          api_url,
			Scid:             scid,
			Respond_amount:   respond_amount,
			Inventory:        inventory,
			Image:            image,
		}
		product_list.Items = append(product_list.Items, product)
		//fmt.Println("product.Image:", image)
	}
	// Check for errors after iterating through the rows
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return product_list
}

func Update(pform Form, pid int) bool {
	//should reset the running totals for scid tokens...
	new_image := false
	image_hash := ""
	if pform.Img_string != "" {

		table := crc32.MakeTable(crc32.IEEE) //0
		checksum := crc32.Checksum([]byte(pform.Img_string), table)

		image_hash = fmt.Sprint(checksum)
	}
	current_hash := getProductImageHash(pid) //needed to defer rows.Close...
	if current_hash != image_hash {
		new_image = true
	}

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("UPDATE products SET p_type = ?, tags = ?, label = ?, details = ?, shipping_policy = ?, out_message = ?, out_message_uuid = ?, api_url = ?, scid = ?, respond_amount = ?, inventory = ?, image = ?, image_hash = ? WHERE p_id = ?")
	if err != nil {
		log.Fatal(err)
	}

	// string to int
	inventory_int, _ := strconv.Atoi(pform.FormElements.Inventory.Text)

	result, _ := statement.Exec(
		pform.FormElements.Selections[pform.FormElements.P_type.SelectedIndex()],
		pform.FormElements.Tags.Text,
		pform.FormElements.Label.Text,
		pform.FormElements.Details.Text,
		pform.FormElements.Shipping_policy.Text,
		pform.FormElements.Out_message.Text,
		pform.FormElements.Out_message_uuid.Checked,
		pform.FormElements.Api_url.Text,
		pform.FormElements.Scid.Text,
		helpers.ConvertToAtomicUnits(pform.FormElements.Respond_amount.Text),
		inventory_int,
		pform.Img_string,
		image_hash,
		pid,
	)

	rows_affected, _ := result.RowsAffected()
	if LOGGING {
		fmt.Println("update?:", rows_affected)
	}
	return new_image

}

func LoadById(pid int) Product {
	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var product Product
	rows, _ := db.Query("SELECT p_id, p_type, tags, label, details, shipping_policy, out_message, out_message_uuid, api_url, scid, respond_amount, inventory, image FROM products WHERE p_id =?", pid)
	var (
		p_id             int
		p_type           string
		tags             string
		label            string
		details          string
		shipping_policy  sql.NullString
		out_message      string
		out_message_uuid bool
		api_url          string
		scid             string
		respond_amount   int
		inventory        int
		image            string
	)
	//fmt.Println(rows)

	//imgstr := ""
	for rows.Next() {
		rows.Scan(&p_id, &p_type, &tags, &label, &details, &shipping_policy, &out_message, &out_message_uuid, &api_url, &scid, &respond_amount, &inventory, &image)

		//	fmt.Printf("yay:%v\n\n", &id)
		/*	if len(image) > 100 {
				imgstr = image[0:100]
			}
			fmt.Println(strconv.Itoa(id) + ": " + p_type + " - " + label + " - " + details + " Image: " + imgstr)
		*/

		product.Id = p_id
		product.P_type = p_type
		product.Tags = tags
		product.Label = label
		product.Details = details
		product.Shipping_policy = shipping_policy
		product.Out_message = out_message
		product.Out_message_uuid = out_message_uuid
		product.Api_url = api_url
		product.Scid = scid
		product.Respond_amount = respond_amount
		product.Inventory = inventory
		product.Image = image

	}
	return product

}

func DeleteById(pid int) bool {

	if productIsProcessing(pid) {
		return false
	}

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(
		"DELETE FROM iaddresses WHERE product_id = ?",
		pid)
	if err != nil {
		return false
	}

	_, err = db.Exec(
		"DELETE FROM products WHERE p_id = ?",
		pid)

	return err == nil

}

func getProductImageHash(pid int) string {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		image_hash = ""
	)
	db.QueryRow("SELECT image_hash FROM products WHERE p_id =?", pid).Scan(&image_hash)
	/*defer rows.Close()
	for rows.Next() {
		rows.Scan(&image_hash)
		//	fmt.Printf("yay:%v\n\n", &image_hash)

		return image_hash //returns without closing the rows, so use defer rows.Close()

	}*/
	return image_hash
}

/**/
func productIsProcessing(pid int) bool {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		count string
	)
	_ = db.QueryRow("SELECT COUNT(*) FROM products "+
		"JOIN iaddresses ON iaddresses.product_id = products.p_id   "+
		"JOIN incoming ON (iaddresses.ia_id = incoming.for_ia_id OR incoming.for_ia_id = ifnull(incoming.for_ia_id,''))  "+
		"JOIN orders ON (orders.incoming_ids = incoming.i_id) OR (orders.incoming_ids LIKE ('%' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || ',%')) OR (orders.incoming_ids LIKE ('%,' || incoming.i_id || '%'))  "+
		"JOIN responses ON (orders.o_id = responses.order_id) "+
		"WHERE products.p_id = ? AND (orders.order_status != 'confirmed' OR responses.confirmed = '0' OR incoming.processed = '0')", pid).Scan(&count)

	return count != "0"
}
