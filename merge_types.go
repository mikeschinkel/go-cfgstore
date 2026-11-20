package cfgstore

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// AppendSlice is a slice field whose later layers append their elements to the
// earlier layers' elements, where a plain slice field would be replaced. It
// marshals as a plain JSON array. A layer that sets the key to null empties it.
type AppendSlice[T any] []T

// UnmarshalJSONFrom appends the array's elements to s.
func (s *AppendSlice[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) (err error) {
	var elems []T

	if dec.PeekKind() == 'n' {
		_, err = dec.ReadToken()
		*s = nil
		goto end
	}
	err = jsonv2.UnmarshalDecode(dec, &elems)
	if err != nil {
		goto end
	}
	*s = append(*s, elems...)
end:
	return err
}

// ReplaceMap is a map field whose later layers replace the earlier layers' map
// as a whole, where a plain map field would merge their entries. It marshals as
// a plain JSON object.
type ReplaceMap[K comparable, V any] map[K]V

// UnmarshalJSONFrom replaces m with the object's entries.
func (m *ReplaceMap[K, V]) UnmarshalJSONFrom(dec *jsontext.Decoder) (err error) {
	var entries map[K]V

	err = jsonv2.UnmarshalDecode(dec, &entries)
	if err != nil {
		goto end
	}
	*m = entries
end:
	return err
}
