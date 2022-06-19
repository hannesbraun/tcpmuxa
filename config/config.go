package config

import (
	"bufio"
	"github.com/hannesbraun/tcpmuxa/tcpmux"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Vars             map[string]string
	ServiceDirectory map[string]tcpmux.Service
}

// ReadConfig reads the configuration out of a given path to a configuration file
func ReadConfig(path string) Config {
	// Open/close the config file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	vars := map[string]string{}
	serviceDirectory := map[string]tcpmux.Service{}

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) <= 0 {
			// Empty line
			continue
		} else if line[0] == '#' {
			// Comment
			continue
		} else if line[0] == '$' {
			// Variable definition
			key, value := parseVariable(line)
			vars[key] = value

		} else {
			// Service definition
			name, service := parseService(line)
			if service != nil {
				serviceDirectory[name] = service
			}
		}
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return Config{
		Vars:             vars,
		ServiceDirectory: serviceDirectory,
	}
}

func parseVariable(line string) (string, string) {
	content := strings.TrimSpace(line[1:])
	key, value, _ := strings.Cut(content, "=")
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	return key, value
}

func parseService(line string) (string, tcpmux.Service) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", nil
	}

	serviceName := strings.ToUpper(fields[0])
	serviceType := strings.ToUpper(fields[1])
	fields = fields[2:]

	if serviceType == "NET" {
		return serviceName, parseNetService(fields)
	} else if serviceType == "LOCAL" {
		return serviceName, parseLocalService(fields)
	} else {
		return serviceName, nil
	}
}

func parseNetService(fields []string) tcpmux.Service {
	if len(fields) < 2 {
		return nil
	}

	ip := net.ParseIP(fields[0])
	port, err := strconv.Atoi(fields[1])
	if ip == nil {
		return nil
	} else if err != nil {
		log.Println(port)
		return nil
	}

	return tcpmux.NetService{Addr: net.TCPAddr{
		IP:   ip,
		Port: port,
	}}
}

func parseLocalService(fields []string) tcpmux.Service {
	if len(fields) == 0 {
		return nil
	}

	return tcpmux.LocalService{
		Path: fields[0],
		Args: fields[1:],
	}
}
