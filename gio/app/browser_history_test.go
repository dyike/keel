package app

import (
	"github.com/dyike/keel/gio/unit"
	"testing"
)

func TestCustomBrowserHistoryOption(t *testing.T) {
	var cfg Config
	if cfg.CustomBrowserHistory {
		t.Fatal("default must retain Gio Back handling")
	}
	CustomBrowserHistory(true)(unit.Metric{}, &cfg)
	if !cfg.CustomBrowserHistory {
		t.Fatal("host history ownership not enabled")
	}
	CustomBrowserHistory(false)(unit.Metric{}, &cfg)
	if cfg.CustomBrowserHistory {
		t.Fatal("history ownership not reset")
	}
}
