package main

import "reflect"

func (g Generator) genSchema(name string, t reflect.Type, computed bool) {
	scope := "Computed"
	if !computed {
		scope = "Required"
	}
	if t.Kind() == reflect.Ptr {
		if scope == "Required" {
			scope = "Optional"
		}
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		g.p("\"%s\": %s")
	}
}
