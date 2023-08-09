package redis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/gopacket"
)

type Redis struct {
	cmd     chan string
	done    chan bool
	version string
	port    int
}

const (
	Port    int    = 6379
	Version string = "0.1"
	CmdPort string = "-p"
)

var redis = &Redis{
	port:    Port,
	version: Version,
}

func NewInstance() *Redis {
	return redis
}

func (red Redis) ResolveStream(net, transport gopacket.Flow, direction string, r io.Reader) {
	buf := bufio.NewReader(r)
	var cmd string
	cmdCount := 0
	for {

		line, _, _ := buf.ReadLine()

		if len(line) == 0 {
			buff := make([]byte, 1)
			_, err := r.Read(buff)
			if err == io.EOF {
				red.done <- true
				return
			}
		}

		// Filtering useless data
		if !strings.HasPrefix(string(line), "*") {
			continue
		}

		// Do not display
		if strings.EqualFold(transport.Src().String(), strconv.Itoa(red.port)) == true {
			continue
		}

		// run
		l := string(line[1])
		cmdCount, _ = strconv.Atoi(l)
		cmd = ""
		for j := 0; j < cmdCount*2; j++ {
			c, _, _ := buf.ReadLine()
			if j&1 == 0 {
				continue
			}
			cmd += " " + string(c)
		}
		fmt.Println(cmd)
	}
}

func (red *Redis) SetFlag(flg []string) {
	c := len(flg)
	if c == 0 {
		return
	}
	if c>>1 != 1 {
		panic("ERR : Redis num of params")
	}
	for i := 0; i < c; i += 2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val)
			redis.port = port
			if err != nil {
				panic("ERR : Port error")
			}
			if port < 0 || port > 65535 {
				panic("ERR : Port(0-65535)")
			}
			break
		default:
			panic("ERR : redis's params")
		}
	}
}

func (red *Redis) BPFFilter() string {
	return "tcp and port " + strconv.Itoa(redis.port)
}

func (red *Redis) Version() string {
	return red.version
}
