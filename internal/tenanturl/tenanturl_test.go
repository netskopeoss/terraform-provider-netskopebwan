package tenanturl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceTenantDomainInURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		repl string
		want string
	}{
		{
			name: "the leading label is the tenant",
			url:  "https://acme.api.infiot.net",
			repl: "tid-42",
			want: "https://tid-42.api.infiot.net",
		},
		{
			// A dot would cost a wildcard certificate a level, so a deployment that
			// cannot afford one spells the separator out. Either way the tenant is
			// what comes first.
			name: "a -dot- separator is a separator too",
			url:  "https://acme-dot-api.example.net",
			repl: "tid-42",
			want: "https://tid-42-dot-api.example.net",
		},
		{
			name: "the nearer separator wins",
			url:  "https://acme.api-dot-example.net",
			repl: "tid-42",
			want: "https://tid-42.api-dot-example.net",
		},
		{
			name: "the port, path and query are left alone",
			url:  "https://acme.api.example.net:8443/v2/things?first=10",
			repl: "tid-42",
			want: "https://tid-42.api.example.net:8443/v2/things?first=10",
		},
		{
			name: "a host that is all tenant is all replaced",
			url:  "https://acme",
			repl: "tid-42",
			want: "https://tid-42",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := ReplaceTenantDomainInURL(test.url, test.repl)
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
