package main

import (
	"testing"
	"time"
)

func TestApp(t *testing.T) {
	ch := make(chan int, 2)
	time.AfterFunc(5 * time.Second, func() {
		t.Log("d")
		ch <- 2
		t.Log("e")
	})
	for {
		select {
		case <- ch:
			t.Log("a")
			break
		}
		t.Log("b")
		break
	}
	t.Log("c")
}
