package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bizworx/billingportal"
	"github.com/go-playground/form"
	"github.com/kr/pretty"
	"github.com/pkg/errors"
)

var (
	filename *string        = flag.String("f", "", "specify the path to the billing portal invoices XML")
	encoder  *form.Encoder  = form.NewEncoder()
	client   *wefact.Client = wefact.New(os.Getenv("WEFACT_API_KEY"))
)

func init() {
	flag.Parse()
}

type updateDebtorRequest struct {
	DebtorCode   string
	CustomFields map[string]interface{}
}

type invoiceCreateRequest struct {
	DebtorCode   string
	Description  string
	InvoiceLines []map[string]interface{}
}

func main() {
	if *filename == "" {
		log.Fatalf("the filename cannot be empty")
	}

	invoices, err := readXMLFile(*filename)
	if err != nil {
		log.Fatalf("failed to process XML file: %v", err)
	}

	for _, invoice := range invoices.Invoice {
		response, err := getDebtor(invoice.DebtorNumber)
		if err != nil || response.Status == "error" {
			log.Printf("failed to get debtor with code: %s (%s)", invoice.DebtorNumber, invoice.CustomerName)
			continue
		}

		var debtor = response.Result["debtor"].(map[string]interface{})

		invoiceDate, err := time.Parse("2-1-2006 15:04:05", invoice.Date)
		if err != nil {
			log.Fatalf("failed to extract invoice date: %v %s", err, invoice.Date)
		}
		lastDebtorInvoice := extractVoipitDate(debtor)

		if invoiceDate.Sub(lastDebtorInvoice) > 0 {
			if err := importInvoice(invoice); err != nil {
				log.Printf("failed to import invoice te wefact: %v", err)
				continue
			}
		}
	}

	if err := os.Remove(*filename); err != nil {
		log.Fatalf("failed to remove XML file %v", err)
	}
}

func getDebtor(code string) (*wefact.Response, error) {
	var data = url.Values{}
	data.Add("DebtorCode", code)
	return client.Request("debtor", "show", data)
}

func updateDebtorVoipitDate(code string, date time.Time) error {
	req := updateDebtorRequest{
		DebtorCode:   code,
		CustomFields: map[string]interface{}{"voipit": date.Format("2006-01-02")},
	}
	values, err := encodeRequest(&req)
	if err != nil {
		return err
	}

	response, err := client.Request("debtor", "edit", values)
	if response.Status == "error" {
		result := &bytes.Buffer{}
		pretty.Fprintf(result, "\n%v", response.Result)
		return errors.Wrap(errors.New("failed to update debtor"), result.String())
	}
	return nil
}

func encodeRequest(req interface{}) (url.Values, error) {
	// must pass a pointer
	values, err := encoder.Encode(req)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to encode formdata")
	}
	return values, nil
}

func importInvoice(invoice *billingportal.Invoice) error {
	log.Printf("creating invoice for %s period %s\n", invoice.CustomerName, invoice.Period)

	var formdata = invoiceCreateRequest{}
	formdata.DebtorCode = invoice.DebtorNumber
	formdata.Description = fmt.Sprintf("Verberuikskosten: %s", invoice.Period)
	formdata.InvoiceLines = make([]map[string]interface{}, len(invoice.Specifications.Specification))
	startPeriod, endPeriod, err := extractInvoicePeriod(invoice.Period)
	if err != nil {
		return err
	}
	for idx, spec := range invoice.Specifications.Specification {
		amount, err := strconv.ParseFloat(strings.Replace(spec.Amount, ",", ".", -1), 32)
		if err != nil {
			return err
		}
		if amount > 0 {
			formdata.InvoiceLines[idx] = map[string]interface{}{
				"Date":          time.Now().Format("2006-01-02"),
				"Description":   fmt.Sprintf("Verbruikskosten: %s", spec.Category),
				"ProductCode":   "P0066",
				"PriceExcl":     strings.Replace(spec.Amount, ",", ".", -1),
				"StartDate":     startPeriod.Format("2006-01-02"),
				"EndDate":       endPeriod.Format("2006-01-02"),
				"TaxCode":       "V21",
				"TaxPercentage": "21",
			}
		}
	}

	// must pass a pointer
	values, err := encodeRequest(formdata)
	if err != nil {
		return err
	}

	response, err := client.Request("invoice", "add", values)
	if err != nil {
		return err
	}
	if response.Status == "error" {
		result := &bytes.Buffer{}
		pretty.Fprintf(result, "\n%v", response.Result)
		return errors.Wrap(errors.New("failed to create invoice"), result.String())
	}

	if err := updateDebtorVoipitDate(invoice.DebtorNumber, endPeriod); err != nil {
		return err
	}

	return nil
}

func extractInvoicePeriod(period string) (time.Time, time.Time, error) {
	var months map[string]time.Month = make(map[string]time.Month, 11)
	months["januari"] = time.January
	months["februari"] = time.February
	months["maart"] = time.March
	months["april"] = time.April
	months["mei"] = time.May
	months["juni"] = time.June
	months["juli"] = time.July
	months["augustus"] = time.August
	months["september"] = time.September
	months["oktober"] = time.October
	months["november"] = time.November
	months["december"] = time.December

	s := strings.Split(period, "t/m")
	s[0] = strings.Trim(s[0], " ")
	s[1] = strings.Trim(s[1], " ")

	var toTime = func(input string) (time.Time, error) {
		sin := strings.Split(input, " ")
		day, month, year := sin[0], sin[1], sin[2]

		d, err := strconv.ParseInt(day, 10, 32)
		if err != nil {
			return time.Time{}, errors.WithMessage(err, "failed to parse day from invoice period")
		}
		y, err := strconv.ParseInt(year, 10, 32)
		if err != nil {
			return time.Time{}, errors.WithMessage(err, "failed to parse year from invoice period")
		}
		m := months[strings.ToLower(month)]
		return time.Date(int(y), m, int(d), 0, 0, 0, 0, time.Local), nil

	}

	start, err := toTime(s[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	stop, err := toTime(s[1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, stop, nil
}

func extractVoipitDate(debtor map[string]interface{}) time.Time {
	fields, ok := debtor["CustomFields"].(map[string]interface{})
	val, ok1 := fields["voipit"]
	if !ok || !ok1 {
		return time.Date(2020, 9, 1, 0, 0, 0, 0, time.Local)
	}
	dt, err := time.Parse("2006-01-02", val.(string))
	if err != nil {
		return time.Date(2020, 9, 1, 0, 0, 0, 0, time.Local)
	}

	return dt
}

func readXMLFile(path string) (*billingportal.Invoices, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadFile")
	}

	invoices, err := billingportal.FromXML(b)
	if err != nil {
		return nil, errors.Wrap(err, "billingportal.FromXML")
	}

	return invoices, nil
}
