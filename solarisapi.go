package solarisapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ProductDetails struct {
	Product struct {
		ID                int64       `json:"id"`
		Title             string      `json:"title"`
		BodyHTML          string      `json:"body_html"`
		Vendor            string      `json:"vendor"`
		ProductType       string      `json:"product_type"`
		CreatedAt         time.Time   `json:"created_at"`
		Handle            string      `json:"handle"`
		UpdatedAt         time.Time   `json:"updated_at"`
		PublishedAt       time.Time   `json:"published_at"`
		TemplateSuffix    interface{} `json:"template_suffix"`
		Status            string      `json:"status"`
		PublishedScope    string      `json:"published_scope"`
		Tags              string      `json:"tags"`
		AdminGraphqlAPIID string      `json:"admin_graphql_api_id"`
		Variants          []struct {
			ID                   int64       `json:"id"`
			Title                string      `json:"title"`
			Price                string      `json:"price"`
			Sku                  string      `json:"sku"`
			Position             int         `json:"position"`
			InventoryPolicy      string      `json:"inventory_policy"`
			CompareAtPrice       interface{} `json:"compare_at_price"`
			FulfillmentService   string      `json:"fulfillment_service"`
			InventoryManagement  string      `json:"inventory_management"`
			Option1              string      `json:"option1"`
			Option2              interface{} `json:"option2"`
			Option3              interface{} `json:"option3"`
			CreatedAt            time.Time   `json:"created_at"`
			UpdatedAt            time.Time   `json:"updated_at"`
			Taxable              bool        `json:"taxable"`
			Barcode              string      `json:"barcode"`
			Grams                int         `json:"grams"`
			ImageID              interface{} `json:"image_id"`
			Weight               float64     `json:"weight"`
			WeightUnit           string      `json:"weight_unit"`
			InventoryItemID      int64       `json:"inventory_item_id"`
			InventoryQuantity    int         `json:"inventory_quantity"`
			OldInventoryQuantity int         `json:"old_inventory_quantity"`
			RequiresShipping     bool        `json:"requires_shipping"`
			AdminGraphqlAPIID    string      `json:"admin_graphql_api_id"`
		} `json:"variants"`
		Options []struct {
			ID        int64    `json:"id"`
			ProductID int64    `json:"product_id"`
			Name      string   `json:"name"`
			Position  int      `json:"position"`
			Values    []string `json:"values"`
		} `json:"options"`
		Images []struct {
			ID                int64         `json:"id"`
			Position          int           `json:"position"`
			CreatedAt         time.Time     `json:"created_at"`
			UpdatedAt         time.Time     `json:"updated_at"`
			Alt               interface{}   `json:"alt"`
			Width             int           `json:"width"`
			Height            int           `json:"height"`
			Src               string        `json:"src"`
			VariantIds        []interface{} `json:"variant_ids"`
			AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
		} `json:"images"`
		Image struct {
			ID                int64         `json:"id"`
			Position          int           `json:"position"`
			CreatedAt         time.Time     `json:"created_at"`
			UpdatedAt         time.Time     `json:"updated_at"`
			Alt               interface{}   `json:"alt"`
			Width             int           `json:"width"`
			Height            int           `json:"height"`
			Src               string        `json:"src"`
			VariantIds        []interface{} `json:"variant_ids"`
			AdminGraphqlAPIID string        `json:"admin_graphql_api_id"`
		} `json:"image"`
	} `json:"product"`
}

const api = "https://json.solarisjapan.com/products/"
const currAPIURL = "https://cdn.shopify.com/s/javascripts/currencies.js"

var reCurr = regexp.MustCompile(`(\w+):(\d*\.*\d+)`)
var reHandle = regexp.MustCompile(`https://solarisjapan\.com.*/products/(.+)`)

// GetItemByURL and return it's data as a struct
func GetItemByURL(URL string) (ProductDetails, error) {
	return GetItemByHandle(ParseHandleFromURL(URL))
}

// GetItemByHandle and return it's data as a struct
func GetItemByHandle(handle string) (ProductDetails, error) {

	data, err := get(api + handle)
	if err != nil {
		return ProductDetails{}, err
	}

	productDetails := ProductDetails{}
	err = json.Unmarshal(data, &productDetails)
	if err != nil {
		return ProductDetails{}, err
	}

	return productDetails, nil
}

// GetCurrencies rates of the important currencies
func GetCurrencies() (map[string]float64, error) {

	data, err := get(currAPIURL)
	if err != nil {
		return nil, err
	}

	matchedCurr := reCurr.FindAllStringSubmatch(string(data), -1) //1=currencyKey 2=qouta
	if len(matchedCurr) < 1 {
		return nil, errors.New("no currencies found")
	}

	currencies := map[string]float64{}
	for _, match := range matchedCurr {
		currRate := parseCurrencyRate(match[2])
		if currRate == 0 {
			continue
		}

		currencies[match[1]] = currRate
	}

	return currencies, nil
}

// ParseHandleFromURL thats needed for the API request
func ParseHandleFromURL(URL string) string {
	matchedHandle := reHandle.FindStringSubmatch(URL) //1=handle
	if len(matchedHandle) < 1 {
		return ""
	}

	return matchedHandle[1]
}

func get(URL string) ([]byte, error) {

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression:  true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			IdleConnTimeout:     5 * time.Second,
		},
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}

	return body, nil

}

func parseCurrencyRate(rate string) float64 {
	if strings.HasPrefix(rate, ".") {
		rate = "0" + rate
	}

	currRate, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return 0
	}

	return currRate
}
