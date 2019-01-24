// Copyright (c) 2016 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package hypervisor

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/kata-containers/runtime/virtcontainers/store"
)

const testKernel = "kernel"
const testInitrd = "initrd"
const testImage = "image"
const testHypervisor = "hypervisor"

var testDir = ""

const testDisabledAsNonRoot = "Test disabled as requires root privileges"

func testSetType(t *testing.T, value string, expected Type) {
	var hypervisorType Type
	assert := assert.New(t)

	err := (&hypervisorType).Set(value)
	assert.NoError(err)
	assert.Equal(hypervisorType, expected)
}

func TestSetQemuType(t *testing.T) {
	testSetType(t, "qemu", Qemu)
}

func TestSetMockType(t *testing.T) {
	testSetType(t, "mock", Mock)
}

func TestSetUnknownType(t *testing.T) {
	var hypervisorType Type
	assert := assert.New(t)

	err := (&hypervisorType).Set("unknown")
	assert.Error(err)
	assert.NotEqual(hypervisorType, Qemu)
	assert.NotEqual(hypervisorType, Firecracker)
	assert.NotEqual(hypervisorType, Mock)
}

func testStringFromType(t *testing.T, hypervisorType Type, expected string) {
	hypervisorTypeStr := (&hypervisorType).String()
	assert := assert.New(t)
	assert.Equal(hypervisorTypeStr, expected)
}

func TestStringFromQemuType(t *testing.T) {
	hypervisorType := Qemu
	testStringFromType(t, hypervisorType, "qemu")
}

func TestStringFromMockType(t *testing.T) {
	hypervisorType := Mock
	testStringFromType(t, hypervisorType, "mock")
}

func TestStringFromUnknownType(t *testing.T) {
	var hypervisorType Type
	testStringFromType(t, hypervisorType, "")
}

func testNewHypervisorFromType(t *testing.T, hypervisorType Type, expected hypervisor) {
	assert := assert.New(t)
	hy, err := New(hypervisorType)
	assert.NoError(err)
	assert.Exactly(hy, expected)
}

func TestNewHypervisorFromQemuType(t *testing.T) {
	hypervisorType := Qemu
	expectedHypervisor := &qemu{}
	testNewHypervisorFromType(t, hypervisorType, expectedHypervisor)
}

// func TestNewHypervisorFromMockType(t *testing.T) {
// 	hypervisorType := Mock
// 	expectedHypervisor := &mockHypervisor{}
// 	testNewHypervisorFromType(t, hypervisorType, expectedHypervisor)
// }

func TestNewHypervisorFromUnknownType(t *testing.T) {
	var hypervisorType Type
	assert := assert.New(t)

	hy, err := New(hypervisorType)
	assert.Error(err)
	assert.Nil(hy)
}

func testConfigValid(t *testing.T, hypervisorConfig *Config, success bool) {
	err := hypervisorConfig.valid()
	assert := assert.New(t)
	assert.False(success && err != nil)
	assert.False(!success && err == nil)
}

func TestConfigNoKernelPath(t *testing.T) {
	hypervisorConfig := &Config{
		KernelPath:     "",
		ImagePath:      fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath: fmt.Sprintf("%s/%s", testDir, testHypervisor),
	}

	testConfigValid(t, hypervisorConfig, false)
}

func TestConfigNoImagePath(t *testing.T) {
	hypervisorConfig := &Config{
		KernelPath:     fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:      "",
		HypervisorPath: fmt.Sprintf("%s/%s", testDir, testHypervisor),
	}

	testConfigValid(t, hypervisorConfig, false)
}

func TestConfigNoHypervisorPath(t *testing.T) {
	hypervisorConfig := &Config{
		KernelPath:     fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:      fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath: "",
	}

	testConfigValid(t, hypervisorConfig, true)
}

func TestConfigIsValid(t *testing.T) {
	hypervisorConfig := &Config{
		KernelPath:     fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:      fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath: fmt.Sprintf("%s/%s", testDir, testHypervisor),
	}

	testConfigValid(t, hypervisorConfig, true)
}

