package inventory

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestHosts(t *testing.T) {
	i := NewInventory()
	t.Run("Test adding hosts", func(t *testing.T) {
		_, err := i.AddHost("test01")
		if err != nil {
			t.Errorf("Error adding host to inventory")
		}
	})
	t.Run("Add duplicate host should fail", func(t *testing.T) {
		_, err := i.AddHost("test01")
		if err == nil {
			t.Errorf("No error created adding duplicate host to inventory")
		}
	})
	t.Run("Test getting host that exists", func(t *testing.T) {
		r, ok := i.GetHost("test01")
		if r == nil && !ok {
			t.Errorf("Expected to get back host and ok")
		}
	})
	t.Run("Test getting host that doesn't exists", func(t *testing.T) {
		r, ok := i.GetHost("test02")
		if r != nil && ok {
			t.Errorf("Expected to get back nil and not ok")
		}
	})

}

func TestGroups(t *testing.T) {
	i := NewInventory()
	t.Run("Test adding group", func(t *testing.T) {
		g, err := i.AddGroup("testgroup001")
		if err != nil && g == nil {
			t.Errorf("Could not add group")
		}
	})
	t.Run("Test adding duplicate group", func(t *testing.T) {
		g, err := i.AddGroup("testgroup001")
		if err == nil && g != nil {
			t.Errorf("Expected to get back nil and not ok")
		}
	})
	t.Run("Get group that exists", func(t *testing.T) {
		g, ok := i.GetGroup("testgroup001")
		if g == nil && ok != false {
			t.Errorf("Expected to get back group and ok")
		}
	})
	t.Run("Get group that does not exist", func(t *testing.T) {
		g, ok := i.GetGroup("testgroup002")
		if g != nil && ok == false {
			t.Errorf("Expected to get back nil and not ok")
		}
	})
	t.Run("Add a host that hasn't been added via inventory add host yet", func(t *testing.T) {
		g, _ := i.GetGroup("testgroup001")
		g.AddHost("comp01")
		h, _ := i.GetHost("comp01")
		if h == nil {
			t.Errorf("Could not add a new host via group")
		}
	})

}
func TestInventory(t *testing.T) {
	i := NewInventory()
	h, _ := i.AddHost("comp01")
	h1, _ := i.AddHost("comp02")
	h.AddVariable("foo", "bar")
	h1.AddVariable("baz", "buzz")
	g, _ := i.AddGroup("group01")
	g.AddVariable("gvar", "gbaz")
	g.AddHost("comp01")

	expected := `
	{
		"_meta": {
			"hostvars": {
				"comp01": {
					"foo": "bar"
				},
				"comp02": {
					"baz": "buzz"
				}
			}
		},
		"all": {
			"hosts": [
				"comp01",
				"comp02"
			],
			"children": [
				"ungrouped",
				"group01"
			]
		},
		"group01": {
			"vars": {
				"gvar": "gbaz"
			},
			"hosts": [
				"comp01"
			]
		},
		"ungrouped": {
			"hosts": [
				"comp02"
			]
		}
	}
`
	actual, _ := json.Marshal(i)

	var expectedObj interface{}
	var actualObj interface{}
	json.Unmarshal([]byte(expected), &expectedObj)
	json.Unmarshal(actual, &actualObj)

	if !reflect.DeepEqual(expectedObj, actualObj) {
		fmt.Printf("Actual: %v", string(actual))
		t.Errorf("Expected: %v, not equal to actual %v", expectedObj, actualObj)
	}
}
