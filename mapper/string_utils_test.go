package mapper

import (
	"testing"
)

func TestLowerCaseFirstRune(t *testing.T) {
	var testCases = []struct{ in, out string }{
		{"Abc", "abc"},
		{"ABC", "aBC"},
		{"abc", "abc"},
		{"Abc Def", "abc Def"},
		{"012", "012"},
		{"012Abc", "012Abc"},
	}
	for _, testCase := range testCases {
		s := lowerCaseFirstRune(testCase.in)
		if s != testCase.out {
			t.Errorf("got %s; want %s", s, testCase.out)
		}
	}
}

func TestUpperCaseFirstRune(t *testing.T) {
	var testCases = []struct{ in, out string }{
		{"abc", "Abc"},
		{"aBC", "ABC"},
		{"ABC", "ABC"},
		{"abc def", "Abc def"},
		{"012", "012"},
		{"012abc", "012abc"},
	}
	for _, testCase := range testCases {
		s := upperCaseFirstRune(testCase.in)
		if s != testCase.out {
			t.Errorf("got %s; want %s", s, testCase.out)
		}
	}
}
