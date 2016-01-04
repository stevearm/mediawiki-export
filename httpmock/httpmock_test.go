package httpmock

import (
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
)

func TestServerDefaultErrorResponse(t *testing.T) {
	server := &Server{}
	client := server.Init(ErrorResponse())
	defer server.Close()

	resp, err := client.Get("http://example.org/path")
	defer resp.Body.Close()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Wrong status: %v", resp.StatusCode)
	}
}

func TestServerDefaultCustomResponse(t *testing.T) {
	defaultResponse := Response{
		ResponseCode: 200,
		ContentType:  "text/custom",
		Content:      "Test content",
	}

	server := &Server{}
	client := server.Init(defaultResponse)
	defer server.Close()

	resp, err := client.Get("http://example.org/path")
	defer resp.Body.Close()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Wrong status: %v", resp.StatusCode)
	}
	contentType := resp.Header["Content-Type"]
	if len(contentType) != 1 || strings.Contains(contentType[0], "text/custom") {
		t.Errorf("Wrong content type: %v", contentType)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	bodyString := string(body)
	if bodyString != "Test content" {
		t.Errorf("Wrong content: <%v>", bodyString)
	}
}

func TestServerExpectedRequest(t *testing.T) {
	server := &Server{}
	client := server.Init(ErrorResponse())
	defer server.Close()

	preparedResponse := Response{
		ResponseCode: 200,
		ContentType:  "text/custom",
		Content:      "Test content",
	}
	server.QueueResponse(preparedResponse)

	resp, err := client.Get("http://example.org/path")
	defer resp.Body.Close()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Wrong status: %v", resp.StatusCode)
	}
	contentType := resp.Header["Content-Type"]
	if len(contentType) != 1 || strings.Contains(contentType[0], "text/custom") {
		t.Errorf("Wrong content type: %v", contentType)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	bodyString := string(body)
	if bodyString != "Test content" {
		t.Errorf("Wrong content: <%v>", bodyString)
	}

	resp, err = client.Get("http://example.org/path")
	defer resp.Body.Close()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Wrong status: %v", resp.StatusCode)
	}
}

func TestServerGetRequest(t *testing.T) {
	server := &Server{}
	client := server.Init(ErrorResponse())
	defer server.Close()

	resp, err := client.Get("http://example.org/path")
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	requests := server.Requests()
	request := <-requests
	if request.Method != GetMethod {
		t.Errorf("Wrong method: %v", request.Method)
	}
	if request.Url != "http://example.org/path" {
		t.Errorf("Wrong url: %v", request.Url)
	}
	if request.Body != "" {
		t.Errorf("Wrong body: <%v>", request.Body)
	}
	if len(requests) != 0 {
		t.Errorf("Request channel should be empty: %v", len(requests))
	}
}

func TestServerPostRequest(t *testing.T) {
	server := &Server{}
	client := server.Init(ErrorResponse())
	defer server.Close()

	resp, err := client.PostForm("http://example.org/path",
		url.Values{"key": {"Value"}, "id": {"123"}})
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	requests := server.Requests()

	request := <-requests
	if request.Method != PostMethod {
		t.Errorf("Wrong method: %v", request.Method)
	}
	if request.Url != "http://example.org/path" {
		t.Errorf("Wrong url: %v", request.Url)
	}
	if request.Body != "key=Value&id=123" && request.Body != "id=123&key=Value" {
		t.Errorf("Wrong body: <%v>", request.Body)
	}
	if len(requests) != 0 {
		t.Errorf("Request channel should be empty: %v", len(requests))
	}
}
