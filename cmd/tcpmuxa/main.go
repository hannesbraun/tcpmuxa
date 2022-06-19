package main

import (
	"fmt"
	"github.com/hannesbraun/tcpmuxa/config"
	"github.com/hannesbraun/tcpmuxa/tcpmux"
	"log"
	"os"
	"strconv"
)

const version = "1.0.0"

func main() {
	fmt.Println("tcpmuxa", version)

	configPath := "tcpmuxa.conf"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	tcpmuxConfig := config.ReadConfig(configPath)

	port, err := strconv.Atoi(tcpmuxConfig.Vars["port"])
	if err != nil {
		log.Println("Unable to parse port number:", err)
		tcpmux.TCPMUX(tcpmuxConfig.ServiceDirectory)
	} else {
		tcpmux.TCPMUXWithPort(tcpmuxConfig.ServiceDirectory, port)
	}
}
