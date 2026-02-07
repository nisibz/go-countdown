package main

import (
	"fmt"
	"time"
)

type Timer struct {
	Name string
	End  time.Time
}

func runTimer(t Timer, done chan<- string) {
	for {
		remaining := time.Until(t.End)

		if remaining <= 0 {
			done <- t.Name
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	timers := []Timer{
		{Name: "Timer A", End: time.Now().Add(5 * time.Second)},
		{Name: "Timer B", End: time.Now().Add(2 * time.Second)},
		{Name: "Timer C", End: time.Now().Add(6 * time.Second)},
	}

	done := make(chan string)

	for _, t := range timers {
		go runTimer(t, done)
	}

	for range timers {
		name := <-done
		fmt.Println("⏰ Finished:", name)
	}
}
