package main

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"math/rand"
	"time"
)

func lawyer(waitRoom chan chan string, read chan chan string) {
	for {
		select {
		case clientChan := <-waitRoom:
			fmt.Println("[LAWYER] A client is waiting. Inviting next client in.")
			fmt.Println("[LAWYER] Starting meeting (waiting-room client).")

			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			fmt.Println("[LAWYER] Finished meeting with waiting client.")
			clientChan <- "met after waiting"

		default:
			fmt.Println("[LAWYER] No clients waiting. Lawyer is reading...")
			clientChan := <-read

			fmt.Println("[LAWYER] A client woke me up! Starting direct meeting.")
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			fmt.Println("[LAWYER] Finished meeting (direct.)")
			clientChan <- "met directly"
		}
	}
}

func client(waitRoom chan chan string, read chan chan string, name string) {
	x := make(chan string)

	fmt.Println("[CLIENT:", name, "] arrived.")
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	select {
	case read <- x:
		fmt.Println("[CLIENT:", name, "] The lawyer is reading -> wake them up.")
		msg := <-x
		fmt.Println(name, " ", msg)

	default:
		fmt.Println("[CLIENT: ", name, "] Lawyer is busy -> going to waiting room.")
		waitRoom <- x
		msg := <-x
		fmt.Println(name, " ", msg)
	}
}

func main() {

	m := 20
	n := 20

	waitRoom := make(chan chan string, n)
	law := make(chan chan string)

	go lawyer(waitRoom, law)

	for i := 0; i < m; i++ {
		go client(waitRoom, law, randomdata.SillyName())
	}

	time.Sleep(10 * time.Second)
}
