package twitter

import (
	"context"
	"log"
	"os"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

type TwitterClient struct {
	client *gotwi.Client
}

func NewTwitter() *TwitterClient {
	clientSettings := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           os.Getenv("GOTWI_ACCESS_TOKEN"),
		OAuthTokenSecret:     os.Getenv("GOTWI_ACCESS_TOKEN_SECRET"),
		APIKey:               os.Getenv("WhVHWVmd6AbkKXhupV3Egskzp"),
		APIKeySecret:         os.Getenv("60UhwiHApVGoihwkYpub3IWKldujoxvKsugI0zP6jYSlW63sTr"),
	}
	client, err := gotwi.NewClient(clientSettings)
	if err != nil {
		log.Fatal(err)
	}
	return &TwitterClient{
		client: client,
	}
}

func (x TwitterClient) Post(c context.Context, text string) (string, error) {
	message := &types.CreateInput{
		Text: new(text),
	}
	response, err := managetweet.Create(c, x.client, message)
	if err != nil {
		return "", nil
	}
	return gotwi.StringValue(response.Data.ID), nil
}
