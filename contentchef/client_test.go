package contentchef

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	mux *http.ServeMux

	ctx = context.TODO()

	client *Client

	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	myDate := time.Now()
	opt := &ClientOptions{
		BaseURL:    server.URL + "/",
		SpaceID:    "my_space",
		TargetDate: myDate,
	}
	client, _ = NewClient(opt)
}

func teardown() {
	server.Close()
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.newRequest(http.MethodGet, "/", nil)
	body := new(foo)
	_, err := client.do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Response body = %v, expected %v", body, expected)
	}
}

func TestDo_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.newRequest(http.MethodGet, "/", nil)
	_, err := client.do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

// Test handling of an error caused by the internal http client's Do()
// function.
func TestDo_redirectLoop(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})

	req, _ := client.newRequest(http.MethodGet, "/", nil)
	_, err := client.do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func TestCheckResponse(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`{"message":"m",
			"errors": [{"resource": "r", "field": "f", "code": "c"}]}`)),
	}
	err := checkResponse(res).(*errorResponse)

	if err == nil {
		t.Fatalf("Expected error response.")
	}

	expected := &errorResponse{
		Response: res,
		Message:  "m",
	}
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("Error = %#v, expected %#v", err, expected)
	}
}

// ensure that we properly handle API errors that do not contain a response
// body
func TestCheckResponse_noBody(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	err := checkResponse(res).(*errorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	expected := &errorResponse{
		Response: res,
	}
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("Error = %#v, expected %#v", err, expected)
	}
}

// ensure that we properly handle API errors that do not contain a response
// body
func TestCheckResponse_errOnBodyClose(t *testing.T) {
	type failToClose = struct{}

	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	err := checkResponse(res).(*errorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	expected := &errorResponse{
		Response: res,
	}
	if !reflect.DeepEqual(err, expected) {
		t.Errorf("Error = %#v, expected %#v", err, expected)
	}
}

func Test_errorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := errorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty errorResponse.Error()")
	}
}

func Test_addOptions(t *testing.T) {
	type myOpts struct {
		Foo string `url:"foo"`
	}
	type args struct {
		path string
		opts interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "add options",
			want: "/something?foo=bar",
			args: args{
				path: "/something",
				opts: myOpts{Foo: "bar"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addOptions(tt.args.path, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("addOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("addOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_get(t *testing.T) {
	setup()
	defer teardown()

	type myType struct {
		PublicID   string      `json:"publicId"`
		Unknown    interface{} `json:"unknown"`
		OnlineDate time.Time   `json:"onlineDate"`
	}

	responseBlob := `{
		"publicId": "MyPublicId",
		"unknown": {
			"title": "MySite title"
		},
		"onlineDate": "2020-04-09T22:00:00.000Z"
	}`

	myDate, _ := time.Parse(time.RFC3339, "2020-04-09T22:00:00.000Z")
	want := &myType{
		PublicID: "MyPublicId",
		Unknown: map[string]interface{}{
			"title": "MySite title",
		},
		OnlineDate: myDate,
	}

	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, responseBlob)
	})

	got := &myType{}
	err := client.get(context.TODO(), "foo", "super_secret", nil, got)

	if err != nil {
		t.Errorf("client.get() returned error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("client.get() = %v, want %v", got, want)
	}
}
