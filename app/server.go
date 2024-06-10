package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func parse_request(input string) Request {
	req_parts := strings.Split(string(input), "\r\n")
	request_line := req_parts[0]
	fmt.Println(request_line)
	req_line_parts := strings.Split(request_line, " ")

	body_start := -1
	headers := make(map[string]string)
	for i := 1; i < len(req_parts); i++ {
		line := req_parts[i]
		if line == "" {
			body_start = i + 1
			break
		}
		header_parts := strings.Split(line, ": ")
		headers[header_parts[0]] = header_parts[1]
	}

	body := ""
	for j := body_start; j < len(req_parts); j++ {
		body += req_parts[j]
	}
	fmt.Printf("Request Body: %s", body)

	return Request{
		Method:  req_line_parts[0],
		Path:    req_line_parts[1],
		Headers: headers,
		Body:    body,
	}
}

type Response struct {
	Code    int
	Status  string
	Headers map[string]interface{}
	Body    string
}

func (resp Response) get_response_string() string {
	str := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.Code, resp.Status)
	for k, v := range resp.Headers {
		str += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	str += "\r\n"
	str += resp.Body

	return str
}

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

	r := parse_request(string(buf))

	// Handle request
	if strings.HasPrefix(r.Path, "/echo/") {
		str, _ := strings.CutPrefix(r.Path, "/echo/")
		echo_handler(conn, str)
	} else if r.Path == "/user-agent" {
		user_agent_handler(conn, r)
	} else if r.Path == "/" {
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

func user_agent_handler(conn net.Conn, r Request) {
	user_agent := r.Headers["User-Agent"]
	resp := Response{
		Code:   200,
		Status: "OK",
		Headers: map[string]interface{}{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(user_agent)),
		},
		Body: user_agent,
	}

	_, err := conn.Write([]byte(resp.get_response_string()))
	if err != nil {
		fmt.Printf("Error while writing response: %s\n", err.Error())
		return
	}
}
