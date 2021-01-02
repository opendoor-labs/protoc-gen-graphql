package mapper

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// LowerUnderscoreToLowerCamelTransformer transforms a string from lower_underscore
// format to lowerCamelCase format.
func LowerUnderscoreToLowerCamelTransformer(name string) string {
	return ToLowerCamel(ParseLowerUnderscore(name))
}

// UpperCamelToLowerCamelTransformer transforms a string from UpperCamelCase format
// to lowerCamelCase format.
func UpperCamelToLowerCamelTransformer(name string) string {
	return ToLowerCamel(ParseUpperCamel(name))
}

// PreserveTransformer is a no-op.
func PreserveTransformer(name string) string {
	return name
}

// LowerCaseFirstRune returns the input with the first rune lower cased.
func LowerCaseFirstRune(v string) string {
	r, size := utf8.DecodeRuneInString(v)
	return string(unicode.ToLower(r)) + v[size:]
}

// UpperCaseFirstRune returns the input with the first rune upper cased.
func UpperCaseFirstRune(v string) string {
	r, size := utf8.DecodeRuneInString(v)
	return string(unicode.ToUpper(r)) + v[size:]
}

func ToLowerCamel(words []string) string {
	// Replicates logic used by the JS compiler, see function ToLowerCamel in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var result string
	for i, word := range words {
		if i == 0 && (word[0] >= 'A' && word[0] <= 'Z') {
			word = LowerCaseFirstRune(word)
		} else if i != 0 && (word[0] >= 'a' && word[0] <= 'z') {
			word = UpperCaseFirstRune(word)
		}
		result += word
	}
	return result
}

func ToUpperCamel(words []string) string {
	// Replicates logic used by the JS compiler, see function ToUpperCamel in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var result string
	for _, word := range words {
		if word[0] >= 'a' && word[0] <= 'z' {
			word = UpperCaseFirstRune(word)
		}
		result += word
	}
	return result
}

func ParseLowerUnderscore(input string) []string {
	// Replicates logic used by the JS compiler, see function ParseLowerUnderscore in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var words []string
	var running string
	for _, r := range input {
		if r == '_' {
			if running != "" {
				words = append(words, running)
				running = ""
			}
		} else {
			running += string(unicode.ToLower(r))
		}
	}
	if running != "" {
		words = append(words, running)
	}
	return words
}

func ParseUpperCamel(input string) []string {
	// Replicates logic used by the JS compiler, see function ParseUpperCamel in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var words []string
	var running string
	for _, r := range input {
		if r >= 'A' && r <= 'Z' && running != "" {
			words = append(words, running)
			running = ""
		}
		running += string(unicode.ToLower(r))
	}
	if running != "" {
		words = append(words, running)
	}
	return words
}

// CamelCaseSlice is like CamelCase, but the argument is a slice of strings to
// be joined with "_".
func CamelCaseSlice(elem []string) string { return CamelCase(strings.Join(elem, "_")) }

// CamelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercases names, it's extremely unlikely to have two fields
// with different capitalizations.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
