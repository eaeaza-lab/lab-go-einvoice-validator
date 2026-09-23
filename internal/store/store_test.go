package store

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestAddAndListNewestFirst(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	t0 := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	if _, err := s.Add(Run{Time: t0, Files: []string{"a.json"}, Valid: true}); err != nil {
		t.Fatal(err)
	}
	id, err := s.Add(Run{Time: t0.Add(time.Minute), Files: []string{"b.json", "c.xml"}, Diagnostics: 3})
	if err != nil {
		t.Fatal(err)
	}
	runs, err := s.List(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 || runs[0].ID != id {
		t.Fatalf("want newest first, got %+v", runs)
	}
	if !reflect.DeepEqual(runs[0].Files, []string{"b.json", "c.xml"}) || runs[0].Valid || runs[0].Diagnostics != 3 {
		t.Errorf("unexpected run %+v", runs[0])
	}
	if !runs[0].Time.Equal(t0.Add(time.Minute)) || !runs[1].Valid {
		t.Errorf("unexpected run data %+v", runs)
	}
}

func TestListLimit(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for i := 0; i < 3; i++ {
		if _, err := s.Add(Run{Files: []string{"x.json"}, Valid: true}); err != nil {
			t.Fatal(err)
		}
	}
	runs, err := s.List(2)
	if err != nil || len(runs) != 2 {
		t.Fatalf("got %d runs, err %v", len(runs), err)
	}
}

func TestPersistsAcrossOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(Run{Files: []string{"a.json"}, Valid: true}); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runs, err := s.List(0)
	if err != nil || len(runs) != 1 {
		t.Fatalf("got %d runs, err %v", len(runs), err)
	}
}

func TestEmptyList(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runs, err := s.List(0)
	if err != nil || runs == nil || len(runs) != 0 {
		t.Fatalf("want empty non-nil slice, got %v, %v", runs, err)
	}
}
