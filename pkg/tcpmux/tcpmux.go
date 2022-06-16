package tcpmux

import (
	"bytes"
	"errors"
	"github.com/hannesbraun/tcpmuxa/pkg/sysutils"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"
)

type Service interface {
	bridge(clientConn *net.TCPConn, initData []byte)
}

type NetService struct {
	Addr net.TCPAddr
}

type LocalService struct {
	Path string
	Args []string
}

func (s NetService) bridge(clientConn *net.TCPConn, initData []byte) {
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

	found(clientConn)

	_, err = serviceConn.Write(initData)
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		log.Println(err)
		return
	}
	go func() {
		_, err := io.Copy(clientConn, serviceConn)
		if err != nil && !errors.Is(err, syscall.EPIPE) {
			log.Println(err)
		}
		err = clientConn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()
	_, err = io.Copy(serviceConn, clientConn)
	if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, net.ErrClosed) {
		log.Println(err)
		return
	}
}

func (s LocalService) bridge(clientConn *net.TCPConn, initData []byte) {
	cmd := exec.Command(s.Path, s.Args...)
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

	sysutils.PrepareCmd(cmd)
	err = cmd.Start()
	if err != nil {
		notFound(clientConn)
		log.Println(err)
		return
	}
	defer func() {
		sysutils.Kill(cmd)

		_, err = cmd.Process.Wait()
		if err != nil {
			log.Println(err)
		}
	}()

	found(clientConn)

	_, err = stdin.Write(initData)
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		log.Println(err)
		return
	}
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

func handleConnection(conn *net.TCPConn, serviceDirectory map[string]Service) {
	defer func() {
		err := conn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Println(err)
		}
	}()

	serviceName, initData := recvServiceDesc(conn)
	service, serviceFound := serviceDirectory[serviceName]
	if serviceName == "HELP" {
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
		service.bridge(conn, initData)
	} else {
		notFound(conn)
	}
}

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

func TcpMux(serviceDirectory map[string]Service) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 1})
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
