package models

import (
	"testing"
)

func TestPropertySet(t *testing.T) {
	flag := theChangableProperties.Match("enabled")
	if !flag {
		t.Errorf("expect true flag but got false")
	}
}
