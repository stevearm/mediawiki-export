// This package contains everything required to work with mediawiki servers
package mediawiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"github.com/golang/glog"
)

// An api client. Either call Client.Login(), or it will auto-login on first use
type Client interface {
	Login() error
	ListArticleTitles() ([]string, error)
	GetArticle(title string) (string, error)
}

func GetClient(host, username, password string) Client {
	return &client{
		host:     host,
		username: username,
		password: password,
	}
}

type client struct {
	host       string
	username   string
	password   string
	httpClient *http.Client
	loginError error
	authLock   sync.Once
}

func (c *client) initHttpClient() {
	if c.httpClient == nil {
		// Setup an http client that manages cookies
		cookieJar, _ := cookiejar.New(nil)
		c.httpClient = &http.Client{
			Jar: cookieJar,
		}
	}
}

func (c *client) login() {
	c.initHttpClient()
	glog.Info("Logging in")
	type loginResponseInner struct {
		Result string `json:"result"`
		Token  string `json:"token"`
	}
	type loginResponse struct {
		Login *loginResponseInner `json:"login"`
	}

	// Make the first call to get a token
	glog.V(1).Info("Making 1/2 HTTP calls")
	values := make(url.Values)
	loginUrl := fmt.Sprintf("http://%s/api.php?action=login&lgname=%s&lgpassword=%s&format=json", c.host, c.username, c.password)
	res, err := c.httpClient.PostForm(loginUrl, values)
	glog.V(2).Info("Finished 1/2 HTTP calls")
	if err != nil {
		c.loginError = err
		return
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var response loginResponse
	c.loginError = decoder.Decode(&response)
	if c.loginError != nil {
		return
	}
	if response.Login == nil {
		c.loginError = errors.New("Missing login object from response")
		return
	}
	if response.Login.Result != "NeedToken" {
		c.loginError = errors.New("Did not ask for token")
		return
	}

	// Do the same call, this time passing back the token
	values.Set("lgtoken", response.Login.Token)
	glog.V(1).Info("Making 2/2 HTTP calls")
	res, c.loginError = c.httpClient.PostForm(loginUrl, values)
	glog.V(2).Info("Finished 2/2 HTTP calls")
	if c.loginError != nil {
		return
	}
	defer res.Body.Close()
}

// Ensure the client instance is properly logged in
func (c *client) Login() error {
	c.authLock.Do(c.login)
	return c.loginError
}

// Get a list of all the articles contained in the wiki
func (c *client) ListArticleTitles() ([]string, error) {
	c.authLock.Do(c.login)
	if c.loginError != nil {
		return nil, c.loginError
	}
	glog.Info("Listing all articles")
	type page struct {
		Title string `json:"title"`
	}
	type query struct {
		AllPages []page `json:"allpages"`
	}
	type result struct {
		Query query `json:"query"`
	}
	listUrl := fmt.Sprintf("http://%s/api.php?format=json&action=query&list=allpages&aplimit=max", c.host)
	res, err := c.httpClient.Get(listUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var response result
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	titles := make([]string, len(response.Query.AllPages))
	for i := 0; i < len(titles); i++ {
		titles[i] = response.Query.AllPages[i].Title
	}
	return titles, nil
}

// Get the raw wikitext of an article
func (c client) GetArticle(title string) (string, error) {
	c.authLock.Do(c.login)
	if c.loginError != nil {
		return "", c.loginError
	}
	articleUrl := fmt.Sprintf("http://%s/index.php?action=raw&title=%s", c.host, url.QueryEscape(title))
	resp, err := c.httpClient.Get(articleUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}
