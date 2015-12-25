package mediawiki

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "io/ioutil"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "strings"
)

type Client struct {
    httpClient  *http.Client
    loggedIn    bool
    loginUrl    string
    listUrl     string
}

func NewClient(host, username, password string) *Client {
    // Setup an http client that manages cookies
    cookieJar, _ := cookiejar.New(nil)
    client := &http.Client{
        Jar: cookieJar,
    }

    loginUrl := fmt.Sprintf("http://%s/api.php?action=login&lgname=%s&lgpassword=%s&format=json", host, username, password)
    listUrl := fmt.Sprintf("http://%s/api.php?format=json&action=query&list=allpages&aplimit=max", host)

    return &Client{
        httpClient: client,
        loginUrl:   loginUrl,
        listUrl:    listUrl,
    }
}

func (c *Client) Login() error {
    if !c.loggedIn {
        type LoginResponseInner struct {
            result, token string
        }
        type LoginResponse struct {
            login *LoginResponseInner
        }

        // Make the first call to get a token
        values := make(url.Values)
        res, err := c.httpClient.PostForm(c.loginUrl, values)
        if err != nil {
            return err
        }
        defer res.Body.Close()
        decoder := json.NewDecoder(res.Body)
        var response LoginResponse
        err = decoder.Decode(&response)
        if err != nil {
            return err
        }
        if response.login == nil {
            return errors.New("Missing login object from response")
        }
        if response.login.result != "NeedToken" {
            return errors.New("Did not ask for token")
        }

        // Do the same call, this time passing back the token
        values.Set("lgtoken", response.login.token)
        res, err = c.httpClient.PostForm(c.loginUrl, values)
        if err != nil {
            return err
        }
        defer res.Body.Close()

        c.loggedIn = true
    }
    return nil
}

func (c *Client) ListArticles() error {
    if !c.loggedIn {
        return errors.New("Client is not logged in")
    }
    type Page struct {
        title string
    }
    type Query struct {
        allpages []Page
    }
    type Result struct {
        query Query
    }
    return nil
}

func DoThing(host, username, password, exportDir string) {
    // Setup an http client that manages cookies
    cookieJar, _ := cookiejar.New(nil)
    client := &http.Client{
        Jar: cookieJar,
    }

    login_url := fmt.Sprintf("http://%s/api.php?action=login&lgname=%s&lgpassword=%s&format=json", host, username, password)

    res, err := client.Post(login_url,"text/plain",strings.NewReader(""))
    if err != nil {
        log.Fatal(err)
    }
    robots, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\n", robots)
}

func login(client http.Client, urlString string, token string) ([]byte, error) {
    values := make(url.Values)
    if token != "" {
        values.Set("lgtoken", token)
    }
    res, err := client.PostForm(urlString, values)
    if err != nil {
        return nil, err
    }
    robots, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err != nil {
        log.Fatal(err)
    }
    return robots, nil
}
