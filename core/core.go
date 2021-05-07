package core

type Core struct {
	Version string
}

var cxt Core

func New() Core {
	cxt.Version = "0.1"
	return cxt
}

func (c *Core) Run() {
	plug := NewPlug()
	cmd := NewCmd(plug)
	cmd.Parse()
	NewDispatch(plug, cmd).Capture()
}