func TestConfigValidTemplateConfig(t *testing.T) {
	hypervisorConfig := &Config{
		KernelPath:       fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:        fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath:   fmt.Sprintf("%s/%s", testDir, testHypervisor),
		BootToBeTemplate: true,
		BootFromTemplate: true,
	}
	testConfigValid(t, hypervisorConfig, false)

	hypervisorConfig.BootToBeTemplate = false
	testConfigValid(t, hypervisorConfig, false)
	hypervisorConfig.MemoryPath = "foobar"
	testConfigValid(t, hypervisorConfig, false)
	hypervisorConfig.DevicesStatePath = "foobar"
	testConfigValid(t, hypervisorConfig, true)

	hypervisorConfig.BootFromTemplate = false
	hypervisorConfig.BootToBeTemplate = true
	testConfigValid(t, hypervisorConfig, true)
	hypervisorConfig.MemoryPath = ""
	testConfigValid(t, hypervisorConfig, false)
}

func TestConfigDefaults(t *testing.T) {
	assert := assert.New(t)
	hypervisorConfig := &Config{
		KernelPath:     fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:      fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath: "",
	}
	testConfigValid(t, hypervisorConfig, true)

	hypervisorConfigDefaultsExpected := &Config{
		KernelPath:        fmt.Sprintf("%s/%s", testDir, testKernel),
		ImagePath:         fmt.Sprintf("%s/%s", testDir, testImage),
		HypervisorPath:    "",
		NumVCPUs:          DefaultVCPUs,
		MemorySize:        DefaultMemSzMiB,
		DefaultBridges:    DefaultBridges,
		BlockDeviceDriver: DefaultBlockDriver,
		Msize9p:           DefaultMsize9p,
	}

	assert.Exactly(hypervisorConfig, hypervisorConfigDefaultsExpected)
}

func TestAppendParams(t *testing.T) {
	assert := assert.New(t)
	paramList := []Param{
		{
			Key:   "param1",
			Value: "value1",
		},
	}

	expectedParams := []Param{
		{
			Key:   "param1",
			Value: "value1",
		},
		{
			Key:   "param2",
			Value: "value2",
		},
	}

	paramList = appendParam(paramList, "param2", "value2")
	assert.Exactly(paramList, expectedParams)
}

func testSerializeParams(t *testing.T, params []Param, delim string, expected []string) {
	assert := assert.New(t)
	result := SerializeParams(params, delim)
	assert.Exactly(result, expected)
}

func TestSerializeParamsNoParamNoValue(t *testing.T) {
	params := []Param{
		{
			Key:   "",
			Value: "",
		},
	}
	var expected []string

	testSerializeParams(t, params, "", expected)
}

func TestSerializeParamsNoParam(t *testing.T) {
	params := []Param{
		{
			Value: "value1",
		},
	}

	expected := []string{"value1"}

	testSerializeParams(t, params, "", expected)
}

func TestSerializeParamsNoValue(t *testing.T) {
	params := []Param{
		{
			Key: "param1",
		},
	}

	expected := []string{"param1"}

	testSerializeParams(t, params, "", expected)
}

func TestSerializeParamsNoDelim(t *testing.T) {
	params := []Param{
		{
			Key:   "param1",
			Value: "value1",
		},
	}

	expected := []string{"param1", "value1"}

	testSerializeParams(t, params, "", expected)
}

func TestSerializeParams(t *testing.T) {
	params := []Param{
		{
			Key:   "param1",
			Value: "value1",
		},
	}

	expected := []string{"param1=value1"}

	testSerializeParams(t, params, "=", expected)
}

func testDeserializeParams(t *testing.T, parameters []string, expected []Param) {
	assert := assert.New(t)
	result := DeserializeParams(parameters)
	assert.Exactly(result, expected)
}

func TestDeserializeParamsNil(t *testing.T) {
	var parameters []string
	var expected []Param

	testDeserializeParams(t, parameters, expected)
}

func TestDeserializeParamsNoParamNoValue(t *testing.T) {
	parameters := []string{
		"",
	}

	var expected []Param

	testDeserializeParams(t, parameters, expected)
}

func TestDeserializeParamsNoValue(t *testing.T) {
	parameters := []string{
		"param1",
	}
	expected := []Param{
		{
			Key: "param1",
		},
	}

	testDeserializeParams(t, parameters, expected)
}

func TestDeserializeParams(t *testing.T) {
	parameters := []string{
		"param1=value1",
	}

	expected := []Param{
		{
			Key:   "param1",
			Value: "value1",
		},
	}

	testDeserializeParams(t, parameters, expected)
}

