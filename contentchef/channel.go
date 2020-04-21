package contentchef

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Response response represents a ContentChef content.
type Response struct {
	PublicID       string         `json:"publicId"`
	Definition     string         `json:"definition"`
	Repository     string         `json:"repository"`
	Payload        interface{}    `json:"payload"`
	OnlineDate     time.Time      `json:"onlineDate"`
	OfflineDate    time.Time      `json:"offlineDate"`
	Metadata       Metadata       `json:"metadata"`
	RequestContext RequestContext `json:"requestContext"`
}

type Metadata struct {
	ID                      int       `json:"id"`
	AuthoringContentID      int       `json:"authoringContentId"`
	ContentVersion          int       `json:"contentVersion"`
	ContentLastModifiedDate time.Time `json:"contentLastModifiedDate"`
	Tags                    []string  `json:"tags"`
	PublishedOn             time.Time `json:"publishedOn"`
}

type PaginatedResponse struct {
	Items          []Response     `json:"items"`
	Total          int            `json:"total"`
	Skip           int            `json:"skip"`
	Take           int            `json:"take"`
	RequestContext RequestContext `json:"requestContext"`
}

type RequestContext struct {
	PublishingChannel string    `json:"publishingChannel"`
	CloudName         string    `json:"cloudName"`
	Timestamp         time.Time `json:"timestamp"`
}

// ContentOptions specifies the parameters to the Channel's Content method.
type ContentOptions struct {
	LegacyMetadata bool `url:"legacyMetadata,omitempty"`
	// The publicId of the content you want to retrieve
	PublicID string `url:"publicId"`
}

// SearchOptions specifies the parameters to the Channel's and Online Channel's Search method.
type SearchOptions struct {
	// The offset you want to have in the search result.
	Skip int `url:"skip"`
	// The number of contents you want to retrieve
	Take int `url:"take"`
	// A slice containing the publicIds of the content you want to retrieve
	PublicID []string `url:"publicId,omitempty"`
	// A slice containing the definitions of the content you want to retrieve
	ContentDefinition []string `url:"contentDefinition,omitempty"`
	Repositories      []string `url:"repositories,omitempty"`
	LegacyMetadata    bool     `url:"legacyMetadata,omitempty"`
	// A slice containing the tags of the content you want to retrieve
	Tags []string `url:"tags,omitempty"`
	// Properties filters you want to apply
	PropFilters PropFilters `url:"propFilters,omitempty"`
	// How you want to sort your content
	Sorting Sorting `url:"sorting,omitempty"`
}

type SortingField struct {
	FieldName string `url:"fieldName"`
	Ascending bool   `url:"ascending"`
}

type Sorting []SortingField

func (s Sorting) EncodeValues(key string, v *url.Values) error {
	v.Set("sorting", serializeSorting(s))
	return nil
}

func serializeSorting(s Sorting) string {
	values := make([]string, len(s))
	for i, field := range s {
		if len(field.FieldName) == 0 {
			continue
		}
		sorter := "-"
		if field.Ascending {
			sorter = "+"
		}
		values[i] = sorter + strings.TrimSpace(field.FieldName)

	}
	return strings.Join(values, ",")
}

type PropFilters struct {
	// The logical operator you want to apply.
	// Possible values:
	// AND, OR
	Condition string           `json:"condition,omitempty"`
	Items     []PropFilterItem `json:"items,omitempty"`
}

