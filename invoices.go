package wefact

import "net/url"

func (c *Client) GetInvoice(code string) (results map[string]interface{}, err error) {
	var data = url.Values{}
	data.Add("InvoiceCode", code)
	err = c.Request("invoice", "show", data, &results)
	return
}
