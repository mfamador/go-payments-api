package test

import (
	"encoding/json"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"io/ioutil"
	"log"
	"strings"
)

type Client struct {
	ServerUrl string
	http      *httpclient.HttpClient
	Resp      *httpclient.Response
	Json      map[string]interface{}
	Text      string
	Err       error
}

func NewClient(serverUrl string) *Client {
	return &Client{
		http:      httpclient.NewHttpClient(),
		ServerUrl: serverUrl,
	}
}

func (c *Client) UrlFor(path string) string {
	return fmt.Sprintf("%s%s", c.ServerUrl, path)
}

func (c *Client) Get(path string) {
	url := c.UrlFor(path)
	res, err := c.http.Get(url)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

func (c *Client) Delete(path string) {
	url := c.UrlFor(path)
	res, err := c.http.Delete(url)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

func (c *Client) Post(path string, data string) {
	url := c.UrlFor(path)
	res, err := c.http.PostJson(url, data)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

func (c *Client) Put(path string, data string) {
	url := c.UrlFor(path)
	res, err := c.http.PutJson(url, data)
	c.Resp = res
	c.Err = err
	c.parseResponse()
}

func (c *Client) parseResponse() {
	if c.Err != nil {
		return
	}

	if c.Resp != nil && c.Resp.Body != nil {
		bytes, err := ioutil.ReadAll(c.Resp.Body)
		if err != nil {
			log.Printf("Could not read bytes from http response")
		} else {
			if c.HasJson() {
				c.maybeParseJson(bytes)
			} else if c.HasText() {
				c.maybeParseText(bytes)
			}
		}
	}
}

func (c *Client) maybeParseJson(bytes []byte) {
	var anyJson map[string]interface{}
	err := json.Unmarshal(bytes, &anyJson)
	if err != nil {
		log.Printf("Could not unmarshall json: %v", string(bytes))
	} else {
		c.Json = anyJson
	}
}

func (c *Client) maybeParseText(bytes []byte) {
	c.Text = string(bytes)
}

func (c *Client) HasJson() bool {
	return c.Resp != nil && strings.Contains(c.Resp.Header.Get("content-type"), "application/json")
}

func (c *Client) HasText() bool {
	return c.Resp != nil && strings.Contains(c.Resp.Header.Get("content-type"), "text/plain")
}
