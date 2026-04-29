package util_test

import (
	"testing"

	"github.com/gonimals/goshawk/pkg/util"
)

func TestSyncMap(t *testing.T) {
	sm := util.NewSyncMap[string, int]()

	sm.Set("test", 1)
	if val := sm.Get("test"); val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	if sm.CompareAndSwap("test", 1, 2) {
		if val := sm.Get("test"); val != 2 {
			t.Errorf("Expected 2, got %v", val)
		}
	} else {
		t.Errorf("CAS failed")
	}

	if sm.CompareAndSwap("test", 1, 3) {
		t.Errorf("CAS succeeded when it should have failed")
	}
}

func TestSyncMapNonComparable(t *testing.T) {
	sm := util.NewSyncMap[string, []int]()

	sm.Set("test", []int{1})
	if sm.CompareAndSwap("test", []int{1}, []int{2}) {
		t.Errorf("CAS succeeded for non-comparable values")
	}
}
