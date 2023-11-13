package main

import (
	"fmt"
	"time"

	hook "github.com/robotn/gohook"
)

func main() {
	hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
		fmt.Println("asdsadd")
	})
	go func() {
		s := hook.Start()
		fmt.Println("prasdsa")
		<-hook.Process(s)
		fmt.Println("done")
	}()
	time.Sleep(time.Second * 5)
	hook.End()
	time.Sleep(time.Second * 1)
}
