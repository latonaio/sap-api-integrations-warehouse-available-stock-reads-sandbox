package sap_api_caller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	sap_api_output_formatter "sap-api-integrations-warehouse-available-stock-reads/SAP_API_Output_Formatter"
	"strings"
	"sync"

	"github.com/latonaio/golang-logging-library-for-sap/logger"
	"golang.org/x/xerrors"
)

type SAPAPICaller struct {
	baseURL string
	apiKey  string
	log     *logger.Logger
}

func NewSAPAPICaller(baseUrl string, l *logger.Logger) *SAPAPICaller {
	return &SAPAPICaller{
		baseURL: baseUrl,
		apiKey:  GetApiKey(),
		log:     l,
	}
}

func (c *SAPAPICaller) AsyncGetWarehouseAvailableStock(eWMWarehouse, product string, accepter []string) {
	wg := &sync.WaitGroup{}
	wg.Add(len(accepter))
	for _, fn := range accepter {
		switch fn {
		case "Header":
			func() {
				c.Header(eWMWarehouse, product)
				wg.Done()
			}()
		default:
			wg.Done()
		}
	}
	wg.Wait()
}

func (c *SAPAPICaller) Header(eWMWarehouse, product string) {
	data, err := c.callWarehouseAvailableStockSrvAPIRequirementHeader("WarehouseAvailableStock", eWMWarehouse, product)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(data)

}

func (c *SAPAPICaller) callWarehouseAvailableStockSrvAPIRequirementHeader(api, eWMWarehouse, product string) (*sap_api_output_formatter.Header, error) {
	url := strings.Join([]string{c.baseURL, "api_whse_availablestock/srvd_a2x/sap/warehouseavailablestock/0001", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithHeader(req, eWMWarehouse, product)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToHeader(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) setHeaderAPIKeyAccept(req *http.Request) {
	req.Header.Set("APIKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (c *SAPAPICaller) getQueryWithHeader(req *http.Request, eWMWarehouse, product string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("eWMWarehouse eq '%s' and product eq '%s'", eWMWarehouse, product))
	req.URL.RawQuery = params.Encode()
}
