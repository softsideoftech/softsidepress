package forwardEmail

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"
)

var localAddress string = ":587"
var backendAddress string = ":25"
var certificatePath string = os.Getenv("SSL_CERT")
var keyPath string = os.Getenv("SSL_KEY")

func StartSsl() {
	if (certificatePath == "" || keyPath == "") {
		log.Println("Missing SSL_CERT or SSL_KEY environment varibles. Will continue without SSL.")
		return
	}
	
	cert, err := tls.LoadX509KeyPair(certificatePath, keyPath)
	if err != nil {
		log.Fatalf("error in tls.LoadX509KeyPair: %s", err)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	listener, err := tls.Listen("tcp", localAddress, &config)
	if err != nil {
		log.Fatalf("error in tls.Listen: %s", err)
	}

	log.Printf("local server on: %s, backend server on: %s", localAddress, backendAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error in listener.Accept: %s", err)
			break
		}

		go handle(conn)
	}
}

func handle(clientConn net.Conn) {
	tlsconn, ok := clientConn.(*tls.Conn)
	if ok {

		err := tlsconn.Handshake()
		if err != nil {
			log.Printf("error in tls.Handshake: %s", err)
			clientConn.Close()
			return
		}

		backendConn, err := net.Dial("tcp", backendAddress)
		if err != nil {
			log.Printf("error in net.Dial: %s", err)
			clientConn.Close()
			return
		}

		go Tunnel(clientConn, backendConn)
		go Tunnel(backendConn, clientConn)
	}
}

func Tunnel(from, to io.ReadWriteCloser) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered while tunneling")
		}
	}()

	io.Copy(from, to)
	to.Close()
	from.Close()
	log.Printf("tunneling is done")
}