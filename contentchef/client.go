package contentchef

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

const (
	libraryVersion = "0.1"
	userAgent      = "contentchef-go/" + libraryVersion
	mediaType      = "application/json"
)

// TargetDateResolver is used to retrieve contents in the preview channel in a specific dare different from the current date
// a valid targetDateResolver must return a valid date expressed using the ISO format like  2019-08-16T12:22:232Z
type TargetDateResolver func() string

// Client manages the communication with the ContenChef API.
type Client struct {
	httpClient         *http.Client
	BaseURL            *url.URL
	SpaceID            string
	TargetDateResolver TargetDateResolver
}

// ClientOptions is the configuration object passed to the Client constructor
type ClientOptions struct {
	// The base URL of your Content Chef instance REQUIRED
	BaseURL string
	// Your Content Chef SpaceID REQUIRED
	SpaceID string
	// If you don't want to use the default HTTP client you can set a custom one
	Client             *http.Client
	TargetDateResolver TargetDateResolver
}

// NewClient return a new Client reference
//
// It takes a ClientOptions reference.
func NewClient(o *ClientOptions) (*Client, error) {
	BaseURL, err := url.Parse(o.BaseURL)
	if err != nil {
		return nil, err
	}
	if o.SpaceID == "" {
		return nil, errors.New("SpaceID must be setted")
	}
	httpClient := o.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	cf := &Client{
		httpClient,
		BaseURL,
		o.SpaceID,
		o.TargetDateResolver,
	}
	return cf, nil
}

func (c *Client) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", mediaType)
	}

	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

type response struct {
	*http.Response
}

func newResponse(r *http.Response) *response {
	return &response{Response: r}
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}

	req = req.WithContext(ctx)
	res, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer res.Body.Close()

	response := newResponse(res)

	err = checkResponse(res)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, res.Body)
		} else {
			decErr := json.NewDecoder(res.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return response, err
}

type errorResponse struct {
	Response *http.Response
	Message  string
	X        map[string]interface{} `json:"-"`
}

func (r *errorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}
	errorResponse := &errorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}
	return errorResponse
}

func addOptions(path string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return path, nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return path, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (c *Client) get(ctx context.Context, path, apiKey string, opts, v interface{}) error {
	path, err := addOptions(path, opts)
	if err != nil {
		return err
	}

	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-SPACE-D-API-Key", apiKey)

	_, err = c.do(ctx, req, v)

	return err
}
