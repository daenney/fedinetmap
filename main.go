package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"golang.org/x/exp/slices"
)

const (
	endpoint = "https://instances.social/api/1.0/instances/list?count=0&include_down=false"
)

type ASMap map[string]ASEntry

type ASEntry struct {
	Name      string `json:"name"`
	Count     uint   `json:"count"`
	ASNumbers []uint `json:"asNumbers"`
}

type ASEntryList []ASEntry

func (a ASEntryList) Len() int           { return len(a) }
func (a ASEntryList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ASEntryList) Less(i, j int) bool { return a[i].Count < a[j].Count }

func main() {
	dbPath := flag.String("db.path", "", "Path to the a MaxMind database with ASN info")
	flag.Parse()

	if *dbPath == "" {
		panic("need a database")
	}

	token := os.Getenv("FEDINETMAP_INSTANCES_TOKEN")
	if token == "" {
		panic("need a token for instances.social")
	}

	db, err := maxminddb.Open(*dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	d, err := fetchInstances(ctx, http.DefaultClient, token)
	if err != nil {
		panic(err)
	}

	all := ASMap{}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(5000) * time.Millisecond,
			}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}

	for i, ins := range d.Instances {
		res, err := nameToIP(ctx, ins.Name, resolver)
		if err != nil || res == "" {
			fmt.Fprintf(os.Stderr, "skipping: %s, resolution failed\n", ins.Name)
			continue
		}
		var r Record
		if err := db.Lookup(net.ParseIP(res), &r); err != nil {
			fmt.Fprintf(os.Stderr, "skipping: %s, IP: %s not found in MaxMind DB\n", ins.Name, res)
			continue
		}

		val, ok := all[r.ASName]
		if !ok {
			all[r.ASName] = ASEntry{
				Name:      r.ASName,
				ASNumbers: []uint{r.ASNumber},
				Count:     1,
			}
		} else {
			val.Count++
			if !slices.Contains(val.ASNumbers, r.ASNumber) {
				val.ASNumbers = append(val.ASNumbers, r.ASNumber)
			}
			all[r.ASName] = val
		}
		if i != 0 && i%10 == 0 {
			time.Sleep(5 * time.Second)
		}
	}

	alist := make(ASEntryList, 0, len(all))

	for _, a := range all {
		alist = append(alist, a)
	}
	sort.Stable(sort.Reverse(alist))

	j := json.NewEncoder(os.Stdout)
	j.SetIndent("", "    ")
	j.Encode(alist)
}

func nameToIP(ctx context.Context, name string, resolver *net.Resolver) (string, error) {
	res, err := resolver.LookupIPAddr(ctx, name)
	if err != nil {
		return "", err
	}
	for _, r := range res {
		if !r.IP.IsGlobalUnicast() {
			continue
		}
		return r.String(), nil
	}
	return "", nil
}

func fetchInstances(ctx context.Context, cl *http.Client, token string) (*Data, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "fedimap (+https://code.dny.dev/fedinetmap)")

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var d Data
	j := json.NewDecoder(resp.Body)
	err = j.Decode(&d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

type Data struct {
	Instances []struct {
		Name string `json:"name"`
	} `json:"instances"`
}

type Instances []string

type Record struct {
	ASNumber uint   `maxminddb:"autonomous_system_number"`
	ASName   string `maxminddb:"autonomous_system_organization"`
}
