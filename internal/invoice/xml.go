package invoice

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// xmlChildren lists the elements allowed under each element path.
var xmlChildren = map[string][]string{
	"invoice":            {"id", "currency", "lines", "net", "tax", "gross"},
	"invoice/lines":      {"line"},
	"invoice/lines/line": {"description", "quantity", "unit_price", "tax_rate_bp"},
}

// ParseXML decodes an XML invoice document, rejecting unknown elements,
// attributes and content after the root element.
func ParseXML(data []byte) (Invoice, error) {
	if err := checkXMLShape(data); err != nil {
		return Invoice{}, fmt.Errorf("invalid XML invoice: %w", err)
	}
	var inv Invoice
	if err := xml.Unmarshal(data, &inv); err != nil {
		return Invoice{}, fmt.Errorf("invalid XML invoice: %w", err)
	}
	inv.XMLName = xml.Name{}
	return inv, nil
}

// checkXMLShape walks the token stream once so that typos in element names
// surface as errors, mirroring the JSON loader's unknown-field rule.
func checkXMLShape(data []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var stack []string
	rootSeen := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			if len(stack) == 0 {
				if rootSeen {
					return fmt.Errorf("unexpected content after document")
				}
				if name != "invoice" {
					return fmt.Errorf("unexpected root element <%s>, want <invoice>", name)
				}
				rootSeen = true
			} else {
				parent := strings.Join(stack, "/")
				if !contains(xmlChildren[parent], name) {
					return fmt.Errorf("unknown element <%s> in <%s>", name, stack[len(stack)-1])
				}
			}
			if len(t.Attr) > 0 {
				return fmt.Errorf("unexpected attribute %q on <%s>", t.Attr[0].Name.Local, name)
			}
			stack = append(stack, name)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 && len(bytes.TrimSpace(t)) > 0 {
				return fmt.Errorf("unexpected content outside the document")
			}
		}
	}
	if !rootSeen {
		return fmt.Errorf("no <invoice> element")
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
