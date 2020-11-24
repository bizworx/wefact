package wefact

import "net/url"

func (c *Client) GetDebtor(code string) (results map[string]interface{}, err error) {
	var data = url.Values{}
	data.Add("DebtorCode", code)
	err = c.Request("debtor", "show", data, &results)
	return
}
