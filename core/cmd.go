package core

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	InternalCmdPrefix = "--"
)

type Cmd struct {
	Device     string
	plugHandle *Plug
}

func NewCmd(p *Plug) *Cmd {
	return &Cmd{plugHandle: p}
}

func (cm *Cmd) Parse() {
	if len(os.Args) <= 1 {
		cm.printHelpMessage()
		os.Exit(1)
	}

	//parse command
	firstArg := os.Args[1]
	if strings.HasPrefix(firstArg, InternalCmdPrefix) {
		cm.parseInternalCmd()
	} else {
		cm.parsePlugCmd()
	}
}

// parseInternalCmd parse internal command like --help, --env, --device.
func (cm *Cmd) parseInternalCmd() {
	arg := os.Args[1]
	cmd := strings.TrimPrefix(arg, InternalCmdPrefix)

	switch cmd {
	case "help":
		cm.printHelpMessage()
	case "env":
		fmt.Println("External plug-in path : " + cm.plugHandle.dir)
	case "list":
		cm.plugHandle.PrintList()
	case "ver":
		fmt.Println(cxt.Version)
	case "dev":
		cm.printDevices()
	}
	os.Exit(1)
}

//usage
func (cm *Cmd) printHelpMessage() {
	fmt.Println("=========================== Usage =================================")
	fmt.Println("gosniffer [device] [plug] [plug's params(optional)]")
	fmt.Println()
	fmt.Println("[exp]")
	fmt.Println("   gosniffer en0 redis                           Capture redis packet")
	fmt.Println("   gosniffer en0 mysql -p 3306                   Capture mysql packet")
	fmt.Println("   gosniffer en0 http -p 3306                    Capture http packet without body")
	fmt.Println("   gosniffer en0 http -p 3306 -b [all/req/rsp]   Capture http packet with body of request or response")
	fmt.Println()
	fmt.Println("   gosniffer --[commend]")
	fmt.Println("    --help \"this page\"")
	fmt.Println("    --env  \"environment variable\"")
	fmt.Println("    --list \"Plug-in list\"")
	fmt.Println("    --ver  \"version\"")
	fmt.Println("    --dev  \"device\"")
	fmt.Println("[exp]")
	fmt.Println("    gosniffer --list \"show all plug-in\"")
	fmt.Println()
	fmt.Println("=============================== Devices ================================")
	cm.printDevices()
}

func (cm *Cmd) printPlugins() {
	l := len(cm.plugHandle.InternalPlugins)
	l += len(cm.plugHandle.ExternalPlugList)
	fmt.Println("# Number of plug-ins : " + strconv.Itoa(l))
}

func (cm *Cmd) printDevices() {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					fmt.Println("[device] : " + iface.Name + " : " + iface.HardwareAddr.String() + "  " + ip4.String())
				}
			}
		}
	}
}

//Parameters needed for plug-ins
func (cm *Cmd) parsePlugCmd() {
	if len(os.Args) < 3 {
		fmt.Println("not found [Plug-in name]")
		fmt.Println("gosniffer [device] [plug] [plug's params(optional)]")
		os.Exit(1)
	}

	cm.Device = os.Args[1]
	plugName := os.Args[2]
	plugParams := os.Args[3:]
	cm.plugHandle.SetOption(plugName, plugParams)
}
