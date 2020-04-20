<div align="center">
  <img src="assets/logo-banner.svg" height="128"/>
</div>

Content Chef Go SDK
===========================

[![Go Report](https://goreportcard.com/badge/github.com/ContentChef/contentchef-go)](https://goreportcard.com/report/github.com/ContentChef/contentchef-go)


[Content Chef](https://www.contentchef.io/) Go SDK.

# Requirements

In order to use this SDK, you will need

* An active ContentChef account
* Go

# Installation

## API

### ContentChef client

First you have to initialize the ContentChef client.

```go
"github.com/ContentChef/contentchef-go/contentchef"

myOptions := &contentchef.Options{
    BaseUrl: "https://api.contentchef.io/",
    APIKey:  "super_secret",
    SpaceID: "your_contentchef_space_id",
}

_, cf := contentchef.New(opt)

```

Here are the fields of the configuration that can be passed to the Constructor

```go
// contentchef.Options

type Options struct {
	// REQUIRED The base url of your contentchef instance
	BaseUrl string
	// REQUIRED Your api key
	APIKey string
	// REQUIRED Your space id
	SpaceID string
	// The http client to communicate with the ContentChef API
	Client *http.Client
	//targetDateResolver is used to retrieve contents in the preview channel in a specific dare different from the current date
	// a valid targetDateResolver must return a valid date expressed using the ISO format like 2019-08-16T12:22:232Z
	TargetDateResolver func() string
}
```

### Channels

A channel is a collector of contents.

The SDK returns two channels: `OnlineChannel` and a `PreviewChannel`.
 
With `OnlineChannel` you can retrieve contents which are in *live* state and which are actually visible, while with the `PreviewChannel` you can retrieve contents which are in in both *stage* and *live* state and even contents that are not visible in the current date 

Both `OnlineChannel` and `PreviewChannel` have two methods which are *GetContent* and *Search* that accepts the same paramters.
eg.

```go
func (s *OnlineChannel) Content(ctx context.Context, config *ContentOptions) (*Response, error) {
    // ...
}
func (s *OnlineChannel) Search(ctx context.Context, config *SearchOptions) (*PaginatedResponse, error) {
    // ...
}
```

So, if you want to acheive polymorphic beaviour you can define an interface like this

```go
type Channel interface {
	Content(ctx context.Context, config *ContentOptions) (*Response, error)
	Search(ctx context.Context, config *SearchOptions) (*PaginatedResponse, error)
}
```
     
You can use the **Content** methods to collect a specific content by it's own `PublicId`, to retrieve, for example to retrieve a single post from your blog, a single image from a gallery or a set of articles from your featured articles list.
Otherwise you can use the **Search** methods to find content with multiple matching criteria, like content definition name, publishing dates and more.

Example:

First you have to get an `OnlineChannel` or `PreviewChannel` instance, Then you cause you instance for query for content.

```go
"github.com/ContentChef/contentchef-go/contentchef"
// ...

myOptions := &contentchef.ClientOptions{
    BaseURL: "https://api.contentchef.io/",
    APIKey:  "super_secret",
    SpaceID: "your_contentchef_space_id",
}
_, cf := contentchef.New(opt)

// An OnlineChannel will query only published contents in live state in the current date
chOnline := cf.GetOnlineChannel("your_channel", "super_secret")

// A PreviewChannel will query only the published content with a staging state
chPreview := cf.GetPreviewChannel(cf, "your_channel", "STAGING")

conf := &channel.GetContentOptions{
    PublicID: "my_public_id",
}

// will retrieve from the channel a single content
// GetContent accepts two parametrs.
// A context object (if you are unsure about it use Context.TODO())
// The GetContent configuration object.
myContent, _, err := ch.GetContent(context.TODO(), conf)

// Here is the GetContent confuguration object.
// channel.GetContentOptions
// ContentOptions specifies the parameters to the Channel's Content method.
type ContentOptions struct {
	LegacyMetadata bool `url:"legacyMetadata,omitempty"`
	// The publicId of the content you want to retrieve
	PublicID string `url:"publicId"`
}

searchConf := &channel.SearchOptions{
	Take:    10,
    Sorting: []chan

// will retrieve from the channel website a single content
// Search accepts two parametrs.
// A context object (if you are unsure about it use Context.TODO())
// The Search configuration object
myPaginatedResponse, err := ch.Search(context.TODO(), conf)

//Here is the Search configuration object
// Search Options specifies the parameters to the Channel's and Online Channel's Search method.
type SearchOptions struct {
	// The offset you want to have in the search result.
	Skip int `url:"skip"`
	// The number of contents you want to retrieve
	Take int `url:"take"`
	// A slice containing the publicIds of the content you want to retrieve
	PublicID []string `url:"publicId,omitempty"`
	// A slice containing the definitions of the content you want to retrieve
	ContentDefinition []string `url:"contentDefiniton,omitempty"`
	Repositories      []string `url:"repositories,omitempty"`
	LegacyMetadata    bool     `url:"legacyMetadata,omitempty"`
	// A slice containing the tags of the content you want to retrieve
	Tags []string `url:"tags,omitempty"`
	// Proprerties filters you want to apply
	PropFilters PropFilters `url:"propFilters,omitempty"`
	// How you want to sort your content
	Sorting Sorting `url:"sorting,omitempty"`
}

}
```
