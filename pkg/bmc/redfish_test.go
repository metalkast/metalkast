package bmc

import (
	"testing"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/stretchr/testify/assert"
)

func TestRedFish_getVirtualMediaCD(t *testing.T) {
	tests := []struct {
		name    string
		media   []*redfish.VirtualMedia
		wantErr bool
	}{
		{
			name: "CD found",
			media: []*redfish.VirtualMedia{
				{
					Entity: common.Entity{
						ID: "Cd",
					},
				},
			},
		},
		{
			name: "CD not found",
			media: []*redfish.VirtualMedia{
				{
					Entity: common.Entity{
						ID: "Cd2",
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "empty media",
			media:   []*redfish.VirtualMedia{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd, err := getVirtualMediaCD(tt.media)

			if tt.wantErr {
				assert.NotNil(t, err, "expected error")
			} else {
				if assert.Nil(t, err, "expected no error") {
					assert.Equal(t, "Cd", cd.ID)
				}
			}
		})
	}
}
