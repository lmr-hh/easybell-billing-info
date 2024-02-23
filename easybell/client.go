package easybell

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Client is the main type allowing you to interact with the easyBell service.
// A client maintains an authenticated connection via cookies.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new easyBell client.
// Before the client can be used you must call [Client.Login] to authenticate the client.
func NewClient() *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &Client{
		httpClient: &http.Client{Jar: jar},
	}
}

// Login authenticates the client against easyBell.
func (c *Client) Login(username string, password string) error {
	resp, err := c.httpClient.PostForm("https://login.easybell.de/login", url.Values{
		"id":       []string{username},
		"password": []string{password},
	})
	if err != nil {
		return err
	}
	// TODO: Handle invalid credentials.
	return resp.Body.Close()
}

// Logout un-authenticates the client.
// After this method returns you must re-authenticate the client before it can be used again.
func (c *Client) Logout() error {
	resp, err := c.httpClient.Get("https://login.easybell.de/logout")
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
