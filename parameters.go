package main

import (
	"fmt"
	"strings"
)

type Parameters struct {
	TimestampTypeName string
	DurationTypeName  string
	WrappersAsNull    bool
	ServiceTypesOnly  bool
}

func NewParameters(parameter string) (*Parameters, error) {
	params := &Parameters{}

	if parameter == "" {
		return params, nil
	}

	parts := strings.Split(parameter, ",")
	for _, part := range parts {
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
		case "service_only":
			params.ServiceTypesOnly = true
		default:
			return nil, fmt.Errorf("unknown parameter: %s", key)
		}
	}

	return params, nil
}
