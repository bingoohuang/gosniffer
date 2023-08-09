package tcp

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/bingoohuang/tproxy/display"
	"github.com/bingoohuang/tproxy/hexdump"
	"github.com/google/gopacket"
)

const (
	Version = "0.1"

	CmdPort    = "-p"
	CmdWidth   = "-w"
	CmdVerbose = "-v"

	bufferSize = 1 << 20
)

type H struct {
	port         string
	version      string
	width        int
	printLock    sync.Mutex
	quiet        bool
	printStrings bool
	verbose      bool
}

func NewInstance() *H {
	return &H{
		port:    "80",
		version: Version,
		width:   32,
	}
}

func (m *H) ResolveStream(net, transport gopacket.Flow, direction string, buf io.Reader) {
	transportString := transport.String()

	src, _ := transport.Endpoints()
	srcValue := fmt.Sprintf("%v", src)

	dumper := hexdump.Config{Width: m.width, PrintStrings: m.verbose}

	data := make([]byte, bufferSize)
	id := 0
	for {
		n, err := buf.Read(data)
		if n > 0 && !m.quiet {
			id++
			m.print(srcValue, id, data, n, dumper)
		}

		if err != nil {
			log.Printf("[%s] %s error: %v", direction, transportString, err)
		}

		if n == 0 {
			break
		}
	}
}

func (m *H) print(source string, id int, data []byte, n int, dumper hexdump.Config) {
	m.printLock.Lock()
	defer m.printLock.Unlock()

	display.PrintfWithTime("from %s [%d]:\n", source, id)
	fmt.Println(dumper.Dump(data[:n]))
}

func (m *H) BPFFilter() string { return "tcp and port " + m.port }
func (m *H) Version() string   { return Version }

func (m *H) SetFlag(flg []string) {
	c := len(flg)
	if c == 0 {
		return
	}

	if c>>1 == 0 {
		fmt.Println("ERR : tcp Number of parameters")
		os.Exit(1)
	}

	for i := 0; i < c; i += 2 {
		key := flg[i]
		switch key {
		case CmdVerbose:
			m.verbose = true
			i -= 1
		case CmdWidth:
			width, err := strconv.Atoi(flg[i+1])
			if err != nil {
				panic("ERR : width")
			}
			m.width = width
		case CmdPort:
			port, err := strconv.Atoi(flg[i+1])
			if err != nil {
				panic("ERR : port")
			}

			if port < 0 || port > 65535 {
				panic("ERR : port(0-65535)")
			}

			m.port = flg[i+1]
		default:
			panic("ERR : mysql's params")
		}
	}
}
