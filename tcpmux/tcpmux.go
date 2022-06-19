package tcpmux

import (
	"bytes"
	"errors"
	"github.com/hannesbraun/tcpmuxa/sysutils"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"
)

// Service represents a service that may be bridged through TCPMUX.
type Service interface {
	bridge(clientConn *net.TCPConn, initData []byte)
}

// NetService represents a service available through a network connection.
type NetService struct {
	Addr net.TCPAddr
}

// LocalService represents a service available as an executable.
type LocalService struct {
	Path string
	Args []string
}

// bridge briges a network service.
func (s NetService) bridge(clientConn *net.TCPConn, initData []byte) {
	// Open/close connection to service
	serviceConn, err := net.DialTCP("tcp", nil, &s.Addr)
	if err != nil {
		notFound(clientConn)
		log.Println(err)
		return
	}
	defer func() {
		err := serviceConn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()

	// Service found
	found(clientConn)

	// Send bytes that were already read previously
	_, err = serviceConn.Write(initData)
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		log.Println(err)
		return
	}
	// Send data from the service to the client
	go func() {
		_, err := io.Copy(clientConn, serviceConn)
		if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
		err = clientConn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()
	// Send data from the client to the service
	_, err = io.Copy(serviceConn, clientConn)
	if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, net.ErrClosed) {
		log.Println(err)
		return
	}
}

// bridge briges a local service.
func (s LocalService) bridge(clientConn *net.TCPConn, initData []byte) {
	// Create command
	cmd := exec.Command(s.Path, s.Args...)
	// Get pipes to stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		notFound(clientConn)
		log.Println(err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		notFound(clientConn)
		log.Println(err)
		return
	}

	// Run process
	sysutils.PrepareCmd(cmd)
	err = cmd.Start()
	if err != nil {
		notFound(clientConn)
		log.Println(err)
		return
	}

	// Kill process when done
	defer func() {
		sysutils.Kill(cmd)

		_, err = cmd.Process.Wait()
		if err != nil {
			log.Println(err)
		}
	}()

	found(clientConn)

	// Write bytes that were already read previously
	_, err = stdin.Write(initData)
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		log.Println(err)
		return
	}
	// Client -> Process (stdin)
	go func() {
		_, err := io.Copy(clientConn, stdout)
		if err != nil && !errors.Is(err, syscall.EPIPE) {
			log.Println(err)
		}
		err = clientConn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()
	// Process (stdout) -> Client
	_, err = io.Copy(stdin, clientConn)
	if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, net.ErrClosed) {
		log.Println(err)
		return
	}
}

func notFound(conn *net.TCPConn) {
	_, err := conn.Write([]byte("-Service not found\r\n"))
	if err != nil {
		log.Println(err)
	}
}

func found(conn *net.TCPConn) {
	_, err := conn.Write([]byte("+\r\n"))
	if err != nil {
		log.Println(err)
	}
}

// handleConnection handles an incoming TCP connection.
func handleConnection(conn *net.TCPConn, serviceDirectory map[string]Service) {
	// Close connection before returning
	defer func() {
		err := conn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()

	// Determine requested service
	serviceName, initData := recvServiceDesc(conn)
	service, serviceFound := serviceDirectory[serviceName]

	if serviceName == "HELP" {
		// HELP service
		// Print the available service names
		_, err := conn.Write([]byte("HELP\r\n"))
		if err != nil {
			log.Println(err)
			return
		}
		for service := range serviceDirectory {
			_, err = conn.Write([]byte(service + "\r\n"))
			if err != nil {
				log.Println(err)
				return
			}
		}
	} else if serviceFound {
		// Service found: run it
		service.bridge(conn, initData)
	} else {
		// Service not found in serviceDictionary
		notFound(conn)
	}
}

// recvServiceDesc reads from a socket to receive the requested service description.
// Both the requested service name and some possibly unprocessed but read bytes will be returned.
// A service name consists of all the characters up until the first occurrence of CRLF.
func recvServiceDesc(conn *net.TCPConn) (string, []byte) {
	var buf []byte
	tmp := make([]byte, 128)

	var serviceBytes, data []byte
	crlfFound := false
	for !crlfFound {
		n, err := conn.Read(tmp)
		if err != nil && err != io.EOF {
			log.Println(err)
			return "", []byte{}
		}
		buf = append(buf, tmp[0:n]...)

		serviceBytes, data, crlfFound = bytes.Cut(buf, []byte{0x0d, 0x0a})
	}

	return strings.ToUpper(string(serviceBytes)), data
}

// TCPMUX runs the TCPMUX service on the default port.
// This function will not return once the socket is created.
func TCPMUX(serviceDirectory map[string]Service) {
	TCPMUXWithPort(serviceDirectory, 1)
}

// TCPMUXWithPort runs the TCPMUX service on the specified port.
// This function will not return once the socket is created.
func TCPMUXWithPort(serviceDirectory map[string]Service, port int) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Started TCPMUX service")

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn, serviceDirectory)
	}
}
