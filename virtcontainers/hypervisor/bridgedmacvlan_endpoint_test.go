// Copyright (c) 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package hypervisor

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/kata-containers/runtime/virtcontainers/types"
)

func TestCreateBridgedMacvlanEndpoint(t *testing.T) {
	macAddr := net.HardwareAddr{0x02, 0x00, 0xCA, 0xFE, 0x00, 0x04}

	expected := &BridgedMacvlanEndpoint{
		NetPair: types.NetworkInterfacePair{
			TapInterface: types.TapInterface{
				ID:   "uniqueTestID-4",
				Name: "br4_kata",
				TAPIface: types.NetworkInterface{
					Name: "tap4_kata",
				},
			},
			VirtIface: types.NetworkInterface{
				Name:     "eth4",
				HardAddr: macAddr.String(),
			},
			NetInterworkingModel: types.DefaultNetInterworkingModel,
		},
		EndpointType: BridgedMacvlanEndpointType,
	}

	result, err := createBridgedMacvlanNetworkEndpoint(4, "", types.DefaultNetInterworkingModel)
	assert.NoError(t, err)

	// the resulting ID  will be random - so let's overwrite to test the rest of the flow
	result.NetPair.ID = "uniqueTestID-4"

	// the resulting mac address will be random - so lets overwrite it
	result.NetPair.VirtIface.HardAddr = macAddr.String()

	assert.Exactly(t, result, expected)
}
