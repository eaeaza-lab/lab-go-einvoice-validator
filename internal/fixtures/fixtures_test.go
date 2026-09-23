package fixtures

import (
	"reflect"
	"testing"

	"example.com/einvoice-validator/internal/invoice"
)

func TestFixtures(t *testing.T) {
	all, err := All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 4 {
		t.Fatalf("expected at least 4 fixtures, got %d", len(all))
	}
	for _, f := range all {
		t.Run(f.Name, func(t *testing.T) {
			inv, err := invoice.ParseJSON(f.Data)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got := []string{}
			for _, d := range invoice.Validate(inv) {
				got = append(got, d.Code)
			}
			want := f.Expected
			if want == nil {
				want = []string{}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("diagnostic codes = %v, want %v", got, want)
			}
		})
	}
}
