package api

import (
	"errors"
	"testing"
)

func TestGuildAllowed(t *testing.T) {
	noFetch := func() ([]string, error) {
		t.Fatal("Discord member fetch should not run")
		return nil, nil
	}

	cases := []struct {
		name        string
		hasConfig   bool
		whitelisted bool
		roleIds     []string
		cached      bool
		fetch       func() ([]string, error)
		want        bool
	}{
		{name: "no server config", hasConfig: false, whitelisted: true, want: false},
		{name: "member whitelisted", hasConfig: true, whitelisted: true, want: true},
		{name: "no whitelist entries", hasConfig: true, roleIds: nil, want: false},
		{name: "role matched from cache", hasConfig: true, roleIds: []string{"r1"}, cached: true, fetch: noFetch, want: true},
		{name: "role matched from API fallback", hasConfig: true, roleIds: []string{"r1"}, fetch: func() ([]string, error) { return []string{"r1", "r2"}, nil }, want: true},
		{name: "role not matched", hasConfig: true, roleIds: []string{"r1"}, fetch: func() ([]string, error) { return []string{"r2"}, nil }, want: false},
		{name: "API fallback error", hasConfig: true, roleIds: []string{"r1"}, fetch: func() ([]string, error) { return nil, errors.New("boom") }, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fetch == nil {
				tc.fetch = noFetch
			}
			got := guildAllowed(tc.hasConfig, tc.whitelisted, tc.roleIds, tc.cached, tc.fetch)
			if got != tc.want {
				t.Fatalf("guildAllowed() = %v, want %v", got, tc.want)
			}
		})
	}
}
