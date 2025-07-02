package main

import (
	"encoding/json"
	"fmt"
)

func parseParams(params []any, unmarshalTo any) error {
	if len(params) > 0 {
		paramsJSON, err := json.Marshal(params[0])
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}
		err = json.Unmarshal(paramsJSON, &unmarshalTo)
		if err != nil {
			return err
		}
	}
	return validate.Struct(unmarshalTo)
}
