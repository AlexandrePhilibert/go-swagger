// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// stripXOrderFromJSON removes all "x-order" keys from a JSON document
// while preserving the order of remaining keys.
//
// This is used after serialization: x-order ensures properties are
// serialized in the correct order, but should not appear in the output.
func stripXOrderFromJSON(input []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()

	var buf bytes.Buffer
	if err := writeValue(dec, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeValue(dec *json.Decoder, buf *bytes.Buffer) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return writeObject(dec, buf)
		case '[':
			return writeArray(dec, buf)
		default:
			return fmt.Errorf("unexpected delimiter: %v", v)
		}
	case string:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(b)
	case json.Number:
		buf.WriteString(v.String())
	case bool:
		if v {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case nil:
		buf.WriteString("null")
	}

	return nil
}

func writeObject(dec *json.Decoder, buf *bytes.Buffer) error {
	buf.WriteByte('{')
	first := true

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		key, ok := tok.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", tok)
		}

		if key == "x-order" {
			// skip the value
			if err := skipValue(dec); err != nil {
				return err
			}

			continue
		}

		if !first {
			buf.WriteByte(',')
		}
		first = false

		b, err := json.Marshal(key)
		if err != nil {
			return err
		}
		buf.Write(b)
		buf.WriteByte(':')

		if err := writeValue(dec, buf); err != nil {
			return err
		}
	}

	// consume closing '}'
	if _, err := dec.Token(); err != nil {
		return err
	}
	buf.WriteByte('}')

	return nil
}

func writeArray(dec *json.Decoder, buf *bytes.Buffer) error {
	buf.WriteByte('[')
	first := true

	for dec.More() {
		if !first {
			buf.WriteByte(',')
		}
		first = false

		if err := writeValue(dec, buf); err != nil {
			return err
		}
	}

	// consume closing ']'
	if _, err := dec.Token(); err != nil {
		return err
	}
	buf.WriteByte(']')

	return nil
}

func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return nil // scalar value, already consumed
	}

	switch delim {
	case '{':
		for dec.More() {
			// skip key
			if _, err := dec.Token(); err != nil {
				return err
			}
			// skip value
			if err := skipValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token() // consume '}'
	case '[':
		for dec.More() {
			if err := skipValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token() // consume ']'
	}

	return err
}
