package parser

import (
	"fmt"
	"reflect"
)

type Foo struct{}

func getMethods(service string) {
	fooType := reflect.TypeOf(Foo{})
	for i := 0; i < fooType.NumMethod(); i++ {
		method := fooType.Method(i)
		fmt.Println(method.Name)
	}
}
