// Copyright (c) 2016 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package hypervisor

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockHypervisorCreateSandbox(t *testing.T) {
	var m *mock
	sandboxID := "mock_sandbox"
	assert := assert.New(t)

	hypervisorConfig := Config{
		KernelPath:     "",
		ImagePath:      "",
		HypervisorPath: "",
	}

	ctx := context.Background()

	// wrong config
	err := m.CreateSandbox(ctx, sandboxID, NetworkNamespace{}, &hypervisorConfig, nil)
	assert.Error(err)

	validHypervisorConfig := Config{
		KernelPath:     fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:      fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath: fmt.Sprintf("%s/%s", testDir, testHypervisor),
	}

	err = m.CreateSandbox(ctx, sandboxID, NetworkNamespace{}, &validHypervisorConfig, nil)
	assert.NoError(err)
}

func TestMockHypervisorStartSandbox(t *testing.T) {
	var m *mock

	assert.NoError(t, m.StartSandbox(10))
}

func TestMockHypervisorStopSandbox(t *testing.T) {
	var m *mock

	assert.NoError(t, m.StopSandbox())
}

func TestMockHypervisorAddDevice(t *testing.T) {
	var m *mock

	assert.NoError(t, m.AddDevice(nil, ImgDev))
}

func TestMockHypervisorGetSandboxConsole(t *testing.T) {
	var m *mock

	expected := ""
	result, err := m.GetSandboxConsole("testSandboxID")
	assert.NoError(t, err)
	assert.Equal(t, result, expected)
}

func TestMockHypervisorSaveSandbox(t *testing.T) {
	var m *mock

	assert.NoError(t, m.SaveSandbox())
}

func TestMockHypervisorDisconnect(t *testing.T) {
	var m *mock

	m.Disconnect()
}
