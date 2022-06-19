package main

import (
	"fmt"
	"github.com/hannesbraun/tcpmuxa/config"
	"github.com/hannesbraun/tcpmuxa/tcpmux"
	"os"
)

const version = "1.0.0"

func main() {
	fmt.Println("tcpmuxa", version)

	configPath := "tcpmuxa.conf"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	tcpmuxConfig := config.ReadConfig(configPath)

	tcpmux.TCPMUX(tcpmuxConfig.ServiceDirectory)
}
