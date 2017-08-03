package prpl

import (
	"testing"

	"github.com/ua-parser/uap-go/uaparser"
)

var capabilities = uaparser.NewFromSaved()

func assertBrowserCapabilities(userAgent string, expect capability) {

}

func TestCapabilities(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Error(err.Error())
		return
	}

	tests := []struct {
		userAgent    string
		capabilities capability
	}{
		// unknown browser has no capabilities
		{"unknown browser", 0},

		// chrome has all the capabilities
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.96 Safari/537.36", es2015 + push + serviceworker},

		// edge es2015 support is predicated on minor browser version
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36 Edge/15.14986", push},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36 Edge/15.15063", es2015 + push},

		// safari push capability is predicated on macOS version
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.1 Safari/603.1.30", es2015},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.1 Safari/603.1.30", es2015 + push},
	}

	for _, test := range tests {
		if capabilities := p.browserCapabilities(test.userAgent); capabilities != test.capabilities {
			t.Errorf("expected %s to have %s: got %s", test.userAgent, test.capabilities, capabilities)
		}
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version version
		values  []int
	}{
		{"37", []int{37}},
		{"10.987.00.1", []int{10, 987, 0, 1}},
		{"4..foo.7", []int{4, -1, -1, 7}},
	}

	for _, test := range tests {
		values := parseVersion(test.version)
		for j, value := range values {
			if value != test.values[j] {
				t.Fail()
			}
		}
	}
}

type version string

func (v version) ToVersionString() string {
	return string(v)
}

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		result  bool
		atLeast []int
		version []int
	}{
		{true, ints(3, 2, 1), ints(3, 2, 1)},
		{true, ints(3, 2, 1), ints(3, 2, 1, 4)},
		{true, ints(3, 2, 1), ints(4, 1, 0)},
		{true, ints(3, 2, 0), ints(3, 2)},

		{false, ints(3, 2, 1), ints(2, 2, 1)},
		{false, ints(3, 2, 1), ints(3, 1, 1)},
		{false, ints(3, 2, 1), ints(3, 1, 0)},
		{false, ints(3, 2, 1), ints(3, 2)},
		{false, ints(3, 2, 1), ints()},
	}

	for _, test := range tests {
		if versionAtLeast(test.version, test.atLeast...) != test.result {
			t.Fail()
		}
	}
}

func ints(ints ...int) []int {
	return ints
}
