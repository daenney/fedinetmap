package dns

import (
	"fmt"
	"net"
	"os"

	"github.com/d3mondev/resolvermt"
)

const (
	Quad91 = "9.9.9.10"
	Quad92 = "149.112.112.112"

	ControlD1 = "76.76.2.0"
	ControlD2 = "76.76.10.0"

	DNS0EU1 = "193.110.81.0"
	DNS0EU2 = "185.253.5.0"
)

var defaultResolver = resolvermt.New(
	[]string{Quad91, Quad92, ControlD1, ControlD2, DNS0EU1, DNS0EU2},
	3, 10, 5,
)

// ResolveInstances resolves the given instances to an IPv4 address. We don't check
// for IPv6 as in order to federate the instance should at least be dual-stack and
// as such an IPv4 address is expected to be available. An IPv6 only instance is
// unlikely to get very far.
func ResolveInstances(instances []string) map[string]net.IP {
	ips := make(map[string]net.IP, len(instances))
	resultsv4 := defaultResolver.Resolve(instances, resolvermt.TypeA)

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
		prevDomain = res.Question
	}

	return ips
}
