package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/pflag"
)

type Inventory interface {
	AddHost(hostname string) (Host, error)
	GetHost(hostname string) (Host, bool)
	AddGroup(groupname string) (Group, error)
	GetGroup(groupname string) (Group, bool)
	Parse() error
	WriteOutput(h string, l bool) error
}

type Host interface {
	AddVariable(k string, v string)
}

type Group interface {
	AddHost(hostname string)
	AddVariable(k string, v string)
	Hosts() []string
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
	mu        sync.Mutex        `json:"-"`
	hosts     []string          `json:"hosts"`
	variables map[string]string `json:"vars"`
	children  []string          `json:"children"`
	inventory *inventory        `json:"-"`
}

func (g *group) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Vars      map[string]string `json:"vars,omitempty""`
		Hosts     []string          `json:"hosts,omitempty"`
		Children  []string          `json:"children,omitempty"`
	}{
		g.variables,
		g.hosts,
		g.children,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (h *host) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Vars map[string]string `json:"vars,omitempty""`
	}{
		h.variables,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
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
	i.Hosts[hostname] = &host{variables: make(map[string]string)}
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
	i.Groups[groupname] = &group{inventory: i, variables: make(map[string]string)}
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
	h := pflag.String("h", "", "Gets variables for individual h")
	pflag.Parse()

	err := i.WriteOutput(*h, *list)
	if err != nil {
		return err
	}
	return nil
}

func (i *inventory) WriteOutput(h string, list bool) error {
	if !list && h == "" {
		return fmt.Errorf("You must specify either --list or --h")
	}

	if list && h != "" {
		return fmt.Errorf("You must specify only one of either --list or --h")
	}

	if list {
		ret := make(map[string]json.RawMessage)

		hostvars := map[string]map[string]*host{
			"hostvars": i.Hosts,
		}
		res, err := json.Marshal(hostvars)
		if err != nil {
			panic(err)
		}
		ret["_meta"] = res

		for n, g := range i.Groups {
			ret[n], _ = json.Marshal(g)
		}
		buf, err := json.Marshal(ret)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(buf)
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
	g.inventory.Groups["ungrouped"].hosts = removeS(g.inventory.Groups["ungrouped"].hosts, hostname)
	g.hosts = append(g.hosts, hostname)
}

func removeS(s []string, item string) []string {
	for i, v := range s {
		if v == item {
			if len(s) == 1 {
				return make([]string, 0)
			}
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
		}
	}
	return s
}

func (g *group) Hosts() []string {
	return g.hosts
}

func (g *group) AddVariable(key string, val string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.variables[key] = val
}
