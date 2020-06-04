package test

import (
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

func EqHash(hash model.Hash) gomock.Matcher {
	return &hashMatcher{expected: hash}
}

type hashMatcher struct {
	expected model.Hash
}

func (h hashMatcher) Matches(x interface{}) bool {
	if actual, ok := x.(model.Hash); !ok {
		return false
	} else {
		return actual.Equals(h.expected)
	}
}

func (h hashMatcher) String() string {
	return "Hash matches: " + h.expected.String()
}

func StringToHash(input string) model.Hash {
	h, _ := model.ParseHash(input)
	return h
}
