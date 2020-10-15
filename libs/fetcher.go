package libs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var concurrentGoroutines = make(chan struct{}, 1)

var Data = Result{}

type FetchType int

const (
	FetchTypeAll FetchType = iota
	FetchTypeVendor
	FetchTypeCatalogs
	FetchTypeItems
)

const key = "6953-oaypsHZN88GndHcNtyVBktnyy62VPAjK4qYAT9ga3XEhm6QytSMSjHqLvweL73yhGBerV9x8mEVvBt3A"

type Result struct {
	Response struct {
		Page struct {
			Current int `json:"current"`
			Next    int `json:"next"`
			Prev    int `json:"prev"`
			Pages   int `json:"pages"`
			Items   int `json:"items"`
		} `json:"page"`
		Vendors []struct {
			Name  string `json:"name"`
			Alias string `json:"alias"`
		} `json:"vendors"`
		Catalogs []struct {
			ID       string `json:"va_catalog_id"`
			ParentID string `json:"va_parent_id"`
			Name     string `json:"name"`
		} `json:"catalogs"`
		Items []Item `json:"items"`
	} `json:"response"`
}

type Item struct {
	Images     []string `json:"images"`
	Code       string   `json:"p_code"`
	Mog        string   `json:"mog"`
	OEMNum     string   `json:"oem_num"`
	OEMBrand   string   `json:"oem_brand"`
	Name       string   `json:"name"`
	Shipment   int      `json:"shipment"`
	Delivery   int      `json:"delivery"`
	Department string   `json:"department"`
	Count      int      `json:"count"`
	CountChel  int      `json:"count_chel"`
	CountEkb   int      `json:"count_ekb"`
	UnitCode   int      `json:"unit_code"`
	Unit       string   `json:"unit"`
	Price      float32  `json:"price"`
	CatalogID  string   `json:"va_catalog_id"`
	ItemID     string   `json:"va_item_id"`
}

func FetchResult(tp FetchType, page int, items ...chan *Item) (result Result, err error) {
	var ch chan *Item
	if len(items) > 0 {
		ch = items[0]
	}

	url := "https://api.v-avto.ru/v1/"

	switch tp {
	case FetchTypeAll:
		vendors, _ := FetchResult(FetchTypeVendor, 0)

		catalogs, _ := FetchResult(FetchTypeCatalogs, 1)

		items, _ := FetchResult(FetchTypeItems, 0)

		result.Response.Vendors = vendors.Response.Vendors
		result.Response.Catalogs = catalogs.Response.Catalogs
		result.Response.Items = items.Response.Items
		return result, err
	case FetchTypeVendor:
		url += "vendors"
	case FetchTypeCatalogs:
		url += "catalogs"
		page = 1
	case FetchTypeItems:
		url += "items"
	}

	if page == 0 {
		rs, _ := FetchResult(tp, 1)
		log.Printf("Total pages: %d", rs.Response.Page.Pages)
		var wg sync.WaitGroup

		for i := 1; i < rs.Response.Page.Pages; i++ {
			concurrentGoroutines <- struct{}{}

			wg.Add(1)

			go func(indx int, wgr *sync.WaitGroup) {
				res, err := FetchResult(tp, indx)
				if err == nil {
					switch tp {
					case FetchTypeVendor:
						result.Response.Vendors = append(result.Response.Vendors, res.Response.Vendors...)
					case FetchTypeCatalogs:
						result.Response.Catalogs = append(result.Response.Catalogs, res.Response.Catalogs...)
					case FetchTypeItems:
						result.Response.Items = append(result.Response.Items, res.Response.Items...)
						if ch != nil {
							go func(itm []Item) {
								for _, p := range res.Response.Items {
									ch <- &p
								}
							}(res.Response.Items)
						}
					}
				}

				<-concurrentGoroutines
				defer func() {
					wgr.Done()
				}()

			}(i, &wg)
		}

		wg.Wait()

	} else {
		reqUrl := fmt.Sprintf("%s?key=%s&page=%d", url, key, page)
		//log.Printf("Req: %s",reqUrl)
		req, err := http.NewRequest("GET", reqUrl, nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("User-Agent", fmt.Sprintf("%d", rand.Int()))

		c := http.Client{}

		resp, err := c.Do(req)
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			time.Sleep(500 * time.Microsecond)
			return FetchResult(tp, page)
		}

		switch tp {
		case FetchTypeVendor:
			Data.Response.Vendors = append(Data.Response.Vendors, result.Response.Vendors...)
		case FetchTypeCatalogs:
			Data.Response.Catalogs = append(Data.Response.Catalogs, result.Response.Catalogs...)
		case FetchTypeItems:
			Data.Response.Items = append(Data.Response.Items, result.Response.Items...)
		}
	}

	log.Printf("V: %d, C: %d, I: %d", len(Data.Response.Vendors), len(Data.Response.Catalogs), len(Data.Response.Items))

	return result, err
}
