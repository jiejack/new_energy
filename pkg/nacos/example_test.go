//go:build ignore

package nacos

import (
	"fmt"
)

func ExampleApplyOptions() {
	opts := ApplyOptions(
		WithNamespace("production"),
		WithServiceName("my-service"),
		WithServicePort(8080),
	)
	fmt.Println(opts.Namespace)
	fmt.Println(opts.ServiceName)
	fmt.Println(opts.ServicePort)
}
