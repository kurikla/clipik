package main

import (
	"errors"
	"fmt"
	"reflect"
)

func setField(cfg *Config, field string, v any) error {
	rv := reflect.ValueOf(cfg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("setField: cfg must be non-nil pointer, got %T", cfg)
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("setField: cfg must point to struct, got %s", elem.Kind())
	}

	f := elem.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("setField: field %q not found in %T", field, cfg)
	}

	if !f.CanSet() {
		return fmt.Errorf("setField: field %q is not settable", field)
	}

	fv := reflect.ValueOf(v)
	if !fv.IsValid() {
		return errors.New("setField: value is nil")
	}

	if !fv.Type().AssignableTo(f.Type()) {
		return fmt.Errorf(
			"setField: cannot assign %s to field %q of type %s",
			fv.Type(), field, f.Type(),
		)
	}

	f.Set(fv)
	return nil
}

func getField(cfg *Config, field string) (any, error) {
	rv := reflect.ValueOf(cfg).Elem()
	f := rv.FieldByName(field)

	if !f.IsValid() {
		return nil, fmt.Errorf("field %q not found", field)
	}

	return f.Interface(), nil
}
