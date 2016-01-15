package mediawiki

import (
	"net/http"
	"testing"

	"github.com/stevearm/mediawiki-export/httpmock"
)

func setup() (Client, *httpmock.Server) {
	server := &httpmock.Server{}
	httpClient := server.Init(httpmock.ErrorResponse())

	client := &client{
		host:     "wiki.example.org",
		username: "myuser",
		password: "mypass",
	}
	client.initHttpClient()
	client.httpClient = &http.Client{
		Jar:       client.httpClient.Jar,
		Transport: httpClient.Transport,
	}
	return client, server
}

func setupLoginResponses(server *httpmock.Server) {
	server.QueueResponse(httpmock.Response{
		ResponseCode: 200,
		ContentType:  "application/json",
		Content:      `{"login":{"result":"NeedToken","token":"tokenvalue1234abcd"}}`,
	})
	server.QueueResponse(httpmock.Response{
		ResponseCode: 200,
		ContentType:  "application/json",
		Content:      `{"login":{"result":"Done","token":"tokenvalue1234abcd"}}`,
	})
}

func checkLoginCalls(t *testing.T, requests <-chan httpmock.Request) {
	request := <-requests
	if request.Method != httpmock.PostMethod || request.Url != "http://wiki.example.org/api.php?action=login&lgname=myuser&lgpassword=mypass&format=json" || request.Body != "" {
		t.Errorf("Bad first call: %v", request)
	}
	request = <-requests
	if request.Method != httpmock.PostMethod || request.Url != "http://wiki.example.org/api.php?action=login&lgname=myuser&lgpassword=mypass&format=json" || request.Body != "lgtoken=tokenvalue1234abcd" {
		t.Errorf("Bad first call: %v", request)
	}
}

func TestLogin(t *testing.T) {
	client, server := setup()
	defer server.Close()
	setupLoginResponses(server)
	err := client.Login()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	err = client.Login()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	requests := server.Requests()
	checkLoginCalls(t, requests)
	if len(requests) != 0 {
		t.Errorf("Found extra requests: %v", len(requests))
	}
}

func TestLoginFailure(t *testing.T) {
	client, server := setup()
	defer server.Close()
	server.QueueResponse(httpmock.Response{
		ResponseCode: 200,
		ContentType:  "application/json",
		Content:      `[{"email":"bob@example.com","status":"sent","reject_reason":"hard-bounce","_id":"1"}]`,
	})
	err := client.Login()
	if err == nil {
		t.Errorf("Should have failed during login")
	}
	err = client.Login()
	if err == nil {
		t.Errorf("Should have failed during login")
	}
}

func TestListWhenUnauthenticated(t *testing.T) {
	client, server := setup()
	defer server.Close()
	_, err := client.ListArticleTitles()
	if err == nil {
		t.Errorf("Should have failed as not authenticated")
	}
}

func TestListTitles(t *testing.T) {
	client, server := setup()
	defer server.Close()
	setupLoginResponses(server)
	server.QueueResponse(httpmock.Response{
		ResponseCode: 200,
		ContentType:  "application/json",
		Content:      `{"query":{"allpages":[{"title":"First article"},{"title":"Second article"}]}}`,
	})
	titles, err := client.ListArticleTitles()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(titles) != 2 || titles[0] != "First article" || titles[1] != "Second article" {
		t.Errorf("Wrong titles: %v", titles)
	}

	requests := server.Requests()
	checkLoginCalls(t, requests)
	request := <-requests
	if request.Method != httpmock.GetMethod || request.Url != "http://wiki.example.org/api.php?format=json&action=query&list=allpages&aplimit=max" {
		t.Errorf("Bad call: %v", request)
	}
	if len(requests) != 0 {
		t.Errorf("Found extra requests: %v", len(requests))
	}

}

func TestDownloadArticle(t *testing.T) {
	client, server := setup()
	defer server.Close()
	setupLoginResponses(server)
	server.QueueResponse(httpmock.Response{
		ResponseCode: 200,
		ContentType:  "application/wikitext",
		Content: `This wiki page contains things
* [[Work]]
* [[Finances]]
* [[Programing]]
`,
	})
	article, err := client.GetArticle("Home Page")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if article != `This wiki page contains things
* [[Work]]
* [[Finances]]
* [[Programing]]
` {
		t.Errorf("Incorrect article body: <%v>", article)
	}

	requests := server.Requests()
	checkLoginCalls(t, requests)
	request := <-requests
	if request.Method != httpmock.GetMethod || request.Url != "http://wiki.example.org/index.php?action=raw&title=Home+Page" {
		t.Errorf("Bad call: %v", request)
	}
	if len(requests) != 0 {
		t.Errorf("Found extra requests: %v", len(requests))
	}

}
