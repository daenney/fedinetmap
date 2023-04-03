# Fedinetmap

⚠️ Please note that the code is very much a work in progress. It's not meant to
be pretty, planet-scale or high architecture. It just loops over some stuff to
collate a bunch of data ⚠️

This uses the ActivityPub instances list at [instances.social][ins] to map the
instance to an Autonomous Network (an ISP). It makes an attempt to group and
dedup networks but the data available on this is a nightmare of a mess.

By default we omit any network or network group that hosts less than 5 instances.
The aim of publishing this set is it roughly understand where the fediverse is
hosted and for that purpose this is good enough.

## Usage

You need a local copy of the [MaxMind ASN database][masn]. You'll also need an
API key for [instances.social][ins]

[ins]: https://instances.social/
[masn]: https://dev.maxmind.com/geoip/docs/databases/asn

```
Usage of ./fedinetmap:
  -db.path string
    	Path to the a MaxMind database with ASN info
  -instances uint
    	amount of instances to fetch (default 10)
  -min-instances uint
    	lower threshold to be included in the output (default 5)
```

## Building

You'll need Go 1.20. `go build` will do the rest.

## Testing

```
go test ./...
```

## Contributing

Right now contributing to the code is probably a waste of your own time until
this is turned into something more sensible. But feel free to send PRs for
the `enduser`, `asnToKey` and `asnToName` maps in `store.go` to help better
name and classify the networks.
