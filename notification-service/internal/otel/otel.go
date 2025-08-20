package otel

import "log"

var Shutdown = func() {}

func Init(serviceName, endpoint string) func() {
	log.Printf("[otel] Init tracing for %s at %s", serviceName, endpoint)
	return func() { log.Println("[otel] Shutdown tracing") }
}
