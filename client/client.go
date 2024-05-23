package client

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/TheJumpCloud/jcapi"

	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"

	"github.com/cloudquery/plugin-pb-go/specs"
	"github.com/cloudquery/plugin-sdk/v3/plugins/source"
	"github.com/cloudquery/plugin-sdk/v3/schema"
	"github.com/rs/zerolog"
)

// the following constants are used for API v2 calls:
const (
	contentType        = "application/json"
	accept             = "application/json"
	searchLimit        = 100
	searchSkipInterval = 100
)

type Client struct {
	Logger          zerolog.Logger
	JumpCloud       *jcapi.JCAPI
	JumpCloudv2     *jcapiv2.APIClient
	JumpCloudv2Auth context.Context
	IsGroups        bool
}

func (c *Client) ID() string {
	// TODO: Change to either your plugin name or a unique dynamic identifier
	return "ID"
}

func New(ctx context.Context, logger zerolog.Logger, s specs.Source, opts source.Options) (schema.ClientMeta, error) {
	var pluginSpec Spec

	if err := s.UnmarshalSpec(&pluginSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin spec: %w", err)
	}
	var (
		apiURL      = config("JUMPCLOUD_API_URL", "https://console.jumpcloud.com/api")
		apiKey      = config("JUMPCLOUD_API_KEY", "")
		apiClientV1 = jcapi.NewJCAPI(apiKey, apiURL)
		apiClientV2 = jcapiv2.NewAPIClient(jcapiv2.NewConfiguration())
	)

	// check if this org is on Groups or Tags:
	isGroups, err := isGroupsOrg(apiURL, apiKey)
	if err != nil {
		log.Fatalf("Could not determine your org type, err='%s'\n", err)
	}
	// if we're on a groups org, instantiate the API client v2:
	var auth context.Context
	if isGroups {
		// instantiate API client v2:
		apiClientV2 = jcapiv2.NewAPIClient(jcapiv2.NewConfiguration())
		apiClientV2.ChangeBasePath(apiURL + "/v2")
		// set up the API key via context:
		auth = context.WithValue(context.TODO(), jcapiv2.ContextAPIKey, jcapiv2.APIKey{
			Key: apiKey,
		})
	}

	return &Client{
		Logger:          logger,
		JumpCloud:       &apiClientV1,
		JumpCloudv2:     apiClientV2,
		IsGroups:        isGroups,
		JumpCloudv2Auth: auth,
	}, nil
}

func config(s, e string) string {
	envVar := os.Getenv(s)
	if envVar != "" {
		return envVar
	}
	return e
}

// isGroupsOrg returns true if this org is groups enabled:
func isGroupsOrg(urlBase string, apiKey string) (bool, error) {
	// instantiate a new API client object:
	client := jcapiv2.NewAPIClient(jcapiv2.NewConfiguration())
	client.ChangeBasePath(urlBase + "/v2")

	// set up the API key via context:
	auth := context.WithValue(context.TODO(), jcapiv2.ContextAPIKey, jcapiv2.APIKey{
		Key: apiKey,
	})

	// set up optional parameters:
	optionals := map[string]interface{}{
		"limit": int32(1), // limit the query to return 1 item
	}
	// in order to check for groups support, we just query for the list of User groups
	// (we just ask to retrieve 1) and check the response status code:
	_, res, err := client.UserGroupsApi.GroupsUserList(auth, contentType, accept, optionals)

	// check if we're using the API v1:
	// we need to explicitly check for 404, since GroupsUserList will also return a json
	// unmarshalling error (err will not be nil) if we're running this endpoint against
	// a Tags org and we don't want to treat this case as an error:
	if res != nil && res.StatusCode == 404 {
		return false, nil
	}

	// if there was any kind of other error, return that:
	if err != nil {
		return false, err
	}

	// if we're using API v2, we're expecting a 200:
	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}
