// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/anypb"

	_go "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"

	"infra/cros/cmd/common_lib/common"
)

func TestDependencyInjectionBasic(t *testing.T) {
	Convey("basic injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
			},
		}
		dut_address_proto := &labapi.IpEndpoint{
			Address: "localhost",
			Port:    4040,
		}
		storage := common.NewInjectableStorage()
		err := storage.Set("dut_primary", dut_address_proto)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		err = common.Inject(original_proto, "dutServer", storage, "dut_primary")

		So(err, ShouldBeNil)
		So(original_proto.DutServer, ShouldNotBeNil)
		So(original_proto.DutServer.Address, ShouldEqual, dut_address_proto.Address)
		So(original_proto.DutServer.Port, ShouldEqual, dut_address_proto.Port)
	})

	Convey("IpEndpoint direct injection", t, func() {
		original_proto := &labapi.IpEndpoint{}
		dut_address_proto := &labapi.IpEndpoint{
			Address: "localhost",
			Port:    4040,
		}
		storage := common.NewInjectableStorage()
		err := storage.Set("cros-dut", dut_address_proto)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		err = common.Inject(original_proto, "", storage, "cros-dut")

		So(err, ShouldBeNil)
		So(original_proto, ShouldNotBeNil)
		So(original_proto.Address, ShouldEqual, dut_address_proto.Address)
		So(original_proto.Port, ShouldEqual, dut_address_proto.Port)
	})

	Convey("test injection", t, func() {
		original_proto := &skylab_test_runner.ContainerRequest{
			DynamicIdentifier: "cros-provision",
			Container: &api.Template{
				Container: &api.Template_CrosProvision{
					CrosProvision: &api.CrosProvisionTemplate{
						InputRequest: &api.CrosProvisionRequest{},
					},
				},
			},
			ContainerImageKey: "cros-provision",
			DynamicDeps: []*skylab_test_runner.DynamicDep{
				{
					Key:   "crosProvision.inputRequest.dut",
					Value: "dut_primary",
				},
				{
					Key:   "crosProvision.inputRequest.dutServer",
					Value: "cros-dut",
				},
			},
		}
		dut_address_proto := &labapi.IpEndpoint{
			Address: "localhost",
			Port:    4040,
		}
		dut := &labapi.Dut{
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{
					Name: "hello",
				},
			},
		}
		storage := common.NewInjectableStorage()
		err := storage.Set("cros-dut", dut_address_proto)
		So(err, ShouldBeNil)
		err = storage.Set("dut_primary", dut)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		for _, dep := range original_proto.DynamicDeps {
			err := common.Inject(original_proto.Container, dep.Key, storage, dep.Value)
			So(err, ShouldBeNil)
		}
	})
}

func TestDependencyInjectionArray(t *testing.T) {
	Convey("array injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
			},
		}
		dut_address_protos := []*labapi.IpEndpoint{
			{
				Address: "not_expected",
				Port:    4040,
			},
			{
				Address: "expected",
				Port:    1234,
			},
		}

		storage := common.NewInjectableStorage()
		err := storage.Set("duts", dut_address_protos)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		err = common.Inject(original_proto, "dutServer", storage, "duts.1")

		So(err, ShouldBeNil)
		So(original_proto.DutServer, ShouldNotBeNil)
		So(original_proto.DutServer.Address, ShouldEqual, dut_address_protos[1].Address)
		So(original_proto.DutServer.Port, ShouldEqual, dut_address_protos[1].Port)
	})
}

func TestDependencyInjectionArrayAppend(t *testing.T) {
	Convey("array injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
				Packages: []*testapi.ProvisionState_Package{
					{
						PackagePath: &_go.StoragePath{
							Path: "a",
						},
					},
					{
						PackagePath: &_go.StoragePath{
							Path: "b",
						},
					},
					{
						PackagePath: &_go.StoragePath{
							Path: "c",
						},
					},
				},
			},
		}
		new_package := &testapi.ProvisionState_Package{
			PackagePath: &_go.StoragePath{
				Path: "d",
			},
		}

		storage := common.NewInjectableStorage()
		err := storage.Set("package", new_package)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		err = common.Inject(original_proto, "provisionState.packages", storage, "package")

		So(err, ShouldBeNil)
		So(original_proto.ProvisionState.Packages, ShouldHaveLength, 4)
		So(original_proto.ProvisionState.Packages[3].PackagePath.Path, ShouldEqual, new_package.PackagePath.Path)
	})
}

func TestDependencyInjectionArrayOverride(t *testing.T) {
	Convey("array override injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
				Packages:      []*testapi.ProvisionState_Package{},
			},
		}
		new_packages := []*testapi.ProvisionState_Package{
			{
				PackagePath: &_go.StoragePath{
					Path: "d",
				},
			},
			{
				PackagePath: &_go.StoragePath{
					Path: "e",
				},
			},
			{
				PackagePath: &_go.StoragePath{
					Path: "f",
				},
			},
		}

		storage := common.NewInjectableStorage()
		err := storage.Set("packages", new_packages)
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		err = common.Inject(original_proto, "provisionState.packages", storage, "packages")

		So(err, ShouldBeNil)
		So(original_proto.ProvisionState.Packages, ShouldHaveLength, 3)
		So(original_proto.ProvisionState.Packages[0].PackagePath.Path, ShouldEqual, new_packages[0].PackagePath.Path)
		So(original_proto.ProvisionState.Packages[1].PackagePath.Path, ShouldEqual, new_packages[1].PackagePath.Path)
		So(original_proto.ProvisionState.Packages[2].PackagePath.Path, ShouldEqual, new_packages[2].PackagePath.Path)
	})
}

