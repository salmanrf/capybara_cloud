package deployment

import (
	"fmt"
	"net"
	"sync"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	paService := NewPortAllocatorService()
	
	t.Run("should not return port less than 1024", func (t *testing.T) {
		defer paService.Clear()
		
		got_port, err := paService.GetFreePort()
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}
		if got_port < 1024 {
			t.Errorf("got port %d, want port greater than or equal to 1024", got_port)
		}
	})

	t.Run("should return the first unused port", func (t *testing.T) {
		defer paService.Clear()
		
		tests := []struct {
			used_port []int
			want_port int
			desc string
		}{
			{
				used_port: []int{1024, 1025, 1026, 1027, 1028, 1029},
				want_port: 1030,
				desc: "should return the first unused port after consecutive used ports",
			},
			{
				used_port: []int{1024, 1025, 1026, 1027, 1028, 1029, 1030, 1031, 1032, 1033, 1034},
				want_port: 1035,
				desc: "should return the first unused port after consecutive used ports",
			},
			{
				used_port: []int{1036, 3000, 8080, 5555},
				want_port: 1024,
				desc: "should return the first unused port after randomly used ports",
			},
		}

		for _, tt := range tests {
			t.Run(tt.desc, func(t *testing.T) {
				for _, port := range tt.used_port {
					listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
					if err != nil {
						t.Fatal(err)
					}
					defer listener.Close()
		
					go func () {
						for {
							conn, err := listener.Accept()
							if err != nil {
								break
							}
							conn.Close()
						}
					}()
				}
				
				want_port := tt.want_port
				got_port, err := paService.GetFreePort()
				
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
				if got_port != want_port {
					t.Errorf("got port %d, want %d", got_port, want_port)
				}
			})
		}
	})

	t.Run("should not reuse previously assigned ports", func (t *testing.T) {
		defer paService.Clear()
		
		tests := []struct {
			it int
		}{
			{1},
			{3},
			{5},
			{8},
			{13},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("should assign %d different ports", tt.it), func (t *testing.T) {
				ports := make(map[int]int)

				for i := 0; i < tt.it; i++ {
					p, err := paService.GetFreePort()
					if err != nil {
						t.Fatal(err)
					}
					ports[p] += 1
				}

				for p, n := range ports {
					if n > 1 {
						t.Errorf("got port %d used %d times, want 1", p, n)
					}
				}
			})
		}
	})

	t.Run("should not return the same ports if called concurrently", func (t *testing.T) {
		defer paService.Clear()
		
		done := make(chan struct{})
		defer func () {
			close(done)
		}()
		
		tests := []struct{
			n_callers int
		}{
			{1},
			{3},
			{5},
			{20},
			{100},
		}

		for i, tt := range tests {
			t.Run(fmt.Sprintf("should return %d different ports", tt.n_callers), func (t *testing.T) {
				ports := make(map[int]int)
				results := make(chan int, tt.n_callers)

				var wg sync.WaitGroup 
				wg.Add(tt.n_callers)
				
				for j := 0; j < tt.n_callers; j++ {
					go func() {
						t.Run(fmt.Sprintf("act of test %d, it %d", i + 1, j), func (t *testing.T) {
							defer wg.Done()
							
							p, err := paService.GetFreePort()
							if err != nil {
								t.Fatal(err)
							}
							results <- p 
						})
					}()
				}

				go func () {
					for p := range results {
						ports[p] += 1
					}
				}()

				wg.Wait()
				close(results)

				got_n_different_ports := 0
				want_n_different_ports := tt.n_callers
				for p, n := range ports {
					got_n_different_ports ++

					// * Bind port to claim as long as the test runs
					listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", p))
					go func () {
						<- done
						listener.Close()
					}()

					if n > 1 {
						t.Errorf("got %d assignments for port %d, want 1", n, p)
					}
				}
				if got_n_different_ports < tt.n_callers {
					t.Errorf("got %d different ports, want %d", got_n_different_ports, want_n_different_ports)
				}
			})
		}
	})
}