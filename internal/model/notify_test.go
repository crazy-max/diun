package model

import (
	"slices"
	"testing"
)

func TestNotifyOnUnchangeIsOptIn(t *testing.T) {
	status := NotifyOnUnchange
	if !status.Valid() {
		t.Fatal("unchange must be accepted")
	}
	if slices.Contains(NotifyOnDefaults, status) {
		t.Fatal("unchange must not be enabled by default")
	}
	invalid := NotifyOn("error")
	if invalid.Valid() {
		t.Fatal("error must not be accepted")
	}
}
