package core

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"plugin"

	"github.com/bingoohuang/gosniffer/plugs/http"
	"github.com/bingoohuang/gosniffer/plugs/mongodb"
	"github.com/bingoohuang/gosniffer/plugs/mysql"
	"github.com/bingoohuang/gosniffer/plugs/packet"
	"github.com/bingoohuang/gosniffer/plugs/redis"
	"github.com/bingoohuang/gosniffer/plugs/tcp"
	"github.com/google/gopacket"
)

type Plug struct {
	ResolveStream func(net gopacket.Flow, transport gopacket.Flow, direction string, r io.Reader)

	InternalPlugins  map[string]Plugin
	ExternalPlugList map[string]ExternalPlug
	dir              string
	BPF              string
	dumpPacket       bool
}

// Plugin internal plug-ins must implement this interface
// ResolvePacket - entry
// BPFFilter     - set BPF, like: mysql(tcp and port 3306)
// SetFlag       - plug-in params
// Version       - plug-in version
type Plugin interface {
	ResolveStream(net gopacket.Flow, transport gopacket.Flow, direction string, r io.Reader)
	BPFFilter() string
	SetFlag([]string)
	Version() string
}

type ExternalPlug struct {
	ResolvePacket func(net gopacket.Flow, transport gopacket.Flow, direction string, r io.Reader)
	BPFFilter     func() string
	SetFlag       func([]string)
	Name          string
	Version       string
}

func NewPlug() *Plug {
	var p Plug
	p.dir, _ = filepath.Abs("./plug/")
	p.LoadInternalPlugins()
	p.LoadExternalPlugins()

	return &p
}

func (p *Plug) LoadInternalPlugins() {
	p.InternalPlugins = map[string]Plugin{
		"mysql":   mysql.NewInstance(),
		"mongodb": mongodb.NewInstance(),
		"redis":   redis.NewInstance(),
		"http":    http.NewInstance(),
		"tcp":     tcp.NewInstance(),
		"packet":  packet.NewInstance(),
	}
}

func (p *Plug) LoadExternalPlugins() {
	dir, err := os.ReadDir(p.dir)
	if err != nil {
		return
	}

	p.ExternalPlugList = make(map[string]ExternalPlug)
	for _, fi := range dir {
		if fi.IsDir() || path.Ext(fi.Name()) != ".so" {
			continue
		}

		plug, err := plugin.Open(p.dir + "/" + fi.Name())
		if err != nil {
			panic(err)
		}

		versionFunc, err := plug.Lookup("Version")
		if err != nil {
			panic(err)
		}

		setFlagFunc, err := plug.Lookup("SetFlag")
		if err != nil {
			panic(err)
		}

		BPFFilterFunc, err := plug.Lookup("BPFFilter")
		if err != nil {
			panic(err)
		}

		ResolvePacketFunc, err := plug.Lookup("ResolvePacket")
		if err != nil {
			panic(err)
		}

		version := versionFunc.(func() string)()
		p.ExternalPlugList[fi.Name()] = ExternalPlug{
			ResolvePacket: ResolvePacketFunc.(func(net gopacket.Flow, transport gopacket.Flow, direction string, r io.Reader)),
			SetFlag:       setFlagFunc.(func([]string)),
			BPFFilter:     BPFFilterFunc.(func() string),
			Version:       version,
			Name:          fi.Name(),
		}
	}
}

func (p *Plug) ChangePath(dir string) { p.dir = dir }

func (p *Plug) PrintList() {
	// Print Internal Plug
	for inPlugName := range p.InternalPlugins {
		fmt.Println("internal plug : " + inPlugName)
	}

	// split
	fmt.Println("-- --- --")

	// print External Plug
	for exPlugName := range p.ExternalPlugList {
		fmt.Println("external plug : " + exPlugName)
	}
}

func (p *Plug) SetOption(plugName string, plugParams []string) {
	p.dumpPacket = plugName == "packet"
	// Load Internal Plug
	if pg, ok := p.InternalPlugins[plugName]; ok {
		p.ResolveStream = pg.ResolveStream
		pg.SetFlag(plugParams)
		p.BPF = pg.BPFFilter()

		return
	}

	// Load External Plug
	plug, err := plugin.Open("./plug/" + plugName)
	if err != nil {
		panic(err)
	}
	resolvePacket, err := plug.Lookup("ResolvePacket")
	if err != nil {
		panic(err)
	}
	setFlag, err := plug.Lookup("SetFlag")
	if err != nil {
		panic(err)
	}
	BPFFilter, err := plug.Lookup("BPFFilter")
	if err != nil {
		panic(err)
	}
	p.ResolveStream = resolvePacket.(func(net gopacket.Flow, transport gopacket.Flow, direction string, r io.Reader))
	setFlag.(func([]string))(plugParams)
	p.BPF = BPFFilter.(func() string)()
}
