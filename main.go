package main

import (
	"fmt"
	"time"
)

func main() {
	end := time.Now().Add(10 * time.Second)

	for {
		remaining := time.Until(end)

		if remaining <= 0 {
			fmt.Println("⏰ Done!")
			break
		}

		fmt.Println("Remaining:", remaining.Round(time.Second))
		time.Sleep(1 * time.Second)
	}
}
