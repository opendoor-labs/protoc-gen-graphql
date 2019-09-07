package mapper

import (
	"unicode"
	"unicode/utf8"
)

func lowerUnderscoreToLowerCamelTransformer(name string) string {
	return toLowerCamel(parseLowerUnderscore(name))
}

func upperCamelToLowerCamelTransformer(name string) string {
	return toLowerCamel(parseUpperCamel(name))
}

func preserveTransformer(name string) string {
	return name
}

func lowerCaseFirstRune(v string) string {
	r, size := utf8.DecodeRuneInString(v)
	return string(unicode.ToLower(r)) + v[size:]
}

func upperCaseFirstRune(v string) string {
	r, size := utf8.DecodeRuneInString(v)
	return string(unicode.ToUpper(r)) + v[size:]
}

func toLowerCamel(words []string) string {
	// Replicates logic used by the JS compiler, see function ToLowerCamel in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var result string
	for i, word := range words {
		if i == 0 && (word[0] >= 'A' && word[0] <= 'Z') {
			word = lowerCaseFirstRune(word)
		} else if i != 0 && (word[0] >= 'a' && word[0] <= 'z') {
			word = upperCaseFirstRune(word)
		}
		result += word
	}
	return result
}

func toUpperCamel(words []string) string {
	// Replicates logic used by the JS compiler, see function ToUpperCamel in:
	// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/compiler/js/js_generator.cc
	var result string
	for _, word := range words {
		if word[0] >= 'a' && word[0] <= 'z' {
			word = upperCaseFirstRune(word)
		}
		result += word
	}
	return result
}

func parseLowerUnderscore(input string) []string {
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

func parseUpperCamel(input string) []string {
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
