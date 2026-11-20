package cfgstore

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mikeschinkel/go-dt"
)

var fieldValidatorType = reflect.TypeFor[FieldValidator]()

// validateFields calls Validate on every value inside rc that implements
// FieldValidator, in nested structs, slice and array elements and map values,
// and returns each error tagged with fp and the value's dotted JSON path.
func validateFields(rc any, fp dt.Filepath) (errs []error) {
	v := reflect.ValueOf(rc)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			goto end
		}
		v = v.Elem()
	}
	errs = walkFields(v, "", fp, errs)
end:
	return errs
}

// walkFields validates v, found at path, then the values inside it.
func walkFields(v reflect.Value, path string, fp dt.Filepath, errs []error) []error {
	var err error

	if path != "" {
		err = validateField(v)
	}
	if err != nil {
		errs = append(errs, NewErr(
			ErrInvalidConfigValue,
			"config_file", fp,
			"json_path", path,
			err,
		))
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			errs = walkFields(v.Elem(), path, fp, errs)
		}
	case reflect.Struct:
		errs = walkStruct(v, path, fp, errs)
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			errs = walkFields(v.Index(i), joinJSONPath(path, strconv.Itoa(i)), fp, errs)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			errs = walkFields(iter.Value(), joinJSONPath(path, fmt.Sprint(iter.Key().Interface())), fp, errs)
		}
	default:
	}
	return errs
}

// walkStruct walks each exported field of v that encoding/json/v2 decodes.
func walkStruct(v reflect.Value, path string, fp dt.Filepath, errs []error) []error {
	t := v.Type()
	for i := range t.NumField() {
		sf := t.Field(i)
		if !sf.IsExported() && !sf.Anonymous {
			continue
		}
		name, inline, skip := jsonFieldName(sf)
		if skip {
			continue
		}
		fieldPath := path
		if !inline {
			fieldPath = joinJSONPath(path, name)
		}
		errs = walkFields(v.Field(i), fieldPath, fp, errs)
	}
	return errs
}

// jsonFieldName returns the JSON name encoding/json/v2 gives sf, whether its
// fields are inlined into the parent object, and whether it is skipped.
func jsonFieldName(sf reflect.StructField) (name string, inline, skip bool) {
	tag, hasTag := sf.Tag.Lookup("json")
	if tag == "-" {
		skip = true
		goto end
	}
	name, _, _ = strings.Cut(tag, ",")
	inline = strings.Contains(tag, ",inline")
	if sf.Anonymous && (!hasTag || name == "") {
		inline = true
	}
	if !sf.IsExported() && !inline {
		skip = true
		goto end
	}
	if name == "" {
		name = sf.Name
	}
end:
	return name, inline, skip
}

// validateField calls v's Validate when v, or a pointer to it, implements
// FieldValidator.
func validateField(v reflect.Value) (err error) {
	switch {
	case v.Kind() == reflect.Pointer, v.Kind() == reflect.Interface:
		// The walk validates the value a pointer or interface holds, so that a
		// pointer to a FieldValidator is validated once.
	case v.Type().Implements(fieldValidatorType):
		if v.CanInterface() {
			err = v.Interface().(FieldValidator).Validate()
		}
	case reflect.PointerTo(v.Type()).Implements(fieldValidatorType):
		if !v.CanAddr() {
			// Map values are not addressable; validate a copy.
			cp := reflect.New(v.Type()).Elem()
			cp.Set(v)
			v = cp
		}
		if v.Addr().CanInterface() {
			err = v.Addr().Interface().(FieldValidator).Validate()
		}
	}
	return err
}

func joinJSONPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}
