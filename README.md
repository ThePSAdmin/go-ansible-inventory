# go-ansible-inventory
A go package for creating ansible inventories

### Example

```go
package main

import (
	"encoding/json"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/thepsadmin/go-ansible-inventory/inventory"
)

var (
	list *bool   = flag.Bool("list", false, "Output all hosts info")
	host *string = flag.String("host", "", "Output a specific host")
)

func main() {
	flag.Parse()
	// Create an inventory object
	i := inventory.NewInventory()

	// Add a host
	h, _ := i.AddHost("comp_01")
	// Add a variable to the host
	h.AddVariable("ANSIBLE_HOST", "10.0.0.2")

	// Add a group
	g, _ := i.AddGroup("group_01")

	// Add the host to the group
	g.AddHost("comp_01")

	// Add a variable to the group
	g.AddVariable("environment", "prod")

	if *list {
		j, _ := json.Marshal(i)
		os.Stdout.Write(j)
	}

	if *host != "" {
		if outHost, ok := i.GetHost(*host); ok {
			j, _ := json.Marshal(outHost)
			os.Stdout.Write(j)
		}
	}
}
```