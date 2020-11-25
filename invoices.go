package wefact

import "net/url"

func (c *Client) GetInvoice(code string) (*Response, error) {
	var data = url.Values{}
	data.Add("InvoiceCode", code)
	return c.Request("invoice", "show", data)
}

func (c *Client) CreateInvoice(code string, data url.Values) (*Response, error) {
	data.Add("DebtorCode", code)
	return c.Request("invoice", "add", data)
}
