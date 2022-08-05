// Package jsonutil provides various utility functions for working
// with JSON data.

package jsonutil

import (
	"encoding/json"
	"fmt"
	"io"
)

var ErrNotFound = fmt.Errorf("not found")

// SliceIterator reads a JSON encoded object from r and iterates over all array items found at the
// provided path. For each item found, getDst is called to construct the destination for unmarshalling.
func SliceIterator(r io.Reader, getDst func() interface{}, path ...string) func() (bool, error) {
	dec := json.NewDecoder(r)
	if err := trackPath(dec, path); err != nil {
		return errIterator(err)
	}
	if err := expectDelim(dec, '['); err != nil {
		return errIterator(err)
	}
	return func() (bool, error) {
		if !dec.More() {
			return false, expectDelim(dec, ']')
		}
		dst := getDst()
		err := dec.Decode(dst)
		if err == io.EOF {
			return false, nil
		}
		return true, err
	}
}

func trackPath(dec *json.Decoder, path []string) error {
	if len(path) == 0 {
		return nil
	}
	var foundCount int
	for i, field := range path {
		// Read the opening bracket of the object we are inspecting
		if err := expectDelim(dec, '{'); err != nil {
			break
		}

		// Search for the field within the object. The object may contain other fields than the one we
		// are searching for. Iterate over all fields until we find the one we are looking for.
		for {
			t, err := dec.Token()
			if err != nil {
				return err
			}

			// If we've reached the end of the object, we couldn't find the field
			if t == json.Delim('}') {
				break
			}

			got, ok := t.(string)
			if !ok {
				return fmt.Errorf("unexpected token. Expected field name, got: %v (%T)", t, t)
			}

			// We found the field.
			if got == field {
				foundCount = i + 1
				break
			}

			// We've read the name of another field. Read the next token to inspect what kind of value it contains.
			t, err = dec.Token()
			if err != nil {
				return err
			}

			// If it's a delimiter, the value is an array or an object. Read through the object before we continue
			// the search for the field in the next iteration. If the value was not an array or object, we've already
			// read it to completion and don't need to do anything more.
			if _, ok := t.(json.Delim); ok {
				if err := readArrayOrObject(dec); err != nil {
					return err
				}
			}
		}
	}
	if foundCount == len(path) {
		return nil
	}
	return ErrNotFound
}

func readArrayOrObject(dec *json.Decoder) error {
	tokenCount := 1
	for tokenCount > 0 {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		d, ok := t.(json.Delim)
		if ok {
			switch d {
			case '{', '[':
				tokenCount++
			case '}', ']':
				tokenCount--
			}
		}
	}
	return nil
}

func expectDelim(dec *json.Decoder, expected json.Delim) error {
	got, err := dec.Token()
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("jsonutil: unexpected token: expected '%v', got '%v'", expected, got)
	}
	return nil
}

func errIterator(err error) func() (bool, error) {
	return func() (bool, error) {
		return false, err
	}
}
