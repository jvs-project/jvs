package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// CanonicalMarshal produces deterministic JSON:
// - keys sorted lexicographically
// - no whitespace
// - UTF-8 encoding
// - null serialized as null
func CanonicalMarshal(v any) ([]byte, error) {
	// First marshal to standard JSON to normalize the value
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonical marshal: %w", err)
	}

	// Unmarshal into generic structure
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("canonical unmarshal: %w", err)
	}

	// Re-serialize with sorted keys
	var buf bytes.Buffer
	if err := writeCanonical(&buf, generic); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCanonical(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			if err := writeCanonical(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')

	case []any:
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')

	default:
		// Primitives: string, float64, bool, nil
		raw, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(raw)
	}
	return nil
}
