// Copyright (c) 2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package hypervisor

import (
	"fmt"

	"github.com/containernetworking/plugins/pkg/ns"
	persistapi "github.com/kata-containers/runtime/virtcontainers/persist/api"
	"github.com/kata-containers/runtime/virtcontainers/types"
)

// IPVlanEndpoint represents a ipvlan endpoint that is bridged to the VM
type IPVlanEndpoint struct {
	NetPair            types.NetworkInterfacePair
	EndpointProperties types.NetworkInfo
	EndpointType       EndpointType
	PCIAddr            string
}

func createIPVlanNetworkEndpoint(idx int, ifName string) (*IPVlanEndpoint, error) {
	if idx < 0 {
		return &IPVlanEndpoint{}, fmt.Errorf("invalid network endpoint index: %d", idx)
	}

	// Use tc filtering for ipvlan, since the other inter networking models will
	// not work for ipvlan.
	interworkingModel := types.NetXConnectTCFilterModel
	netPair, err := createNetworkInterfacePair(idx, ifName, interworkingModel)
	if err != nil {
		return nil, err
	}

	endpoint := &IPVlanEndpoint{
		NetPair:      netPair,
		EndpointType: IPVlanEndpointType,
	}
	if ifName != "" {
		endpoint.NetPair.VirtIface.Name = ifName
	}

	return endpoint, nil
}

// Properties returns properties of the interface.
func (endpoint *IPVlanEndpoint) Properties() types.NetworkInfo {
	return endpoint.EndpointProperties
}

// Name returns name of the veth interface in the network pair.
func (endpoint *IPVlanEndpoint) Name() string {
	return endpoint.NetPair.VirtIface.Name
}

// HardwareAddr returns the mac address that is assigned to the tap interface
// in th network pair.
func (endpoint *IPVlanEndpoint) HardwareAddr() string {
	return endpoint.NetPair.TAPIface.HardAddr
}

// Type identifies the endpoint as a virtual endpoint.
func (endpoint *IPVlanEndpoint) Type() EndpointType {
	return endpoint.EndpointType
}

// SetProperties sets the properties for the endpoint.
func (endpoint *IPVlanEndpoint) SetProperties(properties types.NetworkInfo) {
	endpoint.EndpointProperties = properties
}

// PciAddr returns the PCI address of the endpoint.
func (endpoint *IPVlanEndpoint) PciAddr() string {
	return endpoint.PCIAddr
}

// SetPciAddr sets the PCI address of the endpoint.
func (endpoint *IPVlanEndpoint) SetPciAddr(pciAddr string) {
	endpoint.PCIAddr = pciAddr
}

// NetworkPair returns the network pair of the endpoint.
func (endpoint *IPVlanEndpoint) NetworkPair() *types.NetworkInterfacePair {
	return &endpoint.NetPair
}

// Attach for virtual endpoint bridges the network pair and adds the
// tap interface of the network pair to the hypervisor.
func (endpoint *IPVlanEndpoint) Attach(h Hypervisor) error {
	if err := xConnectVMNetwork(endpoint, h); err != nil {
		networkLogger().WithError(err).Error("Error bridging virtual ep")
		return err
	}

	return h.AddDevice(endpoint, NetDev)
}

// Detach for the virtual endpoint tears down the tap and bridge
// created for the veth interface.
func (endpoint *IPVlanEndpoint) Detach(netNsCreated bool, netNsPath string) error {
	// The network namespace would have been deleted at this point
	// if it has not been created by virtcontainers.
	if !netNsCreated {
		return nil
	}

	return DoNetNS(netNsPath, func(_ ns.NetNS) error {
		return xDisconnectVMNetwork(endpoint)
	})
}

// HotAttach for physical endpoint not supported yet
func (endpoint *IPVlanEndpoint) HotAttach(h Hypervisor) error {
	return fmt.Errorf("IPVlanEndpoint does not support Hot attach")
}

// HotDetach for physical endpoint not supported yet
func (endpoint *IPVlanEndpoint) HotDetach(h Hypervisor, netNsCreated bool, netNsPath string) error {
	return fmt.Errorf("IPVlanEndpoint does not support Hot detach")
}

func (endpoint *IPVlanEndpoint) save() persistapi.NetworkEndpoint {
	netpair := saveNetIfPair(&endpoint.NetPair)

	return persistapi.NetworkEndpoint{
		Type: string(endpoint.Type()),
		IPVlan: &persistapi.IPVlanEndpoint{
			NetPair: *netpair,
		},
	}
}

func (endpoint *IPVlanEndpoint) load(s persistapi.NetworkEndpoint) {
	endpoint.EndpointType = IPVlanEndpointType

	if s.IPVlan != nil {
		netpair := loadNetIfPair(&s.IPVlan.NetPair)
		endpoint.NetPair = *netpair
	}
}