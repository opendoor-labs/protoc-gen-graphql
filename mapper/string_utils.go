package mapper

import (
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
