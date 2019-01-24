// Copyright (c) 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package hypervisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/kata-containers/runtime/virtcontainers/types"
)

func TestCreateMacvtapEndpoint(t *testing.T) {
	netInfo := types.NetworkInfo{
		Iface: types.NetlinkIface{
			Type: "macvtap",
		},
	}
	expected := &MacvtapEndpoint{
		EndpointType:       MacvtapEndpointType,
		EndpointProperties: netInfo,
	}

	result, err := createMacvtapNetworkEndpoint(netInfo)
	assert.NoError(t, err)
	assert.Exactly(t, result, expected)
}
