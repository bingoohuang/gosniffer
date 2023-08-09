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

	if d.Plug.dumpPacket {
		d.dumpPackets(packets)
		return
	}
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
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}

			netLayer := packet.NetworkLayer()
			transportLayer := packet.TransportLayer()
			if netLayer == nil || transportLayer == nil ||
				transportLayer.LayerType() != layers.LayerTypeTCP {
				fmt.Println("ERR : Unknown Packet -_-")
				continue
			}

			tcp := transportLayer.(*layers.TCP)
			assembler.AssembleWithTimestamp(
				netLayer.NetworkFlow(),
				tcp, packet.Metadata().Timestamp,
			)
		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}

func (d *Dispatch) dumpPackets(packets chan gopacket.Packet) {
	// loop until ctrl+z
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}

			netLayer := packet.NetworkLayer()
			transportLayer := packet.TransportLayer()
			if netLayer == nil || transportLayer == nil ||
				transportLayer.LayerType() != layers.LayerTypeTCP {
				fmt.Println("ERR : Unknown Packet -_-")
				continue
			}

			fmt.Println(packet.Dump())
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
