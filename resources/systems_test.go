// TODO: Pull in device specs ferda boyz via v2 api
// TODO: Clean up un-used columns (like passwords)
// TODO: Provide command execution history (like cloudtrail)

package resources

import (
	"context"
	"os"
	"reflect"
	"testing"

	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/cloudquery/plugin-pb-go/specs"
	"github.com/cloudquery/plugin-sdk/v3/plugins/source"
	"github.com/rs/zerolog"
	"github.com/virtualbeck/cq-source-jumpcloud/client"
)

func TestSystem_getHardwareDetails(t *testing.T) {
	t.Skip() // Skip this until tests are mocked
	type fields struct {
		Id string
	}
	type args struct {
		apiClientV2 *jcapiv2.APIClient
		auth        context.Context
	}
	tests := []struct {
		name        string
		fields      fields
		wantUserIds []string
		wantErr     bool
	}{
		{
			name: "test",
			fields: fields{
				Id: "TEST",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			meta, err := client.New(ctx, zerolog.New(os.Stdout), specs.Source{}, source.Options{})
			if err != nil {
				t.Error(err)
			}
			c := meta.(*client.Client)

			system := &System{
				Id: tt.fields.Id,
			}
			gotUserIds, err := system.getHardwareDetails(c.JumpCloudv2, c.JumpCloudv2Auth)
			if (err != nil) != tt.wantErr {
				t.Errorf("System.getHardwareDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUserIds, tt.wantUserIds) {
				t.Errorf("System.getHardwareDetails() = %v, want %v", gotUserIds, tt.wantUserIds)
			}
		})
	}
}
