// This package contains everything required to work with mediawiki servers
package mediawiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/golang/glog"
)

// An api client. Most commands require that the client has been logged in with Client.Login()
type Client struct {
	Host       string
	Username   string
	Password   string
	httpClient *http.Client
	loggedIn   bool
}

func (c *Client) initHttpClient() {
	if c.httpClient == nil {
		// Setup an http client that manages cookies
		cookieJar, _ := cookiejar.New(nil)
		c.httpClient = &http.Client{
			Jar: cookieJar,
		}
	}
}

// Ensure the client instance is properly logged in
func (c *Client) Login() error {
	if !c.loggedIn {
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
		loginUrl := fmt.Sprintf("http://%s/api.php?action=login&lgname=%s&lgpassword=%s&format=json", c.Host, c.Username, c.Password)
		res, err := c.httpClient.PostForm(loginUrl, values)
		glog.V(2).Info("Finished 1/2 HTTP calls")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		decoder := json.NewDecoder(res.Body)
		var response loginResponse
		err = decoder.Decode(&response)
		if err != nil {
			return err
		}
		if response.Login == nil {
			return errors.New("Missing login object from response")
		}
		if response.Login.Result != "NeedToken" {
			return errors.New("Did not ask for token")
		}

		// Do the same call, this time passing back the token
		values.Set("lgtoken", response.Login.Token)
		glog.V(1).Info("Making 2/2 HTTP calls")
		res, err = c.httpClient.PostForm(loginUrl, values)
		glog.V(2).Info("Finished 2/2 HTTP calls")
		if err != nil {
			return err
		}
		defer res.Body.Close()

		c.loggedIn = true
	}
	return nil
}

// Get a list of all the articles contained in the wiki
func (c *Client) ListArticleTitles() ([]string, error) {
	if !c.loggedIn {
		return nil, errors.New("Client is not logged in")
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
	listUrl := fmt.Sprintf("http://%s/api.php?format=json&action=query&list=allpages&aplimit=max", c.Host)
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
func (c Client) GetArticle(title string) (io.ReadCloser, error) {
	articleUrl := fmt.Sprintf("http://%s/index.php?action=raw&title=%s", c.Host, url.QueryEscape(title))
	resp, err := c.httpClient.Get(articleUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
