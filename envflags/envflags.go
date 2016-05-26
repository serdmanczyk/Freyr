// Package envflags package houses the SetFlags and ConfigEmpty methods for
// conveniently specifiying config via a single struct.  Using tags
// on the struct fields, both the cli flag option name and environment
// variable name can be specified it can be specified.  Environment
// variable values will be defaults, with cli options taking preference
// if specified.
//
// example struct:
//    type config struct {
//        StructField string `flag:"struct_field" env:"APP_STRUCTFIELD"`
//    }
//
package envflags

import (
	"flag"
	"os"
	"reflect"
)

// SetFlags takes a config struct and set its fields as options
// via the flags library, with the default value being the field's
// environment variable.
func SetFlags(s interface{}) bool {
	val := reflect.ValueOf(s).Elem()

	for i := 0; i < val.NumField(); i++ {
		fieldSpec := val.Type().Field(i)
		flagName := fieldSpec.Tag.Get("flag")
		envName := fieldSpec.Tag.Get("env")

		fieldAddr := val.Field(i).Addr().Interface().(*string)
		flag.StringVar(fieldAddr, flagName, os.Getenv(envName), envName)
	}

	return false
}

// ConfigEmpty takes a config struct and determines if any of its field's
// values are their zero value, meaning that their associated config flag
// and environment variable were not specified.
func ConfigEmpty(s interface{}) bool {
	val := reflect.ValueOf(s).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		zero := reflect.Zero(field.Type())
		if zero.String() == field.String() {
			return true
		}
	}

	return false
}