type PropFilterItem struct {
	Field string `json:"field,omitempty"`
	// The operator you want to apply.
	// Possible values:
	// CONTAINS, CONTAINS_IC, EQUALS, EQUALS_IC, IN, IN_IC, STARTS_WITH, STARTS_WITH_IC
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

func (p PropFilters) EncodeValues(key string, v *url.Values) error {
	if len(p.Items) == 0 {
		return nil
	}
	j, err := json.Marshal(p)
	if err != nil {
		return err
	}
	v.Set("propFilters", string(j))
	return nil
}

// OnlineChannel retrieves contents that are in live status
type OnlineChannel struct {
	client *Client

	name   string
	apiKey string
}

// OnlineChannel returns an OnlineChannel instance.
//
// It takes the name and the apiKey used to communicate with online channels,
// both fields must not be an empty string.
func (c *Client) OnlineChannel(name, apiKey string) (*OnlineChannel, error) {
	if name == "" {
		return nil, errors.New("name seems to be an empty string")
	}
	if apiKey == "" {
		return nil, errors.New("apiKey seems to be an empty string")
	}
	return &OnlineChannel{
		client: c,
		name:   name,
		apiKey: apiKey,
	}, nil
}

// Content returns a Response reference.
// It will retrieve for a single content using the content's publicID and returns a Content reference.
//
// It takes a context and a a reference to a ContentOptions struct
// if you are not sure about the context to use, use context.TODO()
func (s *OnlineChannel) Content(ctx context.Context, config *ContentOptions) (*Response, error) {
	path := getOnlineEndpoint(s.client.SpaceID, "content", s.name)

	r := &Response{}
	err := s.client.get(ctx, path, s.apiKey, config, r)
	return r, err
}

// Search returns a PaginatedResponse reference.
// It will retrieve for a range of contents using multiple matching criteria
// like content definition name, publishing dates and more.
//
// It takes a context and a a reference to a SearchOptions struct
// if you are not sure about the context to use, use context.TODO()
func (s *OnlineChannel) Search(ctx context.Context, config *SearchOptions) (*PaginatedResponse, error) {

	path := getOnlineEndpoint(s.client.SpaceID, "search/v2", s.name)

	r := &PaginatedResponse{}
	err := s.client.get(ctx, path, s.apiKey, config, r)
	return r, err
}

func getOnlineEndpoint(spaceID string, method, channel string) string {
	return fmt.Sprintf("/space/%s/online/%s/%s", spaceID, method, channel)
}

// PreviewChannel retrieves the contents that are in both in live stage and live state and event contents
type PreviewChannel struct {
	client *Client

	name   string
	apiKey string
	state  string
}

// PreviewChannel retruns a preview channel reference
// It will retrieve for contents that are not visible in the current date.
//
// It takes the name and the apiKey used to communicate with preview channels, and the publishing status of
// the content you want to retrieve.
func (c *Client) PreviewChannel(name, apiKey, state string) (*PreviewChannel, error) {
	if name == "" {
		return nil, errors.New("name seems to be an empty string")
	}
	if apiKey == "" {
		return nil, errors.New("apiKey seems to be an empty string")
	}
	if state != "live" && state != "staging" {
		return nil, errors.New("`State must be either 'live' or 'staging")
	}
	return &PreviewChannel{
		client: c,
		name:   name,
		apiKey: apiKey,

		state: state,
	}, nil
}

// Content returns a Response reference.
// It will retrieve for a single content using the content's publicID and returns a Content reference.
//
// It takes a context and a a reference to a ContentOptions struct
// if you are not sure about the context to use, use context.TODO()
func (s *PreviewChannel) Content(ctx context.Context, config *ContentOptions) (*Response, error) {
	path := getPreviewEndpoint(s.client.SpaceID, "content", s.name, s.state)

	var targetDate string
	if !s.client.TargetDate.IsZero() {
		targetDate = s.client.TargetDate.Format(time.RFC3339)
	}
	urlParams := struct {
		ContentOptions
		TargetDate string `url:"targetDate,omitempty"`
	}{
		*config,
		targetDate,
	}

	r := &Response{}
	err := s.client.get(ctx, path, s.apiKey, urlParams, r)
	return r, err
}

// Search returns a PaginatedResponse reference.
// It will retrieve for a range of contents using multiple matching criteria
// like content definition name, publishing dates and more.
//
// It takes a context and a a reference to a SearchOptions struct
// if you are not sure about the context to use, use context.TODO()
func (s *PreviewChannel) Search(ctx context.Context, config *SearchOptions) (*PaginatedResponse, error) {
	path := getPreviewEndpoint(s.client.SpaceID, "search/v2", s.name, s.state)

	var targetDate string
	if !s.client.TargetDate.IsZero() {
		targetDate = s.client.TargetDate.Format(time.RFC3339)
	}
	urlParams := struct {
		SearchOptions
		TargetDate string `url:"targetDate,omitempty"`
	}{
		*config,
		targetDate,
	}

	r := &PaginatedResponse{}
	err := s.client.get(ctx, path, s.apiKey, urlParams, r)
	return r, err
}

func getPreviewEndpoint(spaceID string, method, channel, state string) string {
	return fmt.Sprintf("/space/%s/preview/%s/%s/%s", spaceID, state, method, channel)
}
