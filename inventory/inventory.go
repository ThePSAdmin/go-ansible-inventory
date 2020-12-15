package inventory

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Inventory interface {
	AddHost(hostname string) (Host, error)
	GetHost(hostname string) (Host, bool)
	AddGroup(groupname string) (Group, error)
	GetGroup(groupname string) (Group, bool)
	MarshalJSON() ([]byte, error)
}

type Host interface {
	AddVariable(k string, v string)
	MarshalJSON() ([]byte, error)
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
	variables map[string]string
}

type group struct {
	mu        sync.Mutex
	hosts     []string
	variables map[string]string
	children  []string
	inventory *inventory
}

func (g *group) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Vars     map[string]string `json:"vars,omitempty"`
		Hosts    []string          `json:"hosts,omitempty"`
		Children []string          `json:"children,omitempty"`
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
	j, err := json.Marshal(h.variables)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *inventory) MarshalJSON() ([]byte, error) {
	var ret = map[string]interface{}{
		"_meta": map[string]map[string]*host{
			"hostvars": i.Hosts,
		},
	}

	for n, g := range i.Groups {
		ret[n] = g
	}

	buf, err := json.Marshal(ret)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// NewInventory creates an inventory object that is used for adding groups
// and hosts. Once all inventory has been added, the json object that ansible expects
// can be created by using json.Marshal.
func NewInventory() Inventory {
	return &inventory{
		Hosts:  make(map[string]*host),
		Groups: map[string]*group{"all": {children: []string{"ungrouped"}}, "ungrouped": {}},
	}
}

// AddHost adds a host to the inventory, if the host already exists, it returns
// an error.
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

// GetHost retrieves an a host already added to the inventory, this
// would be needed if any variables needed to be added to the host for example.
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
	allGroup := i.Groups["all"].children
	allGroupChildren := append(allGroup, groupname)
	i.Groups["all"].children = allGroupChildren
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

func (h *host) AddVariable(key string, val string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.variables[key] = val
}

func (g *group) AddHost(hostname string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.inventory.AddHost(hostname)
	g.inventory.Groups["ungrouped"].hosts = removeS(g.inventory.Groups["ungrouped"].hosts, hostname)
	g.hosts = append(g.hosts, hostname)
}

func (g *group) AddVariable(key string, val string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.variables[key] = val
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
