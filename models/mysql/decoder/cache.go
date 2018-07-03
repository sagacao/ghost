package decoder

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// newCache returns a new cache.
func newCache() *cache {
	c := cache{
		m:    make(map[reflect.Type]*structInfo),
		conv: make(map[reflect.Kind]Converter),
		tag:  "json",
	}
	for k, v := range converters {
		c.conv[k] = v
	}
	return &c
}

// cache caches meta-data about a struct.
type cache struct {
	l    sync.RWMutex
	m    map[reflect.Type]*structInfo
	conv map[reflect.Kind]Converter
	tag  string
}

// get returns a cached structInfo, creating it if necessary.
func (c *cache) Get(t reflect.Type) *structInfo {
	c.l.RLock()
	info := c.m[t]
	c.l.RUnlock()
	if info == nil {
		info = c.create(t, nil)
		c.l.Lock()
		c.m[t] = info
		c.l.Unlock()
	}
	return info
}

// create creates a structInfo with meta-data about a struct.
func (c *cache) create(t reflect.Type, info *structInfo) *structInfo {
	if info == nil {
		info = &structInfo{fields: []*fieldInfo{}}
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		//		if field.Anonymous {
		//			ft := field.Type
		//			if ft.Kind() == reflect.Ptr {
		//				ft = ft.Elem()
		//			}
		//			if ft.Kind() == reflect.Struct {
		//				c.create(ft, info)
		//			}
		//		}
		c.createField(field, info)
	}
	return info
}

func (c *cache) parseTag(field reflect.StructField) (alias string) {
	if tag := field.Tag.Get(c.tag); tag != "" {
		alias = tag
	}
	if alias == "" {
		alias = field.Name
	}
	return alias
}

// createField creates a fieldInfo for the given field.
func (c *cache) createField(field reflect.StructField, info *structInfo) {
	alias := c.parseTag(field)
	if alias == "-" {
		// Ignore this field.
		return
	}
	//	fmt.Println("alias ----> :", alias, " ==> type: ", field.Type)
	info.fields = append(info.fields, &fieldInfo{
		typ:   field.Type,
		name:  field.Name,
		alias: alias,
	})
}

func (c *cache) decode(info *structInfo, dst reflect.Value, src map[string]interface{}) error {

	errors := make([]string, 0)
	for i := 0; i < len(info.fields); i++ {
		field := info.fields[i]
		fieldVal := dst.FieldByName(field.name)

		if !fieldVal.CanSet() || !fieldVal.IsValid() {
			appendErrors(errors, fmt.Errorf("fieldVal named `%s` cannot set", field.name))
			continue
		}

		value := src[field.alias]
		if value == nil {
			errors = appendErrors(errors, fmt.Errorf("no alias `%s` found in json", field.alias))
			continue
		}
		mapval := reflect.ValueOf(value) //map值的反射值
		if field.typ != mapval.Type() {
			conv := c.converter(field.typ)
			if conv == nil {
				errors = appendErrors(errors, fmt.Errorf("field.typ %v is no converter", field.typ))
				continue
			}

			mapval = conv(fmt.Sprintf("%v", value))
		}

		if mapval.IsValid() {
			fieldVal.Set(mapval)
		}
	}
	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

// converter returns the converter for a type.
func (c *cache) converter(t reflect.Type) Converter {
	return c.conv[t.Kind()]
}

// ----------------------------------------------------------------------------

type structInfo struct {
	fields []*fieldInfo
}

func (i *structInfo) get(alias string) *fieldInfo {
	for _, field := range i.fields {
		if strings.EqualFold(field.alias, alias) {
			return field
		}
	}
	return nil
}

type fieldInfo struct {
	typ   reflect.Type
	name  string // field name in the struct.
	ss    bool   // true if this is a slice of structs.
	alias string
}

// ----------------------------------------------------------------------------
