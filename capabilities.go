package prpl

import (
	"strconv"
	"strings"

	"github.com/ua-parser/uap-go/uaparser"
)

type (
	userAgentPredicate func(client *uaparser.Client) bool

	userAgentCapabilities struct {
		es2015        userAgentPredicate
		push          userAgentPredicate
		serviceworker userAgentPredicate
	}
)

var browserPredicates = map[string]userAgentCapabilities{
	"Chrome": {
		es2015:        since(49),
		push:          since(41),
		serviceworker: since(45),
	},
	"Chromium": {
		es2015:        since(49),
		push:          since(41),
		serviceworker: since(45),
	},
	"OPR": {
		es2015:        since(36),
		push:          since(28),
		serviceworker: since(32),
	},
	"Vivaldi": {
		es2015:        since(1),
		push:          since(1),
		serviceworker: since(1),
	},
	"Mobile Safari": {
		es2015:        since(10),
		push:          since(9, 2),
		serviceworker: notyet,
	},
	"Safari": {
		es2015: since(10),
		push: func(client *uaparser.Client) bool {
			return versionAtLeast(parseVersion(client.UserAgent), 9) &&
				// HTTP/2 on desktop Safari requires macOS 10.11 according to
				// caniuse.com.
				versionAtLeast(parseVersion(client.Os), 10, 11)
		},
		// https://webkit.org/status/#specification-service-workers
		serviceworker: notyet,
	},
	"Edge": {
		// Edge versions before 15.15063 may contain a JIT bug affecting ES6
		// constructors (https://github.com/Microsoft/ChakraCore/issues/1496).
		es2015: since(15, 15063),
		push:   since(12),
		// https://developer.microsoft.com/en-us/microsoft-edge/platform/status/serviceworker/
		serviceworker: notyet,
	},
	"Firefox": {
		es2015:        since(51),
		push:          since(36),
		serviceworker: since(44),
	},
}

type capability int

const (
	es2015 capability = 1 << iota
	push
	serviceworker
)

func newCapabilities(browserCapabilities []string) capability {
	var capability capability
	for _, c := range browserCapabilities {
		switch c {
		case "es2015":
			capability &= es2015
		case "push":
			capability &= push
		case "serviceworker":
			capability &= serviceworker
		}
	}
	return capability
}

func (c capability) size() int {
	return 0
}

func (c capability) String() string {
	val := []string{}
	if c&es2015 == es2015 {
		val = append(val, "es2015")
	}
	if c&push == push {
		val = append(val, "push")
	}
	if c&serviceworker == serviceworker {
		val = append(val, "serviceworker")
	}
	return strings.Join(val, ", ")
}

func (p *prpl) browserCapabilities(userAgentString string) capability {
	client := p.parser.Parse(userAgentString)

	predicate, ok := browserPredicates[client.UserAgent.Family]
	if !ok {
		return 0
	}

	var capabilities capability
	if predicate.es2015(client) {
		capabilities += es2015
	}
	if predicate.push(client) {
		capabilities += push
	}
	if predicate.serviceworker(client) {
		capabilities += serviceworker
	}

	return capabilities
}

var notyet = func(client *uaparser.Client) bool {
	return false
}

func since(atLeast ...int) userAgentPredicate {
	return func(client *uaparser.Client) bool {
		version := parseVersion(client.UserAgent)
		return versionAtLeast(version, atLeast...)
	}
}

func versionAtLeast(version []int, atLeast ...int) bool {
	for i, r := range atLeast {
		var v int
		if len(version) > i {
			v = version[i]
		}
		if v > r {
			return true
		}
		if v < r {
			return false
		}
	}
	return true
}

type versioned interface {
	ToVersionString() string
}

func parseVersion(version versioned) []int {
	parts := strings.Split(version.ToVersionString(), ".")
	values := make([]int, len(parts))
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			value = -1
		}
		values[i] = value
	}
	return values
}
