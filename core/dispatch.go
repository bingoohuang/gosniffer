package core

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type Dispatch struct {
	Plug    *Plug
	device  string
	payload []byte
}

func NewDispatch(plug *Plug, cmd *Cmd) *Dispatch {
	return &Dispatch{
		Plug:   plug,
		device: cmd.Device,
	}
}

func (d *Dispatch) Capture() {
	handle, err := pcap.OpenLive(d.device, 65535, false, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
		return
	}

	// set filter
	fmt.Println(d.Plug.BPF)
	if err := handle.SetBPFFilter(d.Plug.BPF); err != nil {
		log.Fatal(err)
	}

	// capture
	src := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := src.Packets()

	// set up assembly
	streamFactory := &ProtocolStreamFactory{
		dispatch: d,
	}
	streamPool := NewStreamPool(streamFactory)
	assembler := NewAssembler(streamPool)
	ticker := time.Tick(time.Minute)

	// loop until ctrl+z
	for {
		select {
		case packet := <-packets:
			if packet.NetworkLayer() == nil ||
				packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				fmt.Println("ERR : Unknown Packet -_-")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(
				packet.NetworkLayer().NetworkFlow(),
				tcp, packet.Metadata().Timestamp,
			)
		case <-ticker:
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}

type ProtocolStreamFactory struct {
	dispatch *Dispatch
}

func (m *ProtocolStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	r := tcpreader.NewReaderStream()
	direction := fmt.Sprintf("%s:%s->%s:%s", net.Src(), transport.Src(), net.Dst(), transport.Dst())
	fmt.Printf("# Start new stream: %s\n", direction)

	// decode packet
	go m.dispatch.Plug.ResolveStream(net, transport, direction, &r)

	return &r
}
