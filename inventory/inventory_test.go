package inventory

import "testing"

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
	// TODO
	// t.Run("Test adding host to group", func(t *testing.T) {
	// 	g, _ := i.GetGroup("testgroup001")
	// 	g.AddHost()
	// })
}