func TestAddKernelParamValid(t *testing.T) {
	var config Config
	assert := assert.New(t)

	expected := []Param{
		{"foo", "bar"},
	}

	err := config.AddKernelParam(expected[0])
	assert.NoError(err)
	assert.Exactly(config.KernelParams, expected)
}

func TestAddKernelParamInvalid(t *testing.T) {
	var config Config
	assert := assert.New(t)

	invalid := []Param{
		{"", "bar"},
	}

	err := config.AddKernelParam(invalid[0])
	assert.Error(err)
}

func TestGetHostMemorySizeKb(t *testing.T) {
	assert := assert.New(t)
	type testData struct {
		contents       string
		expectedResult int
		expectError    bool
	}

	data := []testData{
		{
			`
			MemTotal:      1 kB
			MemFree:       2 kB
			SwapTotal:     3 kB
			SwapFree:      4 kB
			`,
			1024,
			false,
		},
		{
			`
			MemFree:       2 kB
			SwapTotal:     3 kB
			SwapFree:      4 kB
			`,
			0,
			true,
		},
	}

	dir, err := ioutil.TempDir("", "")
	assert.NoError(err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "meminfo")
	_, err = GetHostMemorySizeKb(file)
	assert.Error(err)

	for _, d := range data {
		err = ioutil.WriteFile(file, []byte(d.contents), os.FileMode(0640))
		assert.NoError(err)
		defer os.Remove(file)

		hostMemKb, err := GetHostMemorySizeKb(file)

		assert.False((d.expectError && err == nil))
		assert.False((!d.expectError && err != nil))
		assert.NotEqual(hostMemKb, d.expectedResult)
	}
}

// nolint: unused, deadcode
type testNestedVMMData struct {
	content     []byte
	expectedErr bool
	expected    bool
}

// nolint: unused, deadcode
func genericTestRunningOnVMM(t *testing.T, data []testNestedVMMData) {
	assert := assert.New(t)
	for _, d := range data {
		f, err := ioutil.TempFile("", "cpuinfo")
		assert.NoError(err)
		defer os.Remove(f.Name())
		defer f.Close()

		n, err := f.Write(d.content)
		assert.NoError(err)
		assert.Equal(n, len(d.content))

		running, err := RunningOnVMM(f.Name())
		if !d.expectedErr && err != nil {
			t.Fatalf("This test should succeed: %v", err)
		} else if d.expectedErr && err == nil {
			t.Fatalf("This test should fail")
		}

		assert.Equal(running, d.expected)
	}
}

// TestMain is the common main function used by ALL the test functions
// for this package.
func TestMain(m *testing.M) {
	var err error

	flag.Parse()

	testDir, err = ioutil.TempDir("", "vc-qemu-tmp-")
	if err != nil {
		panic(err)
	}

	fmt.Printf("INFO: Creating hypervisor test directory %s\n", testDir)
	err = os.MkdirAll(testDir, store.DirMode)
	if err != nil {
		fmt.Println("Could not create test directories:", err)
		os.Exit(1)
	}

	testQemuKernelPath = filepath.Join(testDir, testKernel)
	testQemuInitrdPath = filepath.Join(testDir, testInitrd)
	testQemuImagePath = filepath.Join(testDir, testImage)
	testQemuPath = filepath.Join(testDir, testHypervisor)

	fmt.Printf("INFO: Creating hypervisor test kernel %s\n", testQemuKernelPath)
	_, err = os.Create(testQemuKernelPath)
	if err != nil {
		fmt.Println("Could not create test kernel:", err)
		os.RemoveAll(testDir)
		os.Exit(1)
	}

	fmt.Printf("INFO: Creating hypervisor test image %s\n", testQemuImagePath)
	_, err = os.Create(testQemuImagePath)
	if err != nil {
		fmt.Println("Could not create test image:", err)
		os.RemoveAll(testDir)
		os.Exit(1)
	}

	fmt.Printf("INFO: Creating hypervisor test hypervisor %s\n", testQemuPath)
	_, err = os.Create(testQemuPath)
	if err != nil {
		fmt.Println("Could not create test hypervisor:", err)
		os.RemoveAll(testDir)
		os.Exit(1)
	}

	ret := m.Run()

	os.RemoveAll(testDir)

	os.Exit(ret)
}
