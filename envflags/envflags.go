package envflags

import (
	"flag"
	"os"
	"reflect"
)

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
