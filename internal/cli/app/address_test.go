package app

import "testing"

func TestParseGatewayAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		host    string
		port    string
		wantErr bool
	}{
		{name: "default empty", input: "", host: DefaultGatewayHost, port: DefaultGatewayPort},
		{name: "hostname only", input: "gateway.example.com", host: "gateway.example.com", port: DefaultSSHPort},
		{name: "hostname and port", input: "gateway.example.com:2200", host: "gateway.example.com", port: "2200"},
		{name: "implicit host", input: ":2200", host: DefaultGatewayHost, port: "2200"},
		{name: "ipv6 only", input: "::1", host: "[::1]", port: DefaultSSHPort},
		{name: "ipv6 and port", input: "[::1]:2200", host: "[::1]", port: "2200"},
		{name: "invalid port", input: "gateway.example.com:not-a-port", wantErr: true},
		{name: "invalid ipv6 port", input: "[::1]:99999", wantErr: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host, port, err := ParseGatewayAddress(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if host != tc.host || port != tc.port {
				t.Fatalf("ParseGatewayAddress(%q) = (%q, %q), want (%q, %q)", tc.input, host, port, tc.host, tc.port)
			}
		})
	}
}

func TestNormalizeGatewayAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{input: "", want: "127.0.0.1:2222"},
		{input: "gateway.example.com", want: "gateway.example.com:22"},
		{input: "::1", want: "[::1]:22"},
		{input: "[::1]:2200", want: "[::1]:2200"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeGatewayAddress(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeGatewayAddress(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
