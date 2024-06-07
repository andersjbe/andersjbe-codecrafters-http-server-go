package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Printf("Error while reading request: %s\n", err.Error())
	}

	// Extract path
	req_parts := strings.Split(string(buf), "\r\n")
	request_line := req_parts[0]
	fmt.Println(request_line)
	req_line_parts := strings.Split(request_line, " ")
	path := req_line_parts[1]

	// Handle request
	if strings.HasPrefix(path, "/echo/") {
		str, _ := strings.CutPrefix(path, "/echo/")
		echo_handler(conn, str)
	} else if path == "/" {
		resp := "HTTP/1.1 200 OK\r\n\r\n"
		_, err = conn.Write([]byte(resp))
		if err != nil {
			fmt.Printf("Error while writing response: %s\n", err.Error())
			return
		}
	} else {
		resp := "HTTP/1.1 404 Not Found\r\n\r\n"
		_, err = conn.Write([]byte(resp))
		if err != nil {
			fmt.Printf("Error while writing response: %s\n", err.Error())
			return
		}
	}
}

func echo_handler(conn net.Conn, path_var string) {
	resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path_var), path_var)
	_, err := conn.Write([]byte(resp))
	if err != nil {
		fmt.Printf("Error while writing response: %s\n", err.Error())
		return
	}
}
