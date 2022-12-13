package main

import (
	"fmt"
	"strings"
	"sync"
)

type Entry struct {
	Name     string          `json:"name"`
	Count    uint            `json:"count"`
	ASN      uint            `json:"asn,omitempty"`
	EndUser  bool            `json:"endUser,omitempty"`
	Children map[uint]*Entry `json:"children,omitempty"`
}

type Entries []Entry

func (e Entries) Len() int           { return len(e) }
func (e Entries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e Entries) Less(i, j int) bool { return e[i].Count < e[j].Count }

type Store struct {
	asmap map[string]*Entry
	sync.Mutex
}

func (s *Store) AsList() Entries {
	s.Lock()
	defer s.Unlock()
	res := make(Entries, 0, len(s.asmap))
	for _, ent := range s.asmap {
		res = append(res, *ent)
	}
	return res
}

func (s *Store) Upsert(rec Record) {
	s.Lock()
	defer s.Unlock()

	name := KeyForASN(rec)
	key := strings.ToLower(name)

	if val, ok := s.asmap[key]; ok {
		if val.ASN != rec.Number {
			if val.Children == nil {
				s.migrateToGroup(key, name, rec)
			}
			s.updateGroup(key, rec)
			return
		}
		s.update(key)
		return
	}

	s.create(key, rec)
}

func (s *Store) update(asn string) {
	s.asmap[asn].Count++
}

func (s *Store) create(key string, rec Record) {
	s.asmap[key] = &Entry{
		Name:    NameForASN(rec),
		ASN:     rec.Number,
		Count:   1,
		EndUser: IsLikelyEnduser(rec.Number),
	}
}

func (s *Store) migrateToGroup(key, name string, rec Record) {
	val := s.asmap[key]

	s.asmap[key] = &Entry{
		Name:  name,
		Count: val.Count,
		Children: map[uint]*Entry{
			val.ASN: val,
		},
	}
}

func (s *Store) updateGroup(key string, rec Record) {
	parent := s.asmap[key]

	if ch, ok := parent.Children[rec.Number]; ok {
		ch.Count++
		parent.Count++
		return
	}

	parent.Children[rec.Number] = &Entry{
		Name:    NameForASN(rec),
		ASN:     rec.Number,
		Count:   1,
		EndUser: IsLikelyEnduser(rec.Number),
	}
	parent.Count++
}

func NewStore() *Store {
	return &Store{
		asmap: make(map[string]*Entry),
	}
}

var enduser = map[uint]struct{}{
	35540: {},
	6871:  {},
	8447:  {},
	41164: {},
	3249:  {},
	49455: {},
	13285: {},
	12334: {},
	12946: {},
	29518: {},
	45011: {},
	35244: {},
	16202: {},
	15542: {},
	15435: {},
	16246: {},
	16591: {},
	22773: {},
	50266: {},
	2119:  {},
	54858: {},
}

func IsLikelyEnduser(asn uint) bool {
	_, ok := enduser[asn]
	return ok
}

var asnToKey = map[uint]string{
	16276:  "OVH",
	35540:  "OVH",
	24940:  "Hetzner",
	213230: "Hetzner",
	212317: "Hetzner",
	16509:  "Amazon Web Services",
	14618:  "Amazon Web Services",
	8068:   "Microsoft",
	8069:   "Microsoft",
	8075:   "Microsoft",
	8972:   "GoDaddy",
	20773:  "GoDaddy",
	21499:  "GoDaddy",
	34011:  "GoDaddy",
	398101: "GoDaddy",
	20857:  "TransIP",
	35470:  "TransIP",
	57370:  "Swisscom",
	3303:   "Swisscom",
	41164:  "Telia",
	3308:   "Telia",
	3249:   "Telia",
	1759:   "Telia",
	12582:  "Telia",
	25400:  "Telia",
	12929:  "Telia",
	49455:  "Telia",
	3301:   "Telia",
	34610:  "Telia",
	39642:  "Telia",
	8764:   "Telia",
	5378:   "Vodafone",
	1273:   "Vodafone",
	15502:  "Vodafone",
	3209:   "Vodafone",
	33915:  "Vodafone",
	25310:  "Vodafone",
	16019:  "Vodafone",
	30722:  "Vodafone",
	12353:  "Vodafone",
	12430:  "Vodafone",
	6739:   "Vodafone",
	60781:  "LeaseWeb",
	30633:  "LeaseWeb",
	28753:  "LeaseWeb",
	7203:   "LeaseWeb",
	396362: "LeaseWeb",
	59253:  "LeaseWeb",
	396982: "Google",
	15169:  "Google",
	16591:  "Google",
	4713:   "NTT",
	203329: "NTT",
	2514:   "NTT",
	8412:   "T-Mobile",
	50266:  "T-Mobile",
	13127:  "T-Mobile",
	13036:  "T-Mobile",
	132203: "Tencent",
	45090:  "Tencent",
	202053: "UpCloud",
	25697:  "UpCloud",
	2119:   "Telenor",
	9158:   "Telenor",
	22927:  "Telefonica",
	18881:  "Telefonica",
	6805:   "Telefonica",
	3352:   "Telefonica",
	7418:   "Telefonica",
	209:    "Lumen Technologies",
	203:    "Lumen Technologies",
	3549:   "Lumen Technologies",
	9050:   "Orange",
	3215:   "Orange",
	12479:  "Orange",
	47377:  "Orange",
	5617:   "Orange",
	10796:  "Charter",
	11426:  "Charter",
	11427:  "Charter",
	33363:  "Charter",
	20001:  "Charter",
	11351:  "Charter",
	20115:  "Charter",
	12271:  "Charter",
	204548: "Kamatera",
	210329: "Kamatera",
	41436:  "Kamatera",
	36007:  "Kamatera",
	54858:  "Astound",
	29962:  "Astound",
	11404:  "Astound",
	7922:   "Comcast",
	13367:  "Comcast",
}

var asnToName = map[uint]string{
	16276:  "OVHcloud",
	35540:  "OVH Télécom",
	20473:  "Vultr",
	31898:  "Oracle Cloud",
	205766: "Uberspace",
	29802:  "HIVELOCITY, Inc.",
	35916:  "MULTACOM CORPORATION",
	22773:  "Cox Communications Inc.",
	40676:  "Psychz Networks",
	3842:   "InMotion Hosting, Inc.",
	30600:  "Metronet",
	19318:  "Interserver, Inc",
	855:    "Bell Canada",
	6128:   "Cablevision Systems Corp.",
	12033:  "Adams TelSystems, Inc.",
	393552: "NextLight",
	6871:   "Plusnet",
	224:    "Uninett",
	395748: "NetSpeed LLC",
	231:    "Michigan State University",
	33070:  "Rackspace Hosting",
	13354:  "zColo",
	8298:   "IPng Networks",
	17054:  "Expedient",
	62904:  "ServerHub",
	20454:  "SecuredServers",
	131:    "University of California, Santa Barbara",
}

func KeyForASN(rec Record) string {
	if val, ok := asnToKey[rec.Number]; ok {
		return val
	}
	return rec.Name
}

func NameForASN(rec Record) string {
	if val, ok := asnToName[rec.Number]; ok {
		return fmt.Sprintf("%s (%s)", val, rec.Name)
	}
	return rec.Name
}
