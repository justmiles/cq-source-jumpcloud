package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/TheJumpCloud/jcapi"
	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/cloudquery/plugin-sdk/v3/schema"
	"github.com/cloudquery/plugin-sdk/v3/transformers"
	"github.com/virtualbeck/cq-source-jumpcloud/client"
)

const (
	// the following constants are used for API v2 calls:
	contentType        = "application/json"
	accept             = "application/json"
	searchLimit        = 100
	searchSkipInterval = 100
)

type UsersBoundToSystem struct {
	Id               string `json:"_id"`
	DisplayName      string
	Hostname         string
	Active           bool
	AmazonInstanceID string
	OS               string
	Version          string
	AgentVersion     string
	Created          string
	LastContact      string
	UserName         string
	Email            string
}

func UsersBoundToSystemTable() *schema.Table {
	return &schema.Table{
		Name:      "jumpcloud_users_bound_to_system",
		Resolver:  fetchUsersBoundToSystem,
		Transform: transformers.TransformWithStruct(&UsersBoundToSystem{}),
	}
}

func fetchUsersBoundToSystem(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	systemsList, err := c.JumpCloud.GetSystems(false)
	if err != nil {
		return fmt.Errorf("could not read systems, err='%s'", err)
	}

	for _, system := range systemsList {

		outLine := []string{system.Id, system.DisplayName, system.Hostname, fmt.Sprintf("%t", system.Active),
			system.AmazonInstanceID, system.Os, system.Version, system.AgentVersion, system.Created,
			system.LastContact}

		var usersBoundToSystem = UsersBoundToSystem{
			Id:               system.Id,
			DisplayName:      system.DisplayName,
			Hostname:         system.Hostname,
			Active:           system.Active,
			AmazonInstanceID: system.AmazonInstanceID,
			OS:               system.Os,
			Version:          system.Version,
			AgentVersion:     system.AgentVersion,
			Created:          system.Created,
			LastContact:      system.LastContact,
		}

		var userIds []string

		if c.IsGroups {
			userIds, err = getUsersBoundToSystemV2(c.JumpCloudv2, c.JumpCloudv2Auth, system.Id)
		} else {
			userIds, err = getUsersBoundToSystemV1(c.JumpCloud, system.Id)
		}

		if err != nil {
			// if we fail to retrieve users for the current system, log a msg:
			log.Printf("Failed to retrieve system user bindings: err='%s'\n", err)
			// make sure we still write the system details before skipping:
			res <- usersBoundToSystem
			continue
		}

		// get details for each bound user and append it to the current system:
		for _, userId := range userIds {
			user, err := c.JumpCloud.GetSystemUserById(userId, false)
			if err != nil {
				log.Printf("Could not retrieve system user for ID '%s', err='%s'\n", userId, err)
			} else {
				outLine = append(outLine, fmt.Sprintf("%s (%s)", user.UserName, user.Email))
				usersBoundToSystem.UserName = user.UserName
				usersBoundToSystem.Email = user.Email
				res <- usersBoundToSystem
			}
		}

	}

	return nil
}

// getUsersBoundToSystemV1 returns the list of users associated with the given system
// for a Tags org using the /systems/<system_id>/users endpoint:
// This endpoint will return all the system-user bindings including those made
// via tags and via direct system-user binding
func getUsersBoundToSystemV1(apiClientV1 *jcapi.JCAPI, systemId string) (userIds []string, err error) {

	systemUserBindings, err := apiClientV1.GetSystemUserBindingsById(systemId)
	if err != nil {
		return userIds, fmt.Errorf("could not get system user bindings for system %s, err='%s'", systemId, err)
	}
	// add the retrieved user Ids to our userIds list:
	for _, systemUserBinding := range systemUserBindings {
		userIds = append(userIds, systemUserBinding.UserId)
	}
	return
}

// getUsersBoundToSystemV2 returns the list of users associated with the given system
// for a Groups org using the /v2/systems/<system_id>/users endpoint:
func getUsersBoundToSystemV2(apiClientV2 *jcapiv2.APIClient, auth context.Context, systemId string) (userIds []string, err error) {
	var graphs []jcapiv2.GraphObjectWithPaths
	for skip := 0; skip == 0 || len(graphs) == searchLimit; skip += searchSkipInterval {
		// set up optional parameters:
		optionals := map[string]interface{}{
			"limit": int32(searchLimit),
			"skip":  int32(skip),
		}
		graphs, _, err := apiClientV2.SystemsApi.GraphSystemTraverseUser(auth, systemId, contentType, accept, optionals)
		if err != nil {
			return userIds, fmt.Errorf("could not retrieve users for system %s, err='%s'", systemId, err)
		}
		// add the retrieved user Ids to our userIds list:
		for _, graph := range graphs {
			userIds = append(userIds, graph.Id)
		}
	}
	return
}
