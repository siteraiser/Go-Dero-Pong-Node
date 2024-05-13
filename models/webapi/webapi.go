package webapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/maphash"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	iaddresses "node/models/iaddresses"
	products "node/models/products"
	settings "node/models/settings"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const LOGGING = false

/*
	type TransactionList struct {
		Transactions []Product
	}

	type GetAddress struct {
		jsonrpc string
		id      string
		method  string
	}
*/
type Registration struct {
	Username string `json:"username"`
	Wallet   string `json:"wallet"`
}
type RegistrationResult struct {
	Success bool   `json:"success"`
	Reg     string `json:"reg"`
}

type ProductSubmission struct {
	Id              int    `json:"id"`
	P_type          string `json:"p_type"`
	Tags            string `json:"tags"`
	Label           string `json:"label"`
	Details         string `json:"details"`
	Shipping_policy string `json:"shipping_policy"`
	Scid            string `json:"scid"`
	Inventory       int    `json:"inventory"`
	Image           string `json:"image,omitempty"`
}
type GeneralSubmissionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type ProductSubmissionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
type DeleteSubmission struct {
	Id     int    `json:"id"`
	Action string `json:"action"`
}

type IASubmission struct {
	Id           int    `json:"id"`
	Product_id   int    `json:"product_id"`
	Iaddr        string `json:"iaddr"`
	Ask_amount   int    `json:"ask_amount"`
	Comment      string `json:"comment"`
	Ia_scid      string `json:"ia_scid"`
	Status       bool   `json:"status"`
	Ia_inventory int    `json:"ia_inventory"`
}
type IASubmissionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type NewTx struct {
	Uuid  string `json:"uuid"`
	Ia_id int    `json:"ia_id"`
}
type NewTxSubmissionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

