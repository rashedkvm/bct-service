package graphql

import (
	"encoding/json"
)

func (br *UpdateBuildRunRequest) Parse(jsonBytes []byte) error {
	if len(jsonBytes) == 0 {
		return nil
	}
	err := json.Unmarshal(jsonBytes, br)
	if err != nil {
		return err
	}

	return nil
}
