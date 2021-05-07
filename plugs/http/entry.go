package http

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket"
)

const (
	Port    = "80"
	Version = "0.1"
)

const (
	CmdPort = "-p"
	CmdBody = "-b"
)

type H struct {
	port    string
	body    string
	version string
}

func NewInstance() *H {
	return &H{port: Port, version: Version}
}

func (m *H) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {
	bio := bufio.NewReader(buf)
	transportString := transport.String() + ": "

	src, _ := transport.Endpoints()
	srcValue := fmt.Sprintf("%v", src)
	reqBody := m.showBody("req")
	rspBody := m.showBody("rsp")

	if srcValue == m.port {
		for {
			resp, err := http.ReadResponse(bio, nil)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Println(transportString + "[RESPONSE EOF]")
				return
			}

			if err != nil {
				continue
			}

			dump, _ := httputil.DumpResponse(resp, reqBody)
			_ = resp.Body.Close()

			log.Print("\n", string(dump))
		}
	} else {
		for {
			req, err := http.ReadRequest(bio)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Println(transportString + "[REQUEST EOF]")
				return
			}

			if err != nil {
				continue
			}

			dump, _ := httputil.DumpRequest(req, rspBody)
			_ = req.Body.Close()

			log.Print("\n", string(dump))
		}
	}
}

func (m *H) showBody(key string) bool {
	return m.body == "all" || strings.Contains(m.body, key)
}

func (m *H) BPFFilter() string { return "tcp and port " + m.port }
func (m *H) Version() string   { return Version }

func (m *H) SetFlag(flg []string) {
	c := len(flg)
	if c == 0 {
		return
	}

	if c>>1 == 0 {
		fmt.Println("ERR : Http Number of parameters")
		os.Exit(1)
	}

	for i := 0; i < c; i += 2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val)
			m.port = val

			if err != nil {
				panic("ERR : port")
			}

			if port < 0 || port > 65535 {
				panic("ERR : port(0-65535)")
			}
		case CmdBody:
			m.body = val
		default:
			panic("ERR : mysql's params")
		}
	}
}
