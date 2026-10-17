// Package jsonx holds JSON scalar types that tolerate the loosely typed
// mailbox.org API, whose documentation often leaves field types unstated.
package jsonx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func isNull(b []byte) bool { return bytes.Equal(bytes.TrimSpace(b), []byte("null")) }

// Text is a JSON scalar rendered as text. Strings, numbers and booleans are
// accepted; null becomes the empty string; objects and arrays keep their raw JSON.
type Text string

// UnmarshalJSON implements json.Unmarshaler.
func (t *Text) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if isNull(b) {
		*t = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = Text(s)
		return nil
	}
	*t = Text(b)
	return nil
}

// Flag is a boolean that may arrive as true/false, 0/1 or "0"/"1"/"true"/"false".
// null is false.
type Flag bool

// UnmarshalJSON implements json.Unmarshaler.
func (f *Flag) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if isNull(b) {
		*f = false
		return nil
	}
	s := strings.ToLower(strings.Trim(string(b), `"`))
	switch s {
	case "true", "1", "yes", "on":
		*f = true
	case "false", "0", "no", "off", "":
		*f = false
	default:
		return fmt.Errorf("jsonx: cannot read %s as a boolean", b)
	}
	return nil
}

// Count is an integer that may arrive as a number or a numeric string. null is 0.
type Count int64

// UnmarshalJSON implements json.Unmarshaler.
func (c *Count) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if isNull(b) {
		*c = 0
		return nil
	}
	n, err := strconv.ParseInt(strings.Trim(string(b), `"`), 10, 64)
	if err != nil {
		return fmt.Errorf("jsonx: cannot read %s as an integer", b)
	}
	*c = Count(n)
	return nil
}

// TextList is a list of scalars. null and "" become the empty list.
type TextList []Text

// UnmarshalJSON implements json.Unmarshaler.
func (l *TextList) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if isNull(b) || bytes.Equal(b, []byte(`""`)) {
		*l = TextList{}
		return nil
	}
	var items []Text
	if err := json.Unmarshal(b, &items); err != nil {
		return fmt.Errorf("jsonx: cannot read %s as a list", b)
	}
	*l = items
	return nil
}

// Strings returns the list as plain strings; it is never nil.
func (l TextList) Strings() []string {
	out := make([]string, len(l))
	for i, v := range l {
		out[i] = string(v)
	}
	return out
}
