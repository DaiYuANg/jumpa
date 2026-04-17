package main

import (
	"fmt"
	"slices"
	"strings"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/DaiYuANg/jumpa/internal/sshclient"
)

func validateRequestListOptions(options *cli.ListOptions) error {
	if options == nil {
		return nil
	}

	if options.Page < 1 {
		return fmt.Errorf("invalid --page %d: must be >= 1", options.Page)
	}
	if options.PageSize < 1 || options.PageSize > 200 {
		return fmt.Errorf("invalid --page-size %d: must be between 1 and 200", options.PageSize)
	}

	status := strings.ToLower(strings.TrimSpace(options.Status))
	if status == "" {
		options.Status = ""
		return nil
	}
	if !slices.Contains(accessRequestStatusValues, status) {
		return fmt.Errorf("invalid --status %q: allowed values are %s", options.Status, strings.Join(accessRequestStatusValues, ", "))
	}

	options.Status = status
	return nil
}

func parseLocalForwards(specs []string) ([]sshclient.LocalForward, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	forwards := make([]sshclient.LocalForward, 0, len(specs))
	for _, spec := range specs {
		forward, err := sshclient.ParseLocalForward(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid --local-forward %q: %w", spec, err)
		}
		forwards = append(forwards, forward)
	}
	return forwards, nil
}

func parseRemoteForwards(specs []string) ([]sshclient.RemoteForward, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	forwards := make([]sshclient.RemoteForward, 0, len(specs))
	for _, spec := range specs {
		forward, err := sshclient.ParseRemoteForward(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid --remote-forward %q: %w", spec, err)
		}
		forwards = append(forwards, forward)
	}
	return forwards, nil
}

func parseDynamicForwards(specs []string) ([]sshclient.DynamicForward, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	forwards := make([]sshclient.DynamicForward, 0, len(specs))
	for _, spec := range specs {
		forward, err := sshclient.ParseDynamicForward(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid --dynamic-forward %q: %w", spec, err)
		}
		forwards = append(forwards, forward)
	}
	return forwards, nil
}
