package decoder

import (
	"errors"
	"reflect"
)

// NewDecoder returns a new Decoder.
func NewDecoder() *Decoder {
	return &Decoder{cache: newCache(), ignoreUnknownKeys: true}
}

// Decoder decodes values from a map[string][]string to a struct.
type Decoder struct {
	cache             *cache
	ignoreUnknownKeys bool
}

// SetAliasTag changes the tag used to locate custom field aliases.
// The default tag is "schema".
func (d *Decoder) SetAliasTag(tag string) {
	d.cache.tag = tag
}

// IgnoreUnknownKeys controls the behaviour when the decoder encounters unknown
// keys in the map.
// If i is true and an unknown field is encountered, it is ignored. This is
// similar to how unknown keys are handled by encoding/json.
// If i is false then Decode will return an error. Note that any valid keys
// will still be decoded in to the target struct.
//
// To preserve backwards compatibility, the default value is false.
func (d *Decoder) IgnoreUnknownKeys(i bool) {
	d.ignoreUnknownKeys = i
}

// Decode decodes a map[string]interface{} to a struct.
//
// The first parameter must be a pointer to a struct.
//
// The second parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
//
// See the package documentation for a full explanation of the mechanics.
func (d *Decoder) Decode(dst interface{}, src map[string]interface{}) error {
	vals := reflect.ValueOf(dst)
	if vals.Kind() != reflect.Ptr || vals.Elem().Kind() != reflect.Struct {
		return errors.New("decoder: interface must be a pointer to struct")
	}

	t := vals.Elem().Type()
	cachestruct := d.cache.Get(t)
	if cachestruct == nil {
		return errors.New("decoder: invalid struct")
	}

	return d.cache.decode(cachestruct, vals.Elem(), src)
}
