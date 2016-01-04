// Package httpmock provides a way to mock http calls by creating a server that will respond in the intended way, then
// returning an http.Client instance that will redirect all calls to the mocked server.
//
// Inspired by http://keighl.com/post/mocking-http-responses-in-golang/
package httpmock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Method int

const (
	GetMethod Method = iota
	PostMethod
	PutMethod
)

// Used to capture http requests that the server receives
type Request struct {
	Method Method
	Url    string
	Body   string
}

// Pass one of these to Server.QueueResponse() to setup the server for
// responding to http requests
type Response struct {
	ResponseCode int
	ContentType  string
	Content      string
}

// Generic error response usually used as the default response when
// creating a mock server
func ErrorResponse() Response {
	return Response{
		ResponseCode: 500,
	}
}

// Server houses a stub internal server that expects certain calls and will
// log any unexpected ones. Call Init to setup the server and get a client
type Server struct {
	server          *httptest.Server
	requests        chan Request
	responses       chan Response
	defaultResponse Response
}

// Init sets up the server with an empty queue of responses and a default response,
// and returns an http.Client instance that will redirect all traffic to this server
func (s *Server) Init(defaultResponse Response) *http.Client {
	s.Close()
	if s.responses != nil {
		close(s.responses)
		s.responses = nil
	}
	s.defaultResponse = defaultResponse
	const QUEUE_LEN int = 100
	s.responses = make(chan Response, QUEUE_LEN)
	s.requests = make(chan Request, QUEUE_LEN)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		req := Request{
			Url: request.URL.String(),
		}
		switch request.Method {
		case "GET":
			req.Method = GetMethod
		case "POST":
			req.Method = PostMethod
		case "PUT":
			req.Method = PutMethod
		default:
			fmt.Printf("Unknown method: \"%v\"", request.Method)
		}
		switch req.Method {
		case PostMethod:
			fallthrough
		case PutMethod:
			body, err := ioutil.ReadAll(request.Body)
			if err != nil {
				fmt.Printf("Failed to read request body: %v", err)
			}
			req.Body = string(body)
		default:
			break
		}
		s.requests <- req

		response := s.defaultResponse
		select {
		case response = <-s.responses:
			break
		default:
			break
		}
		w.WriteHeader(response.ResponseCode)
		w.Header().Set("Content-Type", response.ContentType)
		fmt.Fprint(w, response.Content)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(s.server.URL)
		},
	}
	return &http.Client{Transport: transport}
}

// Add a response to the queue that should be returned
// on a request
func (s Server) QueueResponse(response Response) {
	s.responses <- response
}

// Get a channel of the all requests that the server received
func (s Server) Requests() <-chan Request {
	return s.requests
}

// Close shuts down any running servers. Call this with defer right after
// calling Init
func (s *Server) Close() {
	if s.server != nil {
		s.server.Close()
		s.server = nil
	}
}
