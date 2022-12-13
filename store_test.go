package main

import "testing"

func TestStoreSingle(t *testing.T) {
	s := NewStore()

	s.Upsert(Record{Number: 5, Name: "Test"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v := s.asmap["test"]
	if v.ASN != 5 {
		t.Fatal()
	}
	if v.Name != "Test" {
		t.Fatal()
	}
}

func TestStoreSingleIncrement(t *testing.T) {
	s := NewStore()

	s.Upsert(Record{Number: 5, Name: "Test"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v := s.asmap["test"]
	if v.ASN != 5 {
		t.Fatal()
	}
	if v.Name != "Test" {
		t.Fatal()
	}
	if v.Count != 1 {
		t.Fatal()
	}
	s.Upsert(Record{Number: 5, Name: "Test"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v = s.asmap["test"]
	if v.ASN != 5 {
		t.Fatal()
	}
	if v.Name != "Test" {
		t.Fatal()
	}
	if v.Count != 2 {
		t.Fatal()
	}
}

func TestStoreMigrateGroup(t *testing.T) {
	s := NewStore()

	s.Upsert(Record{Number: 5, Name: "Test"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v := s.asmap["test"]
	if v.ASN != 5 {
		t.Fatal()
	}
	if v.Name != "Test" {
		t.Fatal()
	}
	s.Upsert(Record{Number: 7, Name: "Test"})
	s.Upsert(Record{Number: 7, Name: "Test"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v = s.asmap["test"]
	if v.ASN != 0 {
		t.Fatal()
	}
	if v.Name != "Test" {
		t.Fatal()
	}
	if v.Count != 3 {
		t.Log(v.Count)
		t.Fatal()
	}
	if len(v.Children) != 2 {
		t.Fatal()
	}
	ch1 := v.Children[5]
	if ch1.ASN != 5 {
		t.Log(ch1.ASN)
		t.Fatal()
	}
	if ch1.Name != "Test" {
		t.Fatal()
	}
	if ch1.Count != 1 {
		t.Fatal()
	}
	ch2 := v.Children[7]
	if ch2.ASN != 7 {
		t.Fatal()
	}
	if ch2.Name != "Test" {
		t.Fatal()
	}
	if ch2.Count != 2 {
		t.Fatal()
	}
}

func TestStorePredefinedGroup(t *testing.T) {
	s := NewStore()

	s.Upsert(Record{Number: 24940, Name: "Hetzner Online GmbH"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v := s.asmap["hetzner"]
	if v.ASN != 24940 {
		t.Fatal()
	}
	if v.Name != "Hetzner Online GmbH" {
		t.Fatal()
	}
	s.Upsert(Record{Number: 213230, Name: "Hetzner Online GmbH"})
	if len(s.asmap) != 1 {
		t.Fatal()
	}
	v = s.asmap["hetzner"]
	if v.Name != "Hetzner" {
		t.Fatal()
	}
	if v.ASN != 0 {
		t.Fatal()
	}
	if len(v.Children) != 2 {
		t.Fatal()
	}
	ch1 := v.Children[24940]
	if ch1.ASN != 24940 {
		t.Fatal()
	}
	if ch1.Name != "Hetzner Online GmbH" {
		t.Fatal()
	}
	if ch1.Count != 1 {
		t.Fatal()
	}
	if ch1.Children != nil {
		t.Fatal()
	}
	ch2 := v.Children[213230]
	if ch2.ASN != 213230 {
		t.Fatal()
	}
	if ch2.Name != "Hetzner Online GmbH" {
		t.Fatal()
	}
	if ch2.Count != 1 {
		t.Fatal()
	}
	if ch1.Children != nil {
		t.Fatal()
	}
}

func TestStoreAsList(t *testing.T) {
	s := NewStore()
	s.Upsert(Record{Number: 24940, Name: "Hetzner Online GmbH"})
	s.Upsert(Record{Number: 213230, Name: "Hetzner Online GmbH"})
	s.Upsert(Record{Number: 5, Name: "Test"})

	l := s.AsList()
	if len(l) != 2 {
		t.Fatal()
	}
	elem1 := l[0]
	if elem1.ASN != 0 {
		t.Fatal()
	}
	if elem1.Name != "Hetzner" {
		t.Fatal()
	}
	if len(elem1.Children) != 2 {
		t.Fatal()
	}
	elem2 := l[1]
	if elem2.ASN != 5 {
		t.Fatal()
	}
	if elem2.Name != "Test" {
		t.Fatal()
	}
	if elem2.Children != nil {
		t.Fatal()
	}
}
