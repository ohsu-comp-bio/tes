package tes

import (
	"io/ioutil"
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/net/context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// NewClient returns a new HTTP client for accessing
// Create/List/Get/Cancel Task endpoints. "address" is the address
// of the TES server.
func NewClient(address string) (*Client, error) {
	re := regexp.MustCompile("^(.+://)?(.[^/]+)(.+)?$")
	endpoint := re.ReplaceAllString(address, "$1$2")

	reScheme := regexp.MustCompile("^.+://")
	if reScheme.MatchString(endpoint) {
		if !strings.HasPrefix(endpoint, "http") {
			return nil, fmt.Errorf("invalid protocol: '%s'; expected: 'http://' or 'https://'", reScheme.FindString(endpoint))
		}
	} else {
		endpoint = "http://" + endpoint
	}

	return &Client{
		address: endpoint,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Client represents the HTTP Task client.
type Client struct {
	address   string
	client    *http.Client
}

// GetTask returns the raw bytes from GET /v1/tasks/{id}
func (c *Client) GetTask(ctx context.Context, req *GetTaskRequest) (*Task, error) {
	// Send request
	u := c.address + "/v1/tasks/" + req.Id + "?view=" + req.View.String()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	body, err := checkResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &Task{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListTasks returns the result of GET /v1/tasks
func (c *Client) ListTasks(ctx context.Context, req *ListTasksRequest) (*ListTasksResponse, error) {
	// Build url query parameters
	v := url.Values{}
	addString(v, "name_prefix", req.GetNamePrefix())
	addUInt32(v, "page_size", req.GetPageSize())
	addString(v, "page_token", req.GetPageToken())
	addString(v, "view", req.GetView().String())

	// Send request
	u := c.address + "/v1/tasks?" + v.Encode()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	body, err := checkResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &ListTasksResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateTask POSTs a Task message to /v1/tasks
func (c *Client) CreateTask(ctx context.Context, task *Task) (*CreateTaskResponse, error) {
	verr := Validate(task)
	if verr != nil {
		return nil, fmt.Errorf("invalid task message: %v", verr)
	}

	var b bytes.Buffer
	err := Marshaler.Marshal(&b, task)
	if err != nil {
		return nil, fmt.Errorf("error marshaling task message: %v", err)
	}

	// Send request
	u := c.address + "/v1/tasks"
	hreq, _ := http.NewRequest("POST", u, &b)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	body, err := checkResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &CreateTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CancelTask POSTs to /v1/tasks/{id}:cancel
func (c *Client) CancelTask(ctx context.Context, req *CancelTaskRequest) (*CancelTaskResponse, error) {
	u := c.address + "/v1/tasks/" + req.Id + ":cancel"
	hreq, _ := http.NewRequest("POST", u, nil)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	body, err := checkResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &CancelTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetServiceInfo returns result of GET /v1/tasks/service-info
func (c *Client) GetServiceInfo(ctx context.Context, req *ServiceInfoRequest) (*ServiceInfo, error) {
	u := c.address + "/v1/tasks/service-info"
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	body, err := checkResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &ServiceInfo{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// WaitForTask polls /v1/tasks/{id} for each Id provided and returns
// once all tasks are in a terminal state.
func (c *Client) WaitForTask(ctx context.Context, taskIDs ...string) error {
	for range time.NewTicker(time.Second * 2).C {
		done := false
		for _, id := range taskIDs {
			r, err := c.GetTask(ctx, &GetTaskRequest{
				Id:   id,
				View: TaskView_MINIMAL,
			})
			if err != nil {
				return err
			}
			switch r.State {
      case Complete:
				done = true
      case ExecutorError, SystemError, Canceled:
				errMsg := fmt.Sprintf("Task %s exited with state %s", id, r.State.String())
				return errors.New(errMsg)
			default:
				done = false
			}
		}
		if done {
			return nil
		}
	}
	return nil
}


// checkResponse does some basic error handling
// and reads the response body into a byte array
func checkResponse(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if (resp.StatusCode / 100) != 2 {
		return nil, fmt.Errorf("[STATUS CODE - %d]\t%s", resp.StatusCode, body)
	}
	return body, nil
}

func addString(u url.Values, key, value string) {
	if value != "" {
		u.Add(key, value)
	}
}
func addUInt32(u url.Values, key string, value uint32) {
	if value != 0 {
		u.Add(key, fmt.Sprint(value))
	}
}
