package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

func parseParams(params []any, unmarshalTo any) error {
	if len(params) == 0 {
		return errors.New("missing parameters")
	}
	paramsJSON, err := json.Marshal(params[0])
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}
	return json.Unmarshal(paramsJSON, &unmarshalTo)
}

func parseOptionalParams(params []any, unmarshalTo any) error {
	if len(params) == 0 {
		return nil
	}
	paramsJSON, err := json.Marshal(params[0])
	if err != nil {
		return fmt.Errorf("failed to parse optional parameters: %w", err)
	}
	return json.Unmarshal(paramsJSON, &unmarshalTo)
}
