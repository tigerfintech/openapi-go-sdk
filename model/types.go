package model

import (
	"encoding/json"
	"fmt"
)

// FlexString unmarshals a JSON value that may be either a string or a number
// (e.g. an integer). The value is stored as its string representation.
//
// This is needed for fields like AddonEntitlement.UserLevel where the server
// inconsistently returns a numeric value instead of a string.
type FlexString string

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexString) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	// Fall back to number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexString(n.String())
		return nil
	}
	return fmt.Errorf("FlexString: cannot unmarshal %s", data)
}

// String returns the string value.
func (f FlexString) String() string {
	return string(f)
}