func TestDependencyInjectionFullTest(t *testing.T) {
	storage := common.NewInjectableStorage()
	req := &skylab_test_runner.CrosTestRunnerRequest{
		StartRequest: &skylab_test_runner.CrosTestRunnerRequest_Build{
			Build: &skylab_test_runner.BuildMode{},
		},
		Params: &skylab_test_runner.CrosTestRunnerParams{},
	}
	err := storage.Set("req", req)
	if err != nil {
		t.Fatalf("%s", err)
	}

	Convey("NoticeChangesInStateKeeper", t, func() {
		req.GetParams().Keyvals = map[string]string{
			"build_target": "drallion",
		}
		err := storage.LoadInjectables()
		So(err, ShouldBeNil)

		buildTarget, err := storage.Get("req.params.keyvals.build_target")
		So(err, ShouldBeNil)
		So(buildTarget, ShouldEqual, "drallion")
	})

	Convey("CanAddStringToInjectables", t, func() {
		err := storage.Set("hello", "world!")
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		str, err := storage.Get("hello")
		So(err, ShouldBeNil)
		So(str, ShouldEqual, "world!")
	})

	Convey("CanAddStringArrayToInjectables", t, func() {
		err := storage.Set("hello", []string{"World1", "World2"})
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		arr, err := storage.Get("hello")
		So(err, ShouldBeNil)
		So(arr, ShouldHaveLength, 2)
	})

	Convey("CanAddAndPullParams", t, func() {
		req.Params.TestSuites = []*api.TestSuite{
			{
				Name: "Test1",
			}, {
				Name: "Test2",
			}, {
				Name: "Test3",
			},
		}
		So(err, ShouldBeNil)
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)

		testSuites, err := storage.Get("req.params.testSuites")
		So(err, ShouldBeNil)
		So(testSuites, ShouldHaveLength, 3)

		req.Params.TestSuites = append(req.Params.TestSuites, &testapi.TestSuite{Name: "Test4"})
		err = storage.LoadInjectables()
		So(err, ShouldBeNil)
		testSuites, err = storage.Get("req.params.testSuites")
		So(err, ShouldBeNil)
		So(testSuites, ShouldHaveLength, 4)
	})

	Convey("CanDoDependencyInjection", t, func() {
		endpoint := &labapi.IpEndpoint{
			Address: "localhost",
			Port:    1234,
		}
		err := storage.Set("cros-test", endpoint)
		So(err, ShouldBeNil)
		req.Params.TestSuites = []*api.TestSuite{
			{
				Name: "Test1",
			}, {
				Name: "Test2",
			}, {
				Name: "Test3",
			},
		}
		testRequest := &skylab_test_runner.TestRequest{
			ServiceAddress: &labapi.IpEndpoint{},
			TestRequest:    &testapi.CrosTestRequest{},
			DynamicDeps: []*skylab_test_runner.DynamicDep{
				{
					Key:   "serviceAddress",
					Value: "cros-test",
				},
				{
					Key:   "testRequest.testSuites",
					Value: "req.params.testSuites",
				},
			},
		}
		req.OrderedTasks = []*skylab_test_runner.CrosTestRunnerRequest_Task{
			{
				Task: &skylab_test_runner.CrosTestRunnerRequest_Task_Test{
					Test: testRequest,
				},
			},
		}

		err = common.InjectDependencies(testRequest, storage, testRequest.DynamicDeps)
		So(err, ShouldBeNil)
		So(testRequest.ServiceAddress.Address, ShouldEqual, endpoint.Address)
		So(testRequest.ServiceAddress.Port, ShouldEqual, endpoint.Port)
		So(testRequest.TestRequest.TestSuites, ShouldHaveLength, len(req.Params.TestSuites))
	})
}

func TestGenericContainerOutputAsDependency(t *testing.T) {
	storage := common.NewInjectableStorage()
	devboardServer, _ := anypb.New(&labapi.IpEndpoint{
		Address: "localhost",
		Port:    12345,
	})
	genericContainerOutput := &api.GenericStopResponse{
		Message: &testapi.GenericMessage{
			Values: map[string]*anypb.Any{
				"devboard-server": devboardServer,
			},
		},
	}
	err := storage.Set("user-container_stop", genericContainerOutput)
	if err != nil {
		t.Fatalf("%s", err)
	}

	Convey("Anypb dep", t, func() {
		testRequest := &skylab_test_runner.TestRequest{
			TestRequest: &api.CrosTestRequest{
				Primary:  &api.CrosTestRequest_Device{},
				Metadata: &anypb.Any{},
			},
			DynamicDeps: []*skylab_test_runner.DynamicDep{
				{
					Key:   "testRequest.primary.devboardServer",
					Value: "user-container_stop.message.values.devboard-server",
				},
			},
		}

		err = common.InjectDependencies(testRequest, storage, testRequest.DynamicDeps)
		So(err, ShouldBeNil)
		So(testRequest.TestRequest.Primary.DevboardServer.Address, ShouldEqual, "localhost")
		So(testRequest.TestRequest.Primary.DevboardServer.Port, ShouldEqual, 12345)
	})
}
