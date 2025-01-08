package testutil

import (
	"reflect"
	"testing"
)

func IsNil(i any) bool {
	if i == nil {
		return true
	}

	//nolint:exhaustive // default clause captures the rest
	switch reflect.TypeOf(i).Kind() {
	case reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}

func AssertNotNilf(t *testing.T, obj any, msg string, f ...any) {
	t.Helper()
	if IsNil(obj) {
		if msg == "" {
			t.Errorf("expected a value but got %v", obj)
			return
		}

		t.Errorf(msg, f...)
	}
}

func AssertNilf(t *testing.T, obj any, msg string, f ...any) {
	t.Helper()
	if !IsNil(obj) {
		if msg == "" {
			t.Errorf("expected nil but got %v", obj)
			return
		}

		t.Errorf(msg, f...)
	}
}

func AssertNil(t *testing.T, e error) {
	t.Helper()
	AssertNilf(t, e, "")
}
