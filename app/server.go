package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"slices"
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

var valid_encodings []string = []string{"gzip"}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	file_dir := flag.String("directory", "/tmp/", "The directory to read files from")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		func(conn net.Conn) {
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Printf("Error while reading request: %s\n", err.Error())
			}

			r := parse_request(string(buf))

			// Handle request
			if strings.HasPrefix(r.Path, "/echo/") {
				echo_handler(conn, r)
			} else if strings.HasPrefix(r.Path, "/files/") {
				if r.Method == "GET" {
					read_file_handler(conn, r, *file_dir)
				} else if r.Method == "POST" {
					upload_file_handler(conn, r, *file_dir)
				}
			} else if r.Path == "/user-agent" {
				user_agent_handler(conn, r)
			} else if r.Path == "/" {
				resp := "HTTP/1.1 200 OK\r\n\r\n"
				_, err = conn.Write([]byte(resp))
				if err != nil {
					fmt.Printf("Error while writing response: %s\n", err.Error())
				}
			} else {
				resp := "HTTP/1.1 404 Not Found\r\n\r\n"
				_, err = conn.Write([]byte(resp))
				if err != nil {
					fmt.Printf("Error while writing response: %s\n", err.Error())
				}
			}
			conn.Close()
		}(conn)
	}
}

func echo_handler(conn net.Conn, r Request) {
	path_var, _ := strings.CutPrefix(r.Path, "/echo/")

	headers := map[string]interface{}{
		"Content-Type":   "text/plain",
		"Content-Length": fmt.Sprintf("%d", len(path_var)),
	}

	encType := r.Headers["Accept-Encoding"]
	if slices.Contains(valid_encodings, encType) {
		headers["Content-Encoding"] = encType
	}

	resp := Response{
		Code:    200,
		Status:  "OK",
		Headers: headers,
		Body:    path_var,
	}.get_response_string()
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

func read_file_handler(conn net.Conn, r Request, dir string) {
	resp := ""
	file_name := strings.TrimPrefix(r.Path, "/files/")
	file_bytes, err := os.ReadFile(dir + file_name)
	if err != nil {
		fmt.Println(dir + file_name)
		fmt.Println(err.Error())
		resp = Response{
			Code:    404,
			Status:  "Not Found",
			Headers: make(map[string]interface{}),
			Body:    "",
		}.get_response_string()
	} else {
		contents := string(file_bytes)
		resp = Response{
			Code:   200,
			Status: "OK",
			Body:   contents,
			Headers: map[string]interface{}{
				"Content-Type":   "application/octet-stream",
				"Content-Length": fmt.Sprintf("%d", len(contents)),
			},
		}.get_response_string()
	}

	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Printf("Error while writing response: %s\n", err.Error())
		return
	}
}

func upload_file_handler(conn net.Conn, r Request, dir string) {
	file_name := strings.TrimPrefix(r.Path, "/files/")
	file, err := os.Create(dir + file_name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("197\t%s", r.Body)
	_, err = file.WriteString(strings.ReplaceAll(r.Body, "\x00", ""))
	file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	resp := Response{
		Status: "Created",
		Code:   201,
	}
	_, err = conn.Write([]byte(resp.get_response_string()))
	if err != nil {
		fmt.Printf("Error while writing response: %s\n", err.Error())
		return
	}
}
