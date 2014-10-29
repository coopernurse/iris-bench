package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/project-iris/iris-go.v1"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	var mode string
	var seconds int
	var concur int
	var runners int
	flag.StringVar(&mode, "m", "client", "Mode to run as: (echo / add / server)")
	flag.IntVar(&seconds, "s", 10, "Seconds to run client for")
	flag.IntVar(&concur, "c", 1, "Concurrency per test runner")
	flag.IntVar(&runners, "r", 1, "Test runners to spawn")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	log.Println("starting with mode:", mode)

	if mode == "server" {
		server()
	} else {
		req := BenchReq{
			Fx:      mode,
			Seconds: seconds,
			Concur:  concur,
		}
		log.Printf("Starting bench: runners=%d req=%+v\n", runners, req)
		resp := benchCluster(runners, req)
		log.Println("Bench done:", resp)
		log.Printf("Req/sec: %.2f\n", float64(resp.Success)/(float64(resp.Duration)/1e9))
		log.Println("Timeouts:", resp.Timeout)
		log.Println("Bad responses:", resp.BadResponse)
	}
}

/////////////////////////////////////
// Client //
////////////

type BenchReq struct {
	Fx      string
	Seconds int
	Concur  int
}

type BenchResp struct {
	Duration    time.Duration
	Success     int
	Timeout     int
	BadResponse int
}

type Result int

const (
	ResultOk Result = iota
	ResultTimeout
	ResultBadResponse
)

func benchCluster(runners int, req BenchReq) BenchResp {
	conn, err := iris.Connect(55555)
	if err != nil {
		log.Fatalf("failed to connect to the Iris relay: %v.", err)
	}
	defer conn.Close()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("failed to marshal req: %v", err)
	}

	out := make(chan BenchResp)
	aggCh := make(chan BenchResp)

	go func() {
		start := time.Now()
		numOk := 0
		numTimeout := 0
		numBadResp := 0
		for r := range out {
			numOk += r.Success
			numBadResp += r.BadResponse
			numTimeout += r.Timeout

			log.Println("Got sub-bench response: ", r)
		}
		aggCh <- BenchResp{
			Duration:    time.Now().Sub(start),
			Success:     numOk,
			Timeout:     numTimeout,
			BadResponse: numBadResp,
		}
	}()

	wg := sync.WaitGroup{}
	timeout := time.Second * time.Duration(req.Seconds+10)
	for i := 0; i < runners; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if reply, err := conn.Request("bench", reqBytes, timeout); err == nil {
				var r BenchResp
				err = json.Unmarshal(reply, &r)
				if err == nil {
					out <- r
				} else {
					log.Fatalf("failed to unmarshal reply: %v - %v", err, string(reply))
				}
			} else {
				log.Println("ERROR: timeout of bench request")
				out <- BenchResp{Timeout: 1}
			}
		}()
	}
	wg.Wait()
	close(out)

	return <-aggCh
}

func bench(req BenchReq) BenchResp {
	out := make(chan Result)
	done := make(chan bool)
	benchRespCh := make(chan BenchResp)

	fx := add
	if req.Fx == "echo" {
		fx = echo
	}

	log.Printf("Starting %d worker(s)\n", req.Concur)
	for i := 0; i < req.Concur; i++ {
		go benchWorker(fx, out, done)
	}

	go func() {
		start := time.Now()
		numOk := 0
		numTimeout := 0
		numBadResp := 0
		for res := range out {
			switch res {
			case ResultOk:
				numOk++
			case ResultTimeout:
				numTimeout++
			case ResultBadResponse:
				numBadResp++
			}
		}
		benchRespCh <- BenchResp{
			Duration:    time.Now().Sub(start),
			Success:     numOk,
			Timeout:     numTimeout,
			BadResponse: numBadResp,
		}
	}()

	log.Println("Reading replies")

	time.Sleep(time.Second * time.Duration(req.Seconds))

	log.Println("Stopping workers")
	for i := 0; i < req.Concur; i++ {
		done <- true
	}
	close(out)

	return <-benchRespCh
}

func benchWorker(gen reqGen, out chan Result, done chan bool) {
	conn, err := iris.Connect(55555)
	if err != nil {
		log.Fatalf("failed to connect to the Iris relay: %v.", err)
	}
	defer conn.Close()

	timeout := time.Second * 10

	for {
		select {
		case <-done:
			return
		default:
			dest, req, expReply := gen()
			if reply, err := conn.Request(dest, req, timeout); err == nil {
				if bytes.Equal(reply, expReply) {
					out <- ResultOk
				} else {
					out <- ResultBadResponse
				}
			} else {
				out <- ResultTimeout
			}
		}
	}
}

type reqGen func() (string, []byte, []byte)

func add() (string, []byte, []byte) {
	x := rand.Intn(99999)
	y := rand.Intn(99999)

	request := []byte(fmt.Sprintf("%d %d", x, y))
	reply := []byte(strconv.Itoa(x + y))
	return "add", request, reply
}

func echo() (string, []byte, []byte) {
	b := []byte(randSeq(20))
	return "echo", b, b
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

/////////////////////////////////////
// Server //
////////////

func benchSvr(req []byte) ([]byte, error) {
	var benchReq BenchReq
	err := json.Unmarshal(req, &benchReq)
	if err != nil {
		return nil, err
	}

	benchResp := bench(benchReq)
	return json.Marshal(benchResp)
}

func echoSvr(req []byte) ([]byte, error) {
	return req, nil
}

func addSvr(req []byte) ([]byte, error) {
	parts := strings.Split(string(req), " ")

	sum := 0
	if len(parts) > 1 {
		x, _ := strconv.Atoi(parts[0])
		y, _ := strconv.Atoi(parts[1])
		sum = x + y
	}
	return []byte(strconv.Itoa(sum)), nil
}

func NewFxHandler(handler func(req []byte) ([]byte, error)) *FxHandler {
	return &FxHandler{handler: handler}
}

type FxHandler struct {
	handler func(req []byte) ([]byte, error)
}

func (b *FxHandler) Init(conn *iris.Connection) error         { return nil }
func (b *FxHandler) HandleBroadcast(msg []byte)               {}
func (b *FxHandler) HandleRequest(req []byte) ([]byte, error) { return b.handler(req) }
func (b *FxHandler) HandleTunnel(tun *iris.Tunnel)            {}
func (b *FxHandler) HandleDrop(reason error)                  {}

func server() {
	echo, err := iris.Register(55555, "echo", NewFxHandler(echoSvr), nil)
	if err != nil {
		log.Fatalf("failed to register echo to the Iris relay: %v.", err)
	}
	defer echo.Unregister()

	add, err := iris.Register(55555, "add", NewFxHandler(addSvr), nil)
	if err != nil {
		log.Fatalf("failed to register add to the Iris relay: %v.", err)
	}
	defer add.Unregister()

	benchS, err := iris.Register(55555, "bench", NewFxHandler(benchSvr), nil)
	if err != nil {
		log.Fatalf("failed to register bench to the Iris relay: %v.", err)
	}
	defer benchS.Unregister()

	for {
		time.Sleep(time.Second)
	}

}
