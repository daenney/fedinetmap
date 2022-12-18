package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"

	"github.com/d3mondev/resolvermt"
	"github.com/oschwald/maxminddb-golang"
	"golang.org/x/exp/slices"
)

const (
	endpoint = "https://instances.social/api/1.0/instances/list?count=%d&include_down=false"
)

func main() {
	dbPath := flag.String("db.path", "", "Path to the a MaxMind database with ASN info")
	numInstances := flag.Uint("instances", 10, "amount of instances to fetch")
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

	d, err := fetchInstances(ctx, http.DefaultClient, token, *numInstances)
	if err != nil {
		panic(err)
	}

	client := resolvermt.New([]string{
		"1.0.0.1",         // Cloudflare
		"1.1.1.1",         // Cloudflare
		"8.8.8.8",         // Google
		"8.8.4.4",         // Google
		"9.9.9.10",        // Quad9
		"149.112.112.112", // Quad9
		"76.76.2.0",       // Control D
		"76.76.10.0",      // Control D
	}, 3, 10, 5)

	instances := make([]string, 0, len(d.Instances))
	for _, ins := range d.Instances {
		// some instances in the DB are malformed, they have a port of even a path
		// attached to them which we try to fix by round-tripping through url.Parse
		u, err := url.Parse("https://" + ins.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping instance: %s, could not parse: %v\n", ins.Name, err)
			continue
		}
		instances = append(instances, u.Hostname())
	}

	ips := make(map[string]net.IP, len(instances))
	hasv4 := []string{}
	fmt.Fprintf(os.Stderr, "resolving %d instances for IPv4\n", len(instances))
	resultsv4 := client.Resolve(instances, resolvermt.TypeA)
	fmt.Fprintf(os.Stderr, "finished resolving %d instances for IPv4\n", len(instances))
	prevDomain := ""
	for _, res := range resultsv4 {
		if res.Type != resolvermt.TypeA {
			continue
		}
		if prevDomain == res.Question {
			continue
		}
		ip := net.ParseIP(res.Answer)
		if !ip.IsGlobalUnicast() {
			fmt.Fprintf(os.Stderr, "got non-unicast IP: %s for: %s\n", ip.String(), res.Question)
			continue
		}
		ips[res.Question] = ip
		hasv4 = append(hasv4, res.Question)
		prevDomain = res.Question
	}

	fmt.Fprintf(os.Stderr, "got %d valid IPs\n", len(ips))

	// assume we've found some IPv6 only instance
	// this may be incorrect if something repeatedly failed to resolve over v4
	// but it's probably a tiny fraction anyway that this doesn't matter
	if len(instances) != len(hasv4) {
		slices.Sort(instances)
		slices.Sort(hasv4)

		v6instances := []string{}
		for _, ins := range instances {
			if _, ok := slices.BinarySearch(hasv4, ins); !ok {
				v6instances = append(v6instances, ins)
			}
		}

		fmt.Fprintf(os.Stderr, "resolving %d instances for IPv6\n", len(v6instances))
		resultsv6 := client.Resolve(v6instances, resolvermt.TypeAAAA)
		fmt.Fprintf(os.Stderr, "finished resolving %d instances for IPv6\n", len(v6instances))
		prevDomain = ""
		for _, res := range resultsv6 {
			if res.Type != resolvermt.TypeAAAA {
				continue
			}
			if prevDomain == res.Question {
				continue
			}
			ip := net.ParseIP(res.Answer)
			if !ip.IsGlobalUnicast() {
				fmt.Fprintf(os.Stderr, "got non-unicast IP: %s for: %s\n", ip.String(), res.Question)
				continue
			}
			ips[res.Question] = ip
			prevDomain = res.Question
		}
	}

	fmt.Fprintf(os.Stderr, "resolved a total of %d instacnes from original set of %d\n", len(ips), len(instances))

	st := NewStore()

	for name, ip := range ips {
		var r Record
		if err := db.Lookup(ip, &r); err != nil {
			fmt.Fprintf(os.Stderr, "failed to lookup entry in MaxMind DB: %v\n", err)
			continue
		}
		if r.Number != 0 {
			st.Upsert(r)
		} else {
			fmt.Fprintf(os.Stderr, "skipping instance %s with IP: %s not found in MaxMind DB\n", name, ip.String())
		}
	}

	all := st.AsList()
	sort.Stable(sort.Reverse(all))

	j := json.NewEncoder(os.Stdout)
	j.SetIndent("", "    ")
	j.Encode(all)
}

func fetchInstances(ctx context.Context, cl *http.Client, token string, instances uint) (*Data, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(endpoint, instances), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "fedinetmap (+https://code.dny.dev/fedinetmap)")

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
type Record struct {
	Number uint   `maxminddb:"autonomous_system_number"`
	Name   string `maxminddb:"autonomous_system_organization"`
}
