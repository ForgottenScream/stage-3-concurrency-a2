/*
Pi-calculus-style model for Part 2 (drop-in + mailbox, with starvation)

Channels:

	read      -- used when the lawyer is reading, clients can wake her up directly
	waitRoom  -- models access to the waiting room
	mailBox   -- inbox for email appointments
	x         -- private channel for a single client-lawyer meeting

Meeting behaviour (same as in Part 1):

	SMeet(x) = x.met.0

Client behaviour (one client):

	Client2 = νx.(
	             read<x>. SMeet(x)                 // try to wake lawyer if she is reading
	           + waitRoom<x>. SMeet(x)             // else, try to join the waiting room
	           + mailBox<x>. SMeet(x)              // else, send email and wait for reply
	          )

Explanation:
  - νx creates a fresh private channel x for this client.
  - The three summands correspond to the three possible ways the client
    interacts with the lawyer in the Go code:
  - send x on 'read'        (direct meeting)
  - send x on 'waitRoom'    (meet after waiting)
  - send x on 'mailBox'     (meet via email)
  - In the Go implementation the order is to first try read, then waitRoom, then mailBox.
  - In the abstract model this is expressed as an external choice between the three actions.

Lawyer behaviour (one lawyer):

	Lawyer2 = rec L.
	            ( waitRoom(y). y.met. L           // serve a waiting-room client
	            + mailBox(y). y.met. L            // or serve a mailbox client
	            + read(y). y.met. L               // or serve a direct client
	            )

Notes:
  - y is the channel received from a client (the client's x).
  - After each meeting (y.met) the lawyer recurses to L to handle more clients.
  - In the Go implementation, the lawyer gives priority to waitRoom over mailBox over read using nested 'select' with 'default'.
  - In this abstract model we show the possible interactions, with an implementation detail enforced by Go's scheduling and 'select' structure.

Whole system (many clients in parallel with one lawyer):

	System = ( Client2_1 | Client2_2 | ... | Client2_n | Lawyer2 )

	where each call to client(waitRoom, read, mailBox, name) in Go corresponds to one Client2_i process, and the single lawyer goroutine corresponds to Lawyer2.

This model captures:
  - fresh per-client channels (νx and SMeet(x))
  - three possible interaction modes: direct, waiting-room, email
  - the possibility of starvation for mailbox clients because Lawyer2 can always keep synchronising on waitRoom(y) when there are many in-person clients, some Client2_i that used mailBox<x> may never see its SMeet(x) continuation executed in a particular run.
*/
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
)

func lawyer(waitRoom chan chan string, read chan chan string, mailBox chan chan string) {
	for {
		select {
		case clientChan := <-waitRoom:

			fmt.Println("[LAWYER] A client is waiting in the waiting room. Inviting them in.")
			fmt.Println("[LAWYER] Starting meeting (waiting-room client).")

			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			fmt.Println("[LAWYER] Finished meeting with waiting-room client")
			clientChan <- "met after waiting"

		default:

			fmt.Println("[LAWYER] No clients in waiting room.")

			select {
			case clientChan := <-mailBox:
				fmt.Println("[LAWYER] Picking a client from the mailbox (email appointment).")
				fmt.Println("[LAWYER] Starting meeting (mailbox client).")

				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

				fmt.Println("[LAWYER] Finished meeting with mailbox client.")
				clientChan <- "met via email"

			default:
				fmt.Println("[LAWYER] No mailbox clients. Lawyer is reading...")
				clientChan := <-read // blocks here until someone wakes her
				fmt.Println("[LAWYER] A client woke me up! Starting direct meeting.")

				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

				fmt.Println("[LAWYER] Finished direct meeting.")
				clientChan <- "met directly"
			}
		}
	}
}

func client(waitRoom chan chan string, read chan chan string, mailBox chan chan string, name string) {
	x := make(chan string)

	fmt.Println("[CLIENT:", name, "] arrived at the office.")

	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	select {
	case read <- x:
		fmt.Println("[CLIENT:", name, "] Lawyer is reading -> waking them up for a direct meeting.")
		msg := <-x
		fmt.Println(name, " ", msg)

	default:
		fmt.Println("[CLIENT:", name, "] Lawyer is not reading but is busy.")

		select {
		case waitRoom <- x:
			fmt.Println("[CLIENT:", name, "] Joined the waiting room.")
			msg := <-x
			fmt.Println(name, " ", msg)

		default:
			fmt.Println("[CLIENT:", name, "] Waiting room is full -> sending email to mailbox.")
			mailBox <- x
			msg := <-x
			fmt.Println(name, " ", msg)
		}
	}
}

func main() {
	n := 3

	waitRoom := make(chan chan string, n)
	read := make(chan chan string)
	mailBox := make(chan chan string, 100)

	fmt.Println("[MAIN] Starting system. (STARVATION DEMO)")

	fmt.Println("[MAIN] Waiting-room capacity:", n)

	go lawyer(waitRoom, read, mailBox)

	go func() {
		for i := 0; ; i++ {
			name := fmt.Sprintf("GREEDY_%d", i)
			go client(waitRoom, read, mailBox, name)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("STARVE_%s%d", randomdata.SillyName(), i)
		go client(waitRoom, read, mailBox, name)
	}

	simTime := 10 * time.Second
	fmt.Println("[MAIN] Running starvation demo for", simTime)
	time.Sleep(simTime)

	fmt.Println("[MAIN] Starvation demo finished.")
}
