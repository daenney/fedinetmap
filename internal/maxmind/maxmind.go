package maxmind

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Entry struct {
	Number uint   `maxminddb:"autonomous_system_number"`
	Name   string `maxminddb:"autonomous_system_organization"`
}

type DB struct {
	mm *maxminddb.Reader
}

func New(path string) (*DB, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}

	return &DB{mm: db}, nil
}

func (d *DB) Lookup(ip net.IP) (Entry, error) {
	var e Entry
	err := d.mm.Lookup(ip, &e)
	return e, err
}

func (d *DB) Close() error {
	return d.mm.Close()
}
