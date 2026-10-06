package utils

import "fmt"

func LogError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func LogDeferredError(fn func() error) {
	LogError(fn())
}
