package packet

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/google/gopacket"
)

const (
	Version = "0.1"

	CmdPort = "-p"
)

type H struct {
	port    string
	version string
}

func NewInstance() *H {
	return &H{
		port:    "80",
		version: Version,
	}
}

func (m *H) ResolveStream(net, transport gopacket.Flow, direction string, buf io.Reader) {
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
