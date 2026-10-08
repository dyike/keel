//go:build !darwin

package main

import "context"

func prepareRunBundle(_ context.Context, _ *Config, binary, _ string) (string, func(), error) {
	return binary, func() {}, nil
}
