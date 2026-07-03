package model

import (
	"encoding/json"
	"fmt"
	"strconv"
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

// FlexInt64 unmarshals a JSON value that may be either a number or a
// quoted-string representation of an integer.
//
// This handles fields like FundDetails.ID where the server inconsistently
// returns a string instead of a JSON number.
type FlexInt64 int64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexInt64) UnmarshalJSON(data []byte) error {
	// Try number first
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		v, err := n.Int64()
		if err != nil {
			return fmt.Errorf("FlexInt64: cannot parse %s as int64: %w", data, err)
		}
		*f = FlexInt64(v)
		return nil
	}
	// Fall back to quoted string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("FlexInt64: cannot parse %q as int64: %w", s, err)
		}
		*f = FlexInt64(v)
		return nil
	}
	return fmt.Errorf("FlexInt64: cannot unmarshal %s", data)
}

// Int64 returns the int64 value.
func (f FlexInt64) Int64() int64 {
	return int64(f)
}