/*
$data["method"] = "submitProduct";

	$params=[];
	$params["id"] = (int)$product_id;
	$params["action"] = 'delete';

	type Product struct {
		Id               int
		P_type           string
		Label            string
		Details          string
		Out_message      string
		Out_message_uuid bool
		Api_url          string
		Respond_amount   int
		Inventory        int
		Image            string
	}
*/
func NewPOST(data []byte) (req *http.Request, cancel context.CancelFunc, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	req, err = http.NewRequestWithContext(ctx, "POST", settings.GetWebConn().Api,
		strings.NewReader(string(data)),
	)
	if err == nil {
		req.SetBasicAuth(settings.GetWebConn().User, settings.GetWebConn().Api_id)
		req.Header.Add("Content-Type", "application/json")
	}
	return req, cancel, err
}
func NewPUT(data []byte) (req *http.Request, cancel context.CancelFunc, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	req, err = http.NewRequestWithContext(ctx, "PUT", settings.GetWebConn().Api,
		strings.NewReader(string(data)),
	)
	if err == nil {
		req.SetBasicAuth(settings.GetWebConn().User, settings.GetWebConn().Api_id)
		req.Header.Add("Content-Type", "application/json")
	}
	return req, cancel, err
}
func Register() string {

	client := &http.Client{}
	var reg Registration
	reg.Username = settings.GetWebConn().User
	reg.Wallet = settings.GetWebConn().Wallet
	data, err := json.Marshal(map[string]interface{}{
		"method": "register",
		"params": reg,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPOST(data)
	defer cancel()
	if err != nil {
		return "New request W/ context: \n" + err.Error()
	}

	resp, err := client.Do(req)
	if err != nil {
		return "Post: " + err.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ReadAll: " + err.Error()
	}

	var result RegistrationResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	fmt.Println(result)
	//Don't save unless we don't have an id and there is one returned from the webapi
	if settings.GetWebConn().Api_id == "" && result.Reg != "" {
		if settings.UpdateSettingByName("web_api_id", result.Reg) {
			return ""
		} else {
			return "Error saving web_api_id to db"
		}

	}
	return ""

	/**/

	// service address can be created client side for now
}

func SubmitProduct(product products.Product, new_image bool) string {
	if LOGGING {
		fmt.Println("SUBMITTING PRODUCT.......")
	}
	client := &http.Client{}
	var productSubmission ProductSubmission
	productSubmission.Id = product.Id
	productSubmission.P_type = product.P_type
	productSubmission.Tags = product.Tags
	productSubmission.Label = product.Label
	productSubmission.Details = product.Details
	productSubmission.Shipping_policy = product.Shipping_policy
	productSubmission.Scid = product.Scid
	productSubmission.Inventory = product.Inventory
	if new_image {
		productSubmission.Image = product.Image
	}

	data, err := json.Marshal(map[string]interface{}{
		"method": "submitProduct",
		"params": productSubmission,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPOST(data)
	defer cancel()
	if err != nil {
		return "New request W/ context: \n" + err.Error()
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		//json_str, _ := json.Marshal(data)
		//logRequest(settings.GetWebConn().Api, string(data), err3.Error(), "submitProduct", strconv.Itoa(product.Id))
		return "Post: " + err3.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ReadAll: " + err.Error()
	}

	var result ProductSubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	if LOGGING {
		fmt.Println("Submit Product result -")
		fmt.Println(result)
	}
	//Don't save unless we don't have an id and there is one returned from the webapi
	if !result.Success && result.Error == "" {
		//json_str, _ := json.Marshal(data)
		//return "ReadAll: " + err.Error()
		//logRequest(settings.GetWebConn().Api, string(data), "API Error", "submitProduct", strconv.Itoa(product.Id))
	}

	return ""

}

func DeleteProduct(pid int) string {
	fmt.Println("Deleting PRODUCT.......")
	client := &http.Client{}
	var deleteProductSubmission DeleteSubmission
	deleteProductSubmission.Id = pid
	deleteProductSubmission.Action = "delete"

	data, err := json.Marshal(map[string]interface{}{
		"method": "submitProduct",
		"params": deleteProductSubmission,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPOST(data)
	defer cancel()
	if err != nil {
		return "New request W/ context: \n" + err.Error()
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		//json_str, _ := json.Marshal(data)
		//	logRequest(settings.GetWebConn().Api, string(data), err3.Error(), "submitProduct", strconv.Itoa(pid))
		return "Post: " + err3.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ReadAll: " + err.Error()
	}

	var result ProductSubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	if LOGGING {
		fmt.Println("Delete Product result -")
		fmt.Println(result)
	}

	if !result.Success && result.Error == "" {
		//json_str, _ := json.Marshal(data)

		//logRequest(settings.GetWebConn().Api, string(data), "API Error", "submitProduct", strconv.Itoa(pid))
	}

	return ""

}

func SubmitIAddress(iaddress iaddresses.IAddress) string {
	if LOGGING {
		fmt.Println("SUBMITTING PRODUCT.......")
	}
	client := &http.Client{}
	var IASubmission IASubmission
	IASubmission.Id = iaddress.Id
	IASubmission.Product_id = iaddress.Product_id
	IASubmission.Iaddr = iaddress.Iaddress
	IASubmission.Ask_amount = iaddress.Ask_amount
	IASubmission.Comment = iaddress.Comment
	IASubmission.Ia_inventory = iaddress.Ia_inventory
	IASubmission.Status = iaddress.Status
	data, err := json.Marshal(map[string]interface{}{
		"method": "submitIAddress",
		"params": IASubmission,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err1 := NewPOST(data)
	defer cancel()
	if err1 != nil {
		return "New request W/ context: \n" + err1.Error()
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		//json_str, _ := json.Marshal(data)
		//	logRequest(settings.GetWebConn().Api, string(data), err3.Error(), "submitIAddress", strconv.Itoa(iaddress.Id))
		return "Post: " + err3.Error()
	}
	defer resp.Body.Close()
	body, err4 := io.ReadAll(resp.Body)
	if err4 != nil {
		return "ReadAll: " + err4.Error()
	}

	var result IASubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	if LOGGING {
		fmt.Println("Submit I.A. Result-")
		fmt.Println(result)
		fmt.Println(result.Success)
		fmt.Println(result.Error)
	}
	if !result.Success && result.Error == "" {
		//json_str, _ := json.Marshal(data)
		//	logRequest(settings.GetWebConn().Api, string(data), "API Error", "submitIAddress", strconv.Itoa(iaddress.Id))
	}

	return ""

}

func DeleteIAddress(iaid int) string {
	if LOGGING {
		fmt.Println("Deleting PRODUCT.......")
	}
	client := &http.Client{}
	var deleteProductSubmission DeleteSubmission
	deleteProductSubmission.Id = iaid
	deleteProductSubmission.Action = "delete"

	data, err := json.Marshal(map[string]interface{}{
		"method": "submitProduct",
		"params": deleteProductSubmission,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPOST(data)
	defer cancel()
	if err != nil {
		return "New request W/ context: \n" + err.Error()
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		//	json_str, _ := json.Marshal(data)
		//	logRequest(settings.GetWebConn().Api, string(data), err3.Error(), "submitIAddress", strconv.Itoa(iaid))
		return "Post: " + err3.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ReadAll: " + err.Error()
	}

	var result ProductSubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	if LOGGING {
		fmt.Println("Delete IAddress result -")
		fmt.Println(result)
	}
	if !result.Success && result.Error == "" {
		//	json_str, _ := json.Marshal(data)

		//	logRequest(settings.GetWebConn().Api, string(data), "API Error", "submitIAddress", strconv.Itoa(iaid))
	}

	return ""

}

/* not seems to works...
func CheckIn() {
	fmt.Println("Checking In.......")
	client := &http.Client{}

	data, err := json.Marshal(map[string]interface{}{
		"method": "checkIn",
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPUT(data)
	defer cancel()
	if err != nil {
		_, _ = client.Do(req)
	}

}
*/

// works...
func CheckIn() {
	if LOGGING {
		fmt.Println("Checking In.......")
	}
	client := &http.Client{}

	data, err := json.Marshal(map[string]interface{}{
		"method": "checkIn",
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewPUT(data)
	defer cancel()
	if err != nil {
		if LOGGING {
			fmt.Printf("%v\nPut: error: " + err.Error())
		}
		return
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		fmt.Printf("%v\nPut: " + err3.Error())
		return
	}
	//might be key to keeping the request going long enough to complete the put, seems without this fails...
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil && LOGGING {
		fmt.Printf("%v\nPut: ReadAll: " + err.Error())

	}
}

// errorHandling(settings.GetWebConn().Api, http_status_code, result, data, "submitProduct", strconv.Itoa(product.Id))
/*func errorHandling(api_url string, http_status_code int, result ProductSubmissionResult, data []byte, method string, applicable_id string) {
	fmt.Println("HTTP status:", http_status_code)
	if !result.Success && result.Error == "" || http_status_code != 200 {
		res_err := "API Error"
		//json_str, _ := json.Marshal(data)
		if http_status_code != 200 {
			fmt.Println("Non-OK HTTP status:", http_status_code)
			res_err = "HTTP Error: " + strconv.Itoa(http_status_code)
		}
		logRequest(api_url, string(data), res_err, method, applicable_id)
	}
}
*/
/***********/
/* pending */
/***********/

func logRequest(url string, jsontxt string, err string, method string, applicable_id string) bool {
	//make sure seller has an account setup
	if settings.GetWebConn().Api_id == "" {
		return false
	}
	//	if($this->id==''){return false;}
	//Save request and error

	if err != "" {

		db, err := sql.Open("sqlite3", "./pong.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		statement, err := db.Prepare(
			"INSERT INTO pending " +
				"(url,json_text,method,aid,error)" +
				"VALUES" +
				"(?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}

		result, _ := statement.Exec(
			url,
			jsontxt,
			method,
			applicable_id,
			err,
		)
		affected_rows, _ := result.RowsAffected()
		if LOGGING {
			fmt.Println("Erroroneus Request Logged to Pending:")
		}
		//last_insert_id, _ := result.LastInsertId()
		//return int(last_insert_id)

		return affected_rows != 0

	}
	//If successful remove any older failed requests for same product or I.A.
	deleteRequests(method, applicable_id)
	return false
}

func deleteRequests(method string, applicable_id string) {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	statement, err := db.Prepare("DELETE FROM pending WHERE method = ? AND aid=?")
	if err != nil {
		log.Fatal(err)
	}

	_, _ = statement.Exec(
		method,
		applicable_id,
	)
	//affected_rows, _ := result.RowsAffected()

}

// still linked to home retry button
func TryPending() {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	retries := []map[string]string{}
	rows, _ := db.Query("SELECT url,json_text,error,method,aid FROM pending WHERE pend_id IN (SELECT MAX(pend_id) FROM pending GROUP BY url,method,aid)")
	var (
		url       string
		json_text string
		oerr      string
		method    string
		aid       string
	)

	for rows.Next() {
		rows.Scan(&url, &json_text, &oerr, &method, &aid)
		fmt.Println("-------------------")
		fmt.Println(url)
		fmt.Println(oerr)
		fmt.Println(method)
		fmt.Println("-------------------")
		retry := make(map[string]string)
		retry["url"] = url
		retry["json_text"] = json_text
		retry["oerr"] = oerr
		retry["method"] = method
		retry["aid"] = aid
		retries = append(retries, retry)

	}

	for _, r := range retries {
		fmt.Println("-------------------")
		fmt.Println(r["oerr"])
		fmt.Println(r["method"])
		fmt.Println(r["aid"])
		fmt.Println("-------------------")
		deleteRequests(r["method"], r["aid"])
		retry(r["url"], r["json_text"], r["oerr"], r["method"], r["aid"])
	}

}

func retry(url string, json_text string, oerr string, method string, aid string) {
	if LOGGING {
		fmt.Println("Retry.......")

		//	fmt.Println(json.Marshal(json_text))
	}

	client := &http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	req, err := http.NewRequestWithContext(ctx, "POST", url,
		strings.NewReader(json_text),
	)
	if err == nil {
		req.SetBasicAuth(settings.GetWebConn().User, settings.GetWebConn().Api_id)
		req.Header.Add("Content-Type", "application/json")
	}

	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	defer cancel()

	resp, err := client.Do(req)
	if err != nil {

		logRequest(url, json_text, oerr, method, aid)
		if LOGGING {
			fmt.Printf("Marshal: %v", err.Error())
		}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil && LOGGING {
		fmt.Printf("ReadAll: %v", err.Error())
	}

	var result GeneralSubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil && LOGGING {
		fmt.Printf("Unmarshal: %v", err.Error())
	}
	if LOGGING {
		fmt.Println("Retry...result")
		fmt.Println(result)
	}
	//Don't save unless we don't have an id and there is one returned from the webapi
	if !result.Success && result.Error == "" {
		logRequest(url, json_text, oerr, method, aid)
	}

	return
}

// check if the are pending failed requests.
func CheckPending() string {

	db, err := sql.Open("sqlite3", "./pong.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//see if registered, if not return false.
	var (
		count string
	)
	err = db.QueryRow("SELECT COUNT(*) FROM pending").Scan(&count)
	switch {
	case err != nil:
		return "0"
	}
	fmt.Println(count + " Pending Failed Web Calls")
	return count
}

func NewCustomPOST(data []byte, api_url string) (req *http.Request, cancel context.CancelFunc, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	req, err = http.NewRequestWithContext(ctx, "POST", api_url,
		strings.NewReader(string(data)),
	)
	if err == nil {
		req.SetBasicAuth(settings.GetWebConn().User, settings.GetWebConn().Api_id)
		req.Header.Add("Content-Type", "application/json")
	}
	return req, cancel, err
}

// Only logging this for now...
// Sends new transaction to a website when uuid is selected and the out_message contains the url (as of now...)
func SendNewTx(tx map[string]any) string {
	api_url := tx["api_url"].(string)
	_, err := url.ParseRequestURI(api_url)
	if err != nil {
		return ""
	}

	client := &http.Client{}
	var newTx NewTx
	if settings.GetNewTxSettings().Send_uuid {
		newTx.Uuid = tx["uuid"].(string)
	}
	if settings.GetNewTxSettings().Send_ia_id {
		newTx.Ia_id = tx["for_ia_id"].(int)
	}
	data, err := json.Marshal(map[string]interface{}{
		"method": "newTX",
		"params": newTx,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	req, cancel, err := NewCustomPOST(data, api_url)
	defer cancel()
	if err != nil {
		return "New request W/ context: \n" + err.Error()
	}

	resp, err3 := client.Do(req)
	if err3 != nil {
		//json_str, _ := json.Marshal(data)
		r := rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64())))
		logRequest(api_url, string(data), err3.Error(), "newTX", strconv.Itoa(r.Intn(10000000)))
		return "Post: " + err3.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ReadAll: " + err.Error()
	}

	var result NewTxSubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "Unmarshal: " + err.Error()
	}
	if LOGGING {
		fmt.Println("Submit New TX Result -")
		fmt.Println(result)
	}
	if !result.Success && result.Error == "" {
		//json_str, _ := json.Marshal(data)

		r := rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64())))
		logRequest(api_url, string(data), "API Error", "newTX", strconv.Itoa(r.Intn(10000000)))
	}

	return ""

}
