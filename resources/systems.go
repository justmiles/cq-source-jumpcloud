// TODO: Pull in device specs ferda boyz via v2 api
// TODO: Clean up un-used columns (like passwords)
// TODO: Provide command execution history (like cloudtrail)

package resources

import (
	"context"
	"fmt"

	"github.com/TheJumpCloud/jcapi"
	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/cloudquery/plugin-sdk/v3/schema"
	"github.com/cloudquery/plugin-sdk/v3/transformers"
	"github.com/virtualbeck/cq-source-jumpcloud/client"
)

type System struct {
	Os                             string  `json:"os,omitempty"`
	TemplateName                   string  `json:"templateName,omitempty"`
	AllowSshRootLogin              bool    `json:"allowSshRootLogin"`
	Id                             string  `json:"_id"`
	LastContact                    string  `json:"lastContact,omitempty"`
	RemoteIP                       string  `json:"remoteIP,omitempty"`
	Active                         bool    `json:"active,omitempty"`
	SshRootEnabled                 bool    `json:"sshRootEnabled"`
	AmazonInstanceID               string  `json:"amazonInstanceID,omitempty"`
	SshPassEnabled                 bool    `json:"sshPassEnabled,omitempty"`
	Version                        string  `json:"version,omitempty"`
	AgentVersion                   string  `json:"agentVersion,omitempty"`
	AllowPublicKeyAuth             bool    `json:"allowPublicKeyAuthentication"`
	Organization                   string  `json:"organization,omitempty"`
	Created                        string  `json:"created,omitempty"`
	Arch                           string  `json:"arch,omitempty"`
	SystemTimezone                 float64 `json:"systemTimeZone,omitempty"`
	AllowSshPasswordAuthentication bool    `json:"allowSshPasswordAuthentication"`
	DisplayName                    string  `json:"displayName"`
	ModifySSHDConfig               bool    `json:"modifySSHDConfig"`
	AllowMultiFactorAuthentication bool    `json:"allowMultiFactorAuthentication"`
	Hostname                       string  `json:"hostname,omitempty"`

	ConnectionHistoryList []string                   `json:"connectionHistory,omitempty"`
	SshdParams            []jcapi.JCSSHDParam        `json:"sshdParams,omitempty"`
	NetworkInterfaces     []jcapi.JCNetworkInterface `json:"networkInterfaces,omitempty"`

	// Derived by JCAPI
	TagList []string `json:"tags,omitempty"`
	Tags    []jcapi.JCTag

	HardwareModel   string `json:"hardware_model,omitempty"`
	HardwareSerial  string `json:"hardware_serial,omitempty"`
	HardwareVendor  string `json:"hardware_vendor,omitempty"`
	HardwareVersion string `json:"hardware_version,omitempty"`
}

func SystemsTable() *schema.Table {
	return &schema.Table{
		Name:      "jumpcloud_systems",
		Resolver:  fetchSystems,
		Transform: transformers.TransformWithStruct(&System{}, transformers.WithPrimaryKeys("Id")),
	}
}

func fetchSystems(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)
	systemsList, err := c.JumpCloud.GetSystems(false)
	if err != nil {
		return fmt.Errorf("could not read systems, err='%s'", err)
	}

	for _, jcSystem := range systemsList {
		system := NewSystem(jcSystem)
		system.getHardwareDetails(c.JumpCloudv2, c.JumpCloudv2Auth)
		res <- system
	}

	return nil
}

func NewSystem(jcSystem jcapi.JCSystem) System {
	return System{
		Os:                             jcSystem.Os,
		TemplateName:                   jcSystem.TemplateName,
		AllowSshRootLogin:              jcSystem.AllowSshRootLogin,
		Id:                             jcSystem.Id,
		LastContact:                    jcSystem.LastContact,
		RemoteIP:                       jcSystem.RemoteIP,
		Active:                         jcSystem.Active,
		SshRootEnabled:                 jcSystem.SshRootEnabled,
		AmazonInstanceID:               jcSystem.AmazonInstanceID,
		SshPassEnabled:                 jcSystem.SshPassEnabled,
		Version:                        jcSystem.Version,
		AgentVersion:                   jcSystem.AgentVersion,
		AllowPublicKeyAuth:             jcSystem.AllowPublicKeyAuth,
		Organization:                   jcSystem.Organization,
		Created:                        jcSystem.Created,
		Arch:                           jcSystem.Arch,
		SystemTimezone:                 jcSystem.SystemTimezone,
		AllowSshPasswordAuthentication: jcSystem.AllowSshPasswordAuthentication,
		DisplayName:                    jcSystem.DisplayName,
		ModifySSHDConfig:               jcSystem.ModifySSHDConfig,
		AllowMultiFactorAuthentication: jcSystem.AllowMultiFactorAuthentication,
		Hostname:                       jcSystem.Hostname,
	}
}

// getUsersBoundToSystemV2 returns the list of users associated with the given system
// for a Groups org using the /v2/systems/<system_id>/users endpoint:
func (system *System) getHardwareDetails(apiClientV2 *jcapiv2.APIClient, auth context.Context) (userIds []string, err error) {
	var graphs []jcapiv2.GraphObjectWithPaths
	for skip := 0; skip == 0 || len(graphs) == searchLimit; skip += searchSkipInterval {
		// set up optional parameters:
		optionals := map[string]interface{}{
			"limit":  int32(searchLimit),
			"skip":   int32(skip),
			"filter": []string{fmt.Sprintf("system_id:eq:%s", system.Id)},
		}
		systemsInsightsInfo, _, err := apiClientV2.SystemInsightsApi.SysteminsightsListSystemInfo(auth, contentType, accept, optionals)
		if err != nil {
			fmt.Println(systemsInsightsInfo)
			return userIds, fmt.Errorf("system %s, err='%s'", system.Id, err)
		}

		for _, info := range systemsInsightsInfo {
			system.HardwareModel = info.HardwareModel
			system.HardwareSerial = info.HardwareSerial
			system.HardwareVendor = info.HardwareVendor
			system.HardwareVersion = info.HardwareVersion
		}
	}
	return
}
