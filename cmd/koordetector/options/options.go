package options

import "flag"

var (
	ServerAddr  = flag.String("addr", ":9416", "port of koordlet server")
	EnablePprof = flag.Bool("enable-pprof", false, "Enable pprof for controller manager.")
	PprofAddr   = flag.String("pprof-addr", ":9417", "The address the pprof binds to.")
)
