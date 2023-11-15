package a

import (
	"fmt"
	"time" // want "imported forbidden package: time"

	"b" // want "imported forbidden package: net/http"
)

func a1() {
	fmt.Println(time.RFC3339)
}

func a2() {
	b.B1()
}
