package jiffybox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const ENV_API_TOKEN = "JIFFYBOX_API_KEY"

type Client struct {
	ApiKey string
}

func abortWith(message string) {
	fmt.Println("ERROR: " + message)
	os.Exit(1)
}

func NewFromEnv() *Client {
	return &Client{ApiKey: os.Getenv(ENV_API_TOKEN)}
}

func (client *Client) Validate() error {
	if client.ApiKey == "" {
		return fmt.Errorf("ApiKey must be set")
	}
	return nil
}

func (client *Client) BaseUrl() string {
	return "https://api.jiffybox.de/" + client.ApiKey + "/v1.0"
}

type Response struct {
	Messages []string `json:"messages"`
	Result   bool     `json:"result"`
}

type HttpResponse struct {
	StatusCode int
	Content    []byte
	Response   *http.Response
}

func (client *Client) unmarshalResponse(rsp *http.Response, i interface{}) error {
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)

	if e != nil {
		return e
	}
	return client.unmarshal(b, i)
}

func (client *Client) unmarshal(b []byte, i interface{}) error {
	er := &ErrorResponse{}
	e := json.Unmarshal(b, er)
	if e == nil {
		allErrors := []string{}
		for _, message := range er.Messages {
			allErrors = append(allErrors, message.Message)
		}
		if len(allErrors) > 0 {
			return fmt.Errorf(strings.Join(allErrors, ", "))
		}
	}
	return json.Unmarshal(b, i)
}

func (client *Client) PostForm(action string, values url.Values) (rsp *HttpResponse, e error) {
	u := client.BaseUrl() + "/" + action
	logger.Print("sending request " + u)
	httpResponse, e := http.PostForm(u, values)
	if e != nil {
		return nil, e
	}
	dbg.Printf("got status %s", httpResponse.Status)
	rsp = &HttpResponse{
		StatusCode: httpResponse.StatusCode,
	}
	rsp.Content, e = ioutil.ReadAll(httpResponse.Body)
	rsp.Response = httpResponse
	if e != nil {
		return nil, e
	}
	return rsp, e
}

var logger = log.New(os.Stderr, "", 0)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

type ErrorResponse struct {
	Messages []*Message `json:"messages"`
	Result   bool       `json:"result"`
}

func (client *Client) LoadResource(action string, i interface{}) error {
	u := client.BaseUrl() + "/" + action
	dbg.Print("loading " + u)
	rsp, e := http.Get(u)
	if e != nil {
		return e
	}
	return client.unmarshalResponse(rsp, i)
}
