package wefact

import "net/url"

func (c *Client) GetDebtor(code string) (results map[string]interface{}, err error) {
	var data = url.Values{}
	data.Add("DebtorCode", code)
	response, err := c.Request("debtor", "show", data)
	_ = response
	return
}

func (c *Client) UpdateDebtor(code string, data url.Values) (err error) {
	data.Add("DebtorCode", code)
	response, err := c.Request("debtor", "edit", data)
	_ = response
	return
}
