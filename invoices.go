package wefact

import "net/url"

func (c *Client) GetInvoice(code string) (results map[string]interface{}, err error) {
	var data = url.Values{}
	data.Add("InvoiceCode", code)
	response, err := c.Request("invoice", "show", data)
	_ = response
	return
}
