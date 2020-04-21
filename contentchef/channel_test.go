package contentchef

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func Test_serializeSorting(t *testing.T) {
	type args struct {
		s Sorting
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "An empty string will not be serialized",
			args: args{s: Sorting{{FieldName: "", Ascending: false}}},
			want: "",
		},
		{
			name: "A string is used, will be returned trimmed",
			args: args{s: Sorting{{FieldName: " publicId", Ascending: true}}},
			want: "+publicId",
		},
		{
			name: "Array with one item with publicId descrending",
			args: args{s: Sorting{{FieldName: "publicId", Ascending: false}}},
			want: "-publicId",
		},
		{
			name: "Array with one item with publicId ascending",
			args: args{s: Sorting{{FieldName: "publicId", Ascending: true}}},
			want: "+publicId",
		},
		{
			name: "Array with one item with publicId ascending",
			args: args{s: Sorting{{FieldName: "publicId", Ascending: true}, {FieldName: "onlineDate", Ascending: false}}},
			want: "+publicId,-onlineDate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serializeSorting(tt.args.s); got != tt.want {
				t.Errorf("serializeSorting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOnlineEndpoint(t *testing.T) {
	type args struct {
		spaceID string
		method  string
		channel string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "getOnlineEndpoint should return the correct endpoint relative url",
			args: args{
				spaceID: "aSpace",
				method:  "content",
				channel: "foo",
			},
			want: "/space/aSpace/online/content/foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOnlineEndpoint(tt.args.spaceID, tt.args.method, tt.args.channel); got != tt.want {
				t.Errorf("getOnlineEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPreviewEndpoint(t *testing.T) {
	type args struct {
		spaceID string
		method  string
		channel string
		state   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "getOnlineEndpoint should return the correct endpoint relative url",
			args: args{
				spaceID: "aSpace",
				method:  "content",
				channel: "foo",
				state:   "live",
			},
			want: "/space/aSpace/preview/foo/live/content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPreviewEndpoint(tt.args.spaceID, tt.args.state, tt.args.method, tt.args.channel); got != tt.want {
				t.Errorf("getOnlineEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetOnlineChannel(t *testing.T) {
	setup()
	type fields struct {
		httpClient *http.Client
		BaseURL    *url.URL
		SpaceID    string
		TargetDate time.Time
	}
	type args struct {
		name   string
		apiKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *OnlineChannel
		wantErr bool
	}{
		{
			name: "Channel's name cannot be an empty string",
			args: args{
				name:   "",
				apiKey: "superSecret",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Channel's api key cannot be an empty string",
			args: args{
				name:   "channelName",
				apiKey: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "With the right parameters, an online channel should be created",
			args: args{
				name:   "channelName",
				apiKey: "superSecret",
			},
			want: &OnlineChannel{
				client: client,
				name:   "channelName",
				apiKey: "superSecret",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GetOnlineChannel(tt.args.name, tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetOnlineChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetOnlineChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetPreviewChannel(t *testing.T) {
	setup()
	type fields struct {
		httpClient *http.Client
		BaseURL    *url.URL
		SpaceID    string
		TargetDate time.Time
	}
	type args struct {
		name   string
		apiKey string
		state  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PreviewChannel
		wantErr bool
	}{
		{
			name: "Channel's name cannot be an empty string",
			args: args{
				name:   "",
				apiKey: "superSecret",
				state:  "live",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Channel's api key cannot be an empty string",
			args: args{
				name:   "channelName",
				apiKey: "",
				state:  "live",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Channel's publishing status cannot be an empty string",
			args: args{
				name:   "channelName",
				apiKey: "superSecret",
				state:  "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "With the right parameters, an online channel should be created",
			args: args{
				name:   "channelName",
				apiKey: "superSecret",
				state:  "live",
			},
			want: &PreviewChannel{
				client: client,
				name:   "channelName",
				apiKey: "superSecret",
				state:  "live",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GetPreviewChannel(tt.args.name, tt.args.apiKey, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetPreviewChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetPreviewChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}
