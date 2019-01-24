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

func TestCreateIPVlanEndpoint(t *testing.T) {
	assert := assert.New(t)
	macAddr := net.HardwareAddr{0x02, 0x00, 0xCA, 0xFE, 0x00, 0x04}

	expected := &IPVlanEndpoint{
		NetPair: types.NetworkInterfacePair{
			TapInterface: types.TapInterface{
				ID:   "uniqueTestID-5",
				Name: "br5_kata",
				TAPIface: types.NetworkInterface{
					Name: "tap5_kata",
				},
			},
			VirtIface: types.NetworkInterface{
				Name:     "eth5",
				HardAddr: macAddr.String(),
			},

			NetInterworkingModel: types.NetXConnectTCFilterModel,
		},
		EndpointType: IPVlanEndpointType,
	}

	result, err := createIPVlanNetworkEndpoint(5, "")
	assert.NoError(err)

	// the resulting ID  will be random - so let's overwrite to test the rest of the flow
	result.NetPair.ID = "uniqueTestID-5"

	// the resulting mac address will be random - so lets overwrite it
	result.NetPair.VirtIface.HardAddr = macAddr.String()

	assert.Exactly(result, expected)
}
