package mapper

import (
	"fmt"
	"strings"
)

const (
	InputModeNone    = "none"
	InputModeService = "service"
	InputModeAll     = "all"
)

const (
	FieldNameDefault  = "lower_camel_case"
	FieldNamePreserve = "preserve"
)

type Parameters struct {
	TimestampTypeName string
	DurationTypeName  string
	WrappersAsNull    bool
	InputMode         string
	String64Bit       bool
	RootTypePrefix    *string
	FieldName         string
	TrimPrefix        string
}

func NewParameters(parameter string) (*Parameters, error) {
	params := &Parameters{}

	parts := strings.Split(parameter, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}

		keyValue := strings.SplitN(part, "=", 2)
		key := keyValue[0]
		var value string
		if len(keyValue) == 2 {
			value = keyValue[1]
		}

		switch key {
		case "timestamp":
			if value == "" {
				return nil, fmt.Errorf("missing type for timestamp")
			}
			params.TimestampTypeName = value
		case "duration":
			if value == "" {
				return nil, fmt.Errorf("missing type for duration")
			}
			params.DurationTypeName = value
		case "null_wrappers":
			params.WrappersAsNull = true
		case "input_mode":
			params.InputMode = value
		case "string_64bit":
			params.String64Bit = true
		case "root_type_prefix":
			params.RootTypePrefix = &value
		case "field_name":
			if value != FieldNamePreserve {
				value = FieldNameDefault
			}
			params.FieldName = value
		case "trim_prefix":
			params.TrimPrefix = value
		}
	}

	if params.InputMode == "" {
		params.InputMode = InputModeService
	}

	return params, nil
}
