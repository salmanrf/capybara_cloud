package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	t.Run("should cancel all but one connection after canceled", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(
			context.Background(),
			time.Now().Add(10 * time.Second),
		)

		listener, err := net.Listen("tcp", "127.0.0.1:")
		if err != nil {
			log.Fatal("unable to open socket")
		}
		defer listener.Close()

		go func () {
			conn, err := listener.Accept()
			if err == nil {	
				conn.Close()
			} else {
				fmt.Println("TCP connection err", err.Error())
			}
		}()

		dial := func (ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
			defer wg.Done()

			var dialer net.Dialer
			c, err := dialer.DialContext(ctx, "tcp", address)
			if err != nil {
				fmt.Printf("client %d was unable to dial server\n", id)
				return
			}
			c.Close()

			select {
				case <-ctx.Done():
				case response <- id:
			}
		}

		reschan := make(chan int, 3)		
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go dial(ctx, listener.Addr().String(), reschan, i + 1, &wg)
		}

		response := <- reschan

		cancel()
		wg.Wait()
		close(reschan)

		if ctx.Err() != context.Canceled {
			t.Errorf("expected canceled context; actual: %s\n", ctx.Err())
		}

		t.Logf("dialer retrieved the resource: %d\n", response)
	})
}

func TestDeadline(t *testing.T) {
	t.Run("should terminate connection on deadline reached", func(t *testing.T) {
		sync := make(chan struct{})

		listener, err := net.Listen("tcp", "127.0.0.1:")
		if err != nil {
			t.Fatal(err)
		}

		go func () {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}
			defer func() {
				conn.Close()
				close(sync)
			}()

			err = conn.SetDeadline(time.Now().Add(5 * time.Second))

			buf := make([]byte, 1) 
			_, err = conn.Read(buf)
			nErr, ok := err.(net.Error)
			if !ok || !nErr.Timeout() {
				t.Errorf("expected timeout, actual: %v\n", err)
			}

			sync <- struct{}{}

			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
			
			_, err = conn.Read(buf)
			if err != nil {
				t.Error(err)
			}
		}()

		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		<- sync
		_, err = conn.Write([]byte("1"))
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1)

		_, err = conn.Read(buf)
		if err != io.EOF {
			t.Errorf("expected server termination, actual: %v", err)
		}
	})
}