package solarisapi

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"
)

// ProductDetails of solarisjapan.com product
type ProductDetails struct {
	Product struct {
		ID                int64             `json:"id"`
		Title             string            `json:"title"`
		BodyHTML          string            `json:"body_html"`
		Vendor            string            `json:"vendor"`
		ProductType       string            `json:"product_type"`
		CreatedAt         time.Time         `json:"created_at"`
		Handle            string            `json:"handle"`
		UpdatedAt         time.Time         `json:"updated_at"`
		PublishedAt       time.Time         `json:"published_at"`
		TemplateSuffix    interface{}       `json:"template_suffix"`
		Status            string            `json:"status"`
		PublishedScope    string            `json:"published_scope"`
		Tags              string            `json:"tags"`
		AdminGraphqlAPIID string            `json:"admin_graphql_api_id"`
		Info              map[string]string `json:"info"`
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

const api = "https://solarisjapan.com/products/"

var reHandle = regexp.MustCompile(`https://solarisjapan\.com.*/products/([^?]+)`)

// reProductInfo finds product information that is supplied by solarisjapan only and it cannot be obtained via shopify
var reProductInfo = regexp.MustCompile(`<strong> ([^:]+): </strong>\s+<span[^>]*>([^<]+)`) //1=Keyword 2=Info -> 1=Dimensions 2=330.0 mm

// GetItemByURL and return it's data as a struct
func GetItemByURL(URL string) (ProductDetails, error) {
	return GetItemByHandle(ParseHandleFromURL(URL))
}

// GetItemByHandle and return it's data as a struct
func GetItemByHandle(handle string) (ProductDetails, error) {

	productURL := api + handle

	data, err := get(productURL + ".json")
	if err != nil {
		return ProductDetails{}, err
	}

	productDetails := ProductDetails{}
	err = json.Unmarshal(data, &productDetails)
	if err != nil {
		return ProductDetails{}, err
	}

	productHTML, err := get(productURL)
	if err != nil {
		return ProductDetails{}, err
	}

	matchedProductInfo := reProductInfo.FindAllSubmatch(productHTML, -1)
	if len(matchedProductInfo) < 1 {
		// no additional product info found
		return productDetails, nil
	}

	productDetails.Product.Info = make(map[string]string, len(matchedProductInfo))
	for _, match := range matchedProductInfo {
		productDetails.Product.Info[string(match[1])] = string(match[2])
	}

	return productDetails, nil
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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}

	return body, nil

}
