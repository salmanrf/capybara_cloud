package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

const default_ping_interval = 3 * time.Second

func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background()) 
	r, w := io.Pipe()
	done := make(chan struct{})
	reset_timer := make(chan time.Duration, 1)
	reset_timer <- time.Second

	go func () {
		Pinger(ctx, w, reset_timer)
		close(done)
	}()

	receive_ping := func (d time.Duration, r io.Reader) {
		if d > 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			reset_timer <- d 
		}

		buf := make([]byte, 1024)
		now := time.Now()
		n, err := r.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100 * time.Millisecond))
	}

	for i, v := range([]int64{0, 200, 300, 0, -1, -1, -1}) {
		fmt.Printf("Run %d:\n", i+1)
		receive_ping(time.Duration(v)*time.Millisecond, r)
	}

	cancel()
	<- done
}

func Pinger(ctx context.Context, w io.Writer, reset <- chan time.Duration) {
	var interval time.Duration
	select {
	case <- ctx.Done():
		return
	case interval = <- reset:
	default:
	}

	if interval <= 0 {
		interval = default_ping_interval
	}

	timer := time.NewTimer(interval)
	defer func () {
		if !timer.Stop() {
			<- timer.C
		}
	}()

	for {
		select {
		case <- ctx.Done():
			return
		case new_interval := <- reset:
			if !timer.Stop() {
				<- timer.C
			}

			if new_interval > 0 {
				interval = new_interval
			}
		case <- timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}

		_ = timer.Reset(interval)
	}
}

func NewKeeper() {
	done := make(chan bool)
	
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		fmt.Println("Connection established, starting heartbeat mechanism")
		
		buf := make([]byte, 4)

		for {
			err := conn.SetDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Read(buf)
			if err != nil {
				log.Fatal(err)			
			}

			fmt.Printf("%s\n", buf)
			
			err = conn.SetDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				log.Fatal(err)
			}
	
			_, err = conn.Write([]byte("pong"))
			if err != nil {
				log.Fatal(err)			
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		log.Fatal(err)
	}

	go func(conn net.Conn) {
		defer conn.Close()
	
		buf := make([]byte, 4)

		for {			
			err := conn.SetDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				log.Fatal(err)
			}
			
			_, err = conn.Write([]byte("ping"))
			if err != nil {
				log.Fatal(err)			
			}
			
			err = conn.SetDeadline(time.Now().Add(3 * time.Second))
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Read(buf)
			if err != nil {
				log.Fatal(err)			
			}

			fmt.Printf("%s\n", buf)

			err = conn.SetDeadline(time.Now().Add(3 * time.Second))
			time.Sleep(2 * time.Second)
		}
	}(conn)

	<- done
}

func NewUploader() {
	listener, err := net.Listen("tcp4", "127.0.0.1:")

	if err != nil {
		log.Fatal("Failed to initiate socket")
	}

	done := make(chan struct{})

	go func() {
		defer func() { done <- struct{}{} }()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal("Unable to accept connection", err)
			}

			go func (conn net.Conn) {
				defer func() {
					conn.Close()
					done <- struct{}{}  
				}()

				buff := make([]byte, 1024)
				for {
					n, err := conn.Read(buff)
					if err != nil {
						if err != io.EOF {
							fmt.Println("Error reading from connection buffer", err)
						}
						return
					}
					fmt.Printf("Received: %q\n", buff[:n])
				}
			}(conn)
		}
	}()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5_000 * time.Millisecond))
	defer cancel()
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp4", listener.Addr().String())
	if err != nil {
		log.Fatal("Unable to initate connection with server", err)
	}

	fmt.Println("Connection with server has been established")
	
	conn.Close()
	<- done
	listener.Close()
	<- done
}

func main() {
	path, err := exec.LookPath("tar")
	if err != nil {
		fmt.Println("Unable to find tar executable", err)
		os.Exit(1)
	}

	cmd := exec.Command(path, "-cvf", "samples.tar", "./samples")
	cmd.Stderr = os.Stdout

	if err := cmd.Start(); err != nil {
		fmt.Println("Unable to start deployment bundling", err)
		os.Exit(1)
	}
	
	if err := cmd.Wait(); err != nil {
		fmt.Println("Failed to execute deployment bundling", err)
		os.Exit(1)
	}

	fmt.Println("Deployment bundled successfully")

	// NewUploader()
	// NewKeeper()
	ExamplePinger()
}