package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"

	"code.dny.dev/fedinetmap/internal/dns"
	"code.dny.dev/fedinetmap/internal/instances"
	"code.dny.dev/fedinetmap/internal/maxmind"
	"code.dny.dev/fedinetmap/internal/store"
)

func main() {
	dbPath := flag.String("db.path", "", "Path to the a MaxMind database with ASN info")
	numInstances := flag.Uint("instances", 10, "amount of instances to fetch")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	token := os.Getenv("FEDINETMAP_INSTANCES_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "FEDINETMAP_INSTANCES_TOKEN must be set and not be an empty string\n")
		os.Exit(1)
	}

	if *dbPath == "" {
		fmt.Fprintf(os.Stderr, "path to Maxmind database file cannot be empty\n")
		os.Exit(1)
	}

	db, err := maxmind.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open Maxmind DB: %s\n", err)
	}
	defer db.Close()

	instances, err := instances.Get(ctx, http.DefaultClient, instances.Endpoint, token, *numInstances)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to retrieve data from instances.social: %s\n", err)
		db.Close()
		os.Exit(1)
	}

	if len(instances) == 0 {
		fmt.Fprintf(os.Stderr, "instances.social returned no instances\n")
		db.Close()
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "attempting to resolve %d instances\n", len(instances))
	ips := dns.ResolveInstances(instances)
	fmt.Fprintf(os.Stderr, "resolved a total of %d instacnes from original set of %d\n", len(ips), len(instances))
	st := store.New(len(ips))

	for name, ip := range ips {
		e, err := db.Lookup(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to lookup instance: %s with IP: %s in MaxMind DB: %v\n", name, ip.String(), err)
			continue
		}
		if e.Number != 0 {
			st.Upsert(e)
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
