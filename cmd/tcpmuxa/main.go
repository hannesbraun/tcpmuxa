package main

import (
	"github.com/hannesbraun/tcpmuxa/config"
	"github.com/hannesbraun/tcpmuxa/tcpmux"
	"os"
)

func main() {
	configPath := "tcpmuxa.conf"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	tcpmuxConfig := config.ReadConfig(configPath)

	tcpmux.TcpMux(tcpmuxConfig.ServiceDirectory)
}
