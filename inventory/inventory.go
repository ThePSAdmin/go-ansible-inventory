package inventory

import (
	"fmt"
	"sync"

	"github.com/spf13/pflag"
)

type Inventory interface {
	AddHost(hostname string) (Host, error)
	GetHost(hostname string) (Host, bool)
	AddGroup(groupname string) (Group, error)
	GetGroup(groupname string) (Group, bool)
	Parse() error
}

type Host interface {
	AddVariable(k string, v string)
}

type Group interface {
	AddHost(hostname string)
	AddVariable(k string, v string)
}

type inventory struct {
	mu     sync.Mutex
	Hosts  map[string]*host
	Groups map[string]*group
}

type host struct {
	mu        sync.Mutex
	variables map[string]string `json:"vars"`
}

type group struct {
	mu        sync.Mutex
	hosts     []string          `json:"hosts"`
	variables map[string]string `json:"vars"`
	children  []string          `json:"children"`
	inventory *inventory        `json:"-"`
}

func NewInventory() Inventory {
	return &inventory{
		Hosts:  make(map[string]*host),
		Groups: map[string]*group{"all": {children: []string{"ungrouped"}}, "ungrouped": {}},
	}
}

func (i *inventory) AddHost(hostname string) (Host, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.Hosts[hostname]; ok {
		err := fmt.Errorf("Hostname %v already exists in inventory", hostname)
		return nil, err
	}
	i.Groups["all"].hosts = append(i.Groups["all"].hosts, hostname)
	i.Groups["ungrouped"].hosts = append(i.Groups["all"].hosts, hostname)
	i.Hosts[hostname] = &host{}
	return i.Hosts[hostname], nil
}

func (i *inventory) GetHost(hostname string) (Host, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if h, ok := i.Hosts[hostname]; ok {
		return h, true
	}
	return nil, false

}

func (i *inventory) AddGroup(groupname string) (Group, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.Groups[groupname]; ok {
		err := fmt.Errorf("Group %v already exists in inventory", groupname)
		return nil, err
	}
	i.Groups[groupname] = &group{inventory: i}
	return i.Groups[groupname], nil
}

func (i *inventory) GetGroup(groupname string) (Group, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if g, ok := i.Groups[groupname]; ok {
		return g, true
	}
	return nil, false
}

func (i *inventory) Parse() error {
	list := pflag.Bool("list", false, "Lists ansible inventory")
	host := pflag.String("host", "", "Gets variables for individual host")
	pflag.Parse()

	if !*list && *host == "" {
		err := fmt.Errorf("You must specify either --list or --host")
		return err
	}

	if *list && *host != "" {
		err := fmt.Errorf("You must specify only one of either --list or --host")
		return err
	}

	if *list {

		return nil
	}
	return nil
}

func (h *host) AddVariable(key string, val string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.variables[key] = val
}

func (g *group) AddHost(hostname string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, h := range g.inventory.Groups["ungrouped"].hosts {
		if h == hostname {
			g.inventory.Groups["ungrouped"].hosts = removeS(g.inventory.Groups["ungrouped"].hosts, i)
		}
	}
	g.hosts = append(g.hosts, hostname)
}

func removeS(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (g *group) AddVariable(key string, val string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.variables[key] = val
}
