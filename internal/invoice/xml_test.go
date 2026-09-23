package invoice

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"example.com/einvoice-validator/internal/fixtures"
)

const validXML = `<?xml version="1.0"?>
<invoice><id>INV-0001</id><currency>EUR</currency>
<lines><line><description>Widget</description><quantity>2</quantity><unit_price>1000</unit_price><tax_rate_bp>2000</tax_rate_bp></line></lines>
<net>2000</net><tax>400</tax><gross>2400</gross></invoice>`

func TestParseXMLOK(t *testing.T) {
	inv, err := ParseXML([]byte(validXML))
	if err != nil {
		t.Fatal(err)
	}
	if inv.ID != "INV-0001" || len(inv.Lines) != 1 || inv.Lines[0].UnitPrice != 1000 || inv.Gross != 2400 {
		t.Fatalf("unexpected invoice: %+v", inv)
	}
	if d := Validate(inv); len(d) != 0 {
		t.Fatalf("expected valid, got %v", d)
	}
}

func TestParseXMLRejects(t *testing.T) {
	cases := map[string]string{
		"unknown element":   `<invoice><id>A</id><bogus>1</bogus></invoice>`,
		"unknown line elem": `<invoice><lines><line><colour>red</colour></line></lines></invoice>`,
		"attribute":         `<invoice id="A"></invoice>`,
		"wrong root":        `<bill><id>A</id></bill>`,
		"trailing root":     `<invoice><id>A</id></invoice><invoice><id>B</id></invoice>`,
		"trailing text":     `<invoice><id>A</id></invoice> junk`,
		"malformed":         `<invoice><id>A</invoice>`,
		"bad number":        `<invoice><net>abc</net></invoice>`,
		"empty":             ``,
	}
	for name, doc := range cases {
		if _, err := ParseXML([]byte(doc)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if _, err := ParseXML([]byte(`<invoice><bogus/></invoice>`)); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Errorf("expected error naming the element, got %v", err)
	}
}

func TestLoadXMLByExtension(t *testing.T) {
	p := filepath.Join(t.TempDir(), "inv.XML")
	if err := os.WriteFile(p, []byte(validXML), 0o600); err != nil {
		t.Fatal(err)
	}
	inv, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if inv.ID != "INV-0001" {
		t.Fatalf("unexpected invoice: %+v", inv)
	}
}

func TestLoadSampleXML(t *testing.T) {
	inv, err := Load(filepath.Join("..", "..", "testdata", "valid.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if d := Validate(inv); len(d) != 0 {
		t.Fatalf("expected valid, got %v", d)
	}
}

// TestXMLParityWithJSONFixtures re-encodes every JSON fixture as XML and
// checks the loader yields the same invoice and the same diagnostics.
func TestXMLParityWithJSONFixtures(t *testing.T) {
	all, err := fixtures.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range all {
		fromJSON, err := ParseJSON(f.Data)
		if err != nil {
			t.Fatalf("%s: %v", f.Name, err)
		}
		doc, err := xml.MarshalIndent(fromJSON, "", "  ")
		if err != nil {
			t.Fatalf("%s: %v", f.Name, err)
		}
		fromXML, err := ParseXML(doc)
		if err != nil {
			t.Fatalf("%s: %v\n%s", f.Name, err, doc)
		}
		fromJSON.XMLName = xml.Name{}
		if !reflect.DeepEqual(fromJSON, fromXML) {
			t.Errorf("%s: XML round trip differs:\n json %+v\n xml  %+v", f.Name, fromJSON, fromXML)
		}
		if !reflect.DeepEqual(Validate(fromJSON), Validate(fromXML)) {
			t.Errorf("%s: diagnostics differ between JSON and XML", f.Name)
		}
	}
}
