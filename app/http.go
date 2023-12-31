package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var lineTerminator = ("\r\n")
var headerTerminator = []byte("\r\n\r\n")

type Request struct {
	Method  string
	Path    string
	Version string
	// Headers []*Header
	// a map is better suited because it allows fast access to certain headers
	Headers map[string]string
	Body    []byte
}

type Response struct {
	Version    string
	StatusCode int
	StatusText string
	// Headers    []*Header
	Headers map[string]string
	Body    []byte
}

func (r *Response) SetStatusOK() {
	r.StatusCode = 200
	r.StatusText = "OK"
}

func (r *Response) SetStatusNotFound() {
	r.StatusCode = 404
	r.StatusText = "Not Found"
}

func (r *Response) SetStatusCreated() {
	r.StatusCode = 201
	r.StatusText = "Created"
}

func (r *Response) Encode() []byte {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s %s %s%s", r.Version, strconv.Itoa(r.StatusCode), r.StatusText, lineTerminator))

	for k, v := range r.Headers {
		builder.WriteString(fmt.Sprintf("%s: %s%s", k, v, lineTerminator))
	}

	// add additional line terminator to indicate end of headers
	builder.WriteString(lineTerminator)

	// add body
	builder.Write(r.Body)

	return []byte(builder.String())
}

func (r *Response) SetStatus(status int) {
	r.StatusCode = status

}

type Header struct {
	Key   string
	Value string
}

func ParseRequest(data []byte) *Request {
	request := &Request{
		Headers: make(map[string]string),
	}

	headerBytes, offset := getHeaderBytes(data)
	headerString := string(headerBytes)
	// remove trailing header termination (\r\n\r\n)
	headerString = strings.TrimRight(headerString, string(headerTerminator))
	lines := strings.Split(headerString, lineTerminator)

	for i, line := range lines {
		if i == 0 {
			// first line contains the method, path and version
			method, path, version, err := parseFirstLine(line)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}
			request.Method = method
			request.Path = path
			request.Version = version
		} else {
			// other lines are simple headers
			key, value := parseHeader(line)
			request.Headers[key] = value
		}
	}

	request.Body = data[offset:]

	return request
}

func parseFirstLine(line string) (method, path, version string, err error) {
	split := strings.Split(line, " ")
	if len(split) != 3 {
		return "", "", "", errors.New("firstline malformed")
	}

	return split[0], split[1], split[2], nil
}

func parseHeader(line string) (key, value string) {
	keyValue := strings.Split(line, ": ")
	if len(keyValue) != 2 {
		fmt.Println("error parsing header line")
		return "", ""
	}

	return keyValue[0], keyValue[1]
}

// returns the header bytes and the index where the body starts
func getHeaderBytes(data []byte) ([]byte, int) {
	index := bytes.Index(data, headerTerminator)
	if index == -1 {
		return nil, -1
	}

	// increment the index to skip the header termination squence
	index += 4

	return data[:index], index
}

func CreateResponse(request *Request) *Response {
	response := &Response{
		Version: request.Version,
		Headers: make(map[string]string),
	}

	// handle request based on the different stages - the path inside the request determines which response is needed
	if request.Path == "/" {
		response.SetStatusOK()
	} else if strings.Contains(request.Path, "/echo/") {
		// set status codes
		response.SetStatusOK()

		// set content-type
		response.Headers["Content-Type"] = "text/plain"
		// get content
		content := strings.TrimPrefix(request.Path, "/echo/")
		// set content-length
		response.Headers["Content-Length"] = strconv.Itoa(len(content))
		// set content
		response.Body = []byte(content)
	} else if request.Path == "/user-agent" {
		response.SetStatusOK()
		response.Headers["Content-Type"] = "text/plain"
		content, ok := request.Headers["User-Agent"]
		if !ok {
			fmt.Println("header does not exist")
		}
		response.Headers["Content-Length"] = strconv.Itoa(len(content))
		response.Body = []byte(content)

	} else if request.Method == "GET" && strings.HasPrefix(request.Path, "/files/") {
		filename := filepath.Base(request.Path)
		filePath := filepath.Join(directoryArg, filename)

		if fileExists(filePath) {
			response.SetStatusOK()
			response.Headers["Content-Type"] = "application/octet-stream"
			content, _ := getFileContent(filePath)
			response.Headers["Content-Length"] = strconv.Itoa(len(content))
			response.Body = []byte(content)
		} else {
			response.SetStatusNotFound()
		}

	} else if request.Method == "POST" && strings.HasPrefix(request.Path, "/files/") {
		filename := filepath.Base(request.Path)
		filePath := filepath.Join(directoryArg, filename)
		// create file
		writeToFile(filePath, request.Body)
		response.SetStatusCreated()

	} else {
		response.SetStatusNotFound()
	}

	return response
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func getFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func writeToFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
