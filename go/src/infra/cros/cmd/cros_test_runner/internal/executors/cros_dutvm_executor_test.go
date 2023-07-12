// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	vmlabapi "infra/libs/vmlab/api"
)

type mockImageApi struct {
	vmlabapi.ImageApi
	getImage func(builderPath string, wait bool) (*vmlabapi.GceImage, error)
}

func (m *mockImageApi) GetImage(builderPath string, wait bool) (*vmlabapi.GceImage, error) {
	return m.getImage(builderPath, wait)
}

func TestCrosDutVmExecutor_GetImage(t *testing.T) {
	t.Parallel()

	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmGetImageCmd {
		cmd := commands.NewDutVmGetImageCmd(exec)
		keyVals := make(map[string]string, 0)
		keyVals["build"] = "betty/R101"
		cmd.CftTestRequest = &skylab_test_runner.CFTTestRequest{
			AutotestKeyvals: keyVals,
		}
		return cmd
	}

	Convey("GetImage success", t, func() {
		expected := &vmlabapi.GceImage{Name: "image-1", Project: "project-1"}
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.ImageApi = &mockImageApi{
			getImage: func(builderPath string, wait bool) (*vmlabapi.GceImage, error) {
				So(builderPath, ShouldEqual, "betty/R101")
				So(wait, ShouldBeTrue)
				return expected, nil
			},
		}
		cmd := getCmd(exec)

		err := exec.ExecuteCommand(ctx, cmd)

		So(cmd.DutVmGceImage, ShouldResemble, expected)
		So(err, ShouldBeNil)
	})

	Convey("GetImage error", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.ImageApi = &mockImageApi{
			getImage: func(builderPath string, wait bool) (*vmlabapi.GceImage, error) {
				return nil, fmt.Errorf("vmlab lib error")
			},
		}
		cmd := getCmd(exec)

		err := exec.ExecuteCommand(ctx, cmd)

		So(cmd.DutVmGceImage, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

type mockInstanceApi struct {
	vmlabapi.InstanceApi
	create func(*vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error)
	delete func(*vmlabapi.VmInstance) error
}

func (m *mockInstanceApi) Create(ctx context.Context, req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
	return m.create(req)
}

func (m *mockInstanceApi) Delete(ctx context.Context, ins *vmlabapi.VmInstance) error {
	return m.delete(ins)
}

func TestCrosDutVmExecutor_ReleaseVm(t *testing.T) {
	t.Parallel()

	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmReleaseCmd {
		cmd := commands.NewDutVmReleaseCmd(exec)
		cmd.DutVm = &vmlabapi.VmInstance{
			Name: "instance-1",
			Config: &vmlabapi.Config{
				Backend: &vmlabapi.Config_GcloudBackend{
					GcloudBackend: &vmlabapi.Config_GCloudBackend{
						Project: "vmlab-project",
						Zone:    "us-west-2",
					},
				},
			},
		}
		return cmd
	}

	Convey("ReleaseVm success", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		cmd := getCmd(exec)
		exec.InstanceApi = &mockInstanceApi{
			delete: func(ins *vmlabapi.VmInstance) error {
				So(ins, ShouldEqual, cmd.DutVm)
				return nil
			},
		}

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldBeNil)
	})

	Convey("ReleaseVm error", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.InstanceApi = &mockInstanceApi{
			delete: func(ins *vmlabapi.VmInstance) error {
				return fmt.Errorf("vmlab lib error")
			},
		}
		cmd := getCmd(exec)

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldBeNil)
	})
}

func TestCrosDutVmExecutor_LeaseVm_GcloudBackend(t *testing.T) {
	t.Parallel()
	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmLeaseCmd {
		cmd := commands.NewDutVmLeaseCmd(exec)
		cmd.DutVmGceImage = &vmlabapi.GceImage{
			Name:    "image-1",
			Project: "project-1",
		}
		return cmd
	}

	Convey("LeaseVm success", t, func() {
		expected := &vmlabapi.VmInstance{
			Name: "instance-1",
			Ssh:  &vmlabapi.AddressPort{Address: "1.2.3.4", Port: 22},
		}
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		cmd := getCmd(exec)
		exec.InstanceApi = &mockInstanceApi{
			create: func(req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
				So(req, ShouldNotBeNil)
				So(req.GetConfig().GetGcloudBackend().GetImage(), ShouldEqual, cmd.DutVmGceImage)
				return expected, nil
			},
		}
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldBeNil)
		So(cmd.DutVm, ShouldResemble, expected)
	})

	Convey("LeaseVm error", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.InstanceApi = &mockInstanceApi{
			create: func(req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
				return nil, fmt.Errorf("vmlab lib error")
			},
		}
		cmd := getCmd(exec)
		err := exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "vmlab lib error")
	})
}

func TestCrosDutVmExecutor_LeaseVm_VmLeaserBackend(t *testing.T) {
	t.Parallel()

	Convey("LeaseVm VmLeaserBackend tests", t, func() {
		Convey("LeaseVm VmLeaserBackend - success", func() {
			ctx := context.Background()
			initialBuild := &bbpb.Build{
				Input: &bbpb.Build_Input{
					Experiments: []string{
						"chromeos.cros_infra_config.vmleaser.launch",
					},
				},
			}
			state, ctx, err := build.Start(ctx, initialBuild)
			defer func() { state.End(err) }()

			getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmLeaseCmd {
				cmd := commands.NewDutVmLeaseCmd(exec)
				cmd.DutVmGceImage = &vmlabapi.GceImage{
					Name:    "image-1",
					Project: "project-1",
				}
				cmd.BuildState = state
				return cmd
			}

			expected := &vmlabapi.VmInstance{
				Name: "instance-1",
				Ssh:  &vmlabapi.AddressPort{Address: "1.2.3.4", Port: 22},
			}
			exec := buildCrosDutVmExecutor()
			cmd := getCmd(exec)
			exec.InstanceApi = &mockInstanceApi{
				create: func(req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
					So(req, ShouldNotBeNil)
					So(req.GetConfig().GetVmLeaserBackend().GetVmRequirements().GetGceImage(), ShouldEqual, fmt.Sprintf("projects/%v/global/images/%v", cmd.DutVmGceImage.GetProject(), cmd.DutVmGceImage.GetName()))
					return expected, nil
				},
			}
			err = exec.ExecuteCommand(ctx, cmd)
			So(err, ShouldBeNil)
			So(cmd.DutVm, ShouldResemble, expected)
		})

		Convey("LeaseVm VmLeaserBackend - error", func() {
			ctx := context.Background()
			initialBuild := &bbpb.Build{
				Input: &bbpb.Build_Input{
					Experiments: []string{
						"chromeos.cros_infra_config.vmleaser.launch",
					},
				},
			}
			state, ctx, err := build.Start(ctx, initialBuild)
			defer func() { state.End(err) }()

			getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmLeaseCmd {
				cmd := commands.NewDutVmLeaseCmd(exec)
				cmd.DutVmGceImage = &vmlabapi.GceImage{
					Name:    "image-1",
					Project: "project-1",
				}
				cmd.BuildState = state
				return cmd
			}

			exec := buildCrosDutVmExecutor()
			exec.InstanceApi = &mockInstanceApi{
				create: func(req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
					return nil, fmt.Errorf("vmlab lib error")
				},
			}
			cmd := getCmd(exec)
			err = exec.ExecuteCommand(ctx, cmd)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "vmlab lib error")
		})
	})
}
func TestCrosDutVmExecutor_LeaseVm_Experiment(t *testing.T) {
	t.Parallel()

	Convey("LeaseVm experiment - different experiment success", t, func() {
		ctx := context.Background()
		initialBuild := &bbpb.Build{
			Input: &bbpb.Build_Input{
				Experiments: []string{
					"random-experiment-1",
				},
			},
		}
		state, ctx, err := build.Start(ctx, initialBuild)
		defer func() { state.End(err) }()

		getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmLeaseCmd {
			cmd := commands.NewDutVmLeaseCmd(exec)
			cmd.DutVmGceImage = &vmlabapi.GceImage{
				Name:    "image-1",
				Project: "project-1",
			}
			return cmd
		}

		expected := &vmlabapi.VmInstance{
			Name: "instance-2",
			Ssh:  &vmlabapi.AddressPort{Address: "2.2.2.2", Port: 11},
		}
		exec := buildCrosDutVmExecutor()
		cmd := getCmd(exec)
		exec.InstanceApi = &mockInstanceApi{
			create: func(req *vmlabapi.CreateVmInstanceRequest) (*vmlabapi.VmInstance, error) {
				So(req, ShouldNotBeNil)
				So(req.GetConfig().GetGcloudBackend().GetImage(), ShouldEqual, cmd.DutVmGceImage)
				return expected, nil
			},
		}
		err = exec.ExecuteCommand(ctx, cmd)
		So(err, ShouldBeNil)
		So(cmd.DutVm, ShouldResemble, expected)
	})
}

func TestCrosDutVmExecutor_StartCrosDut(t *testing.T) {
	t.Parallel()

	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutServiceStartCmd {
		cmd := commands.NewDutServiceStartCmd(exec)
		cmd.CacheServerAddress = &labapi.IpEndpoint{Address: "4.3.2.1", Port: 8080}
		cmd.DutSshAddress = &labapi.IpEndpoint{Address: "1.2.3.4", Port: 22}
		return cmd
	}

	Convey("StartCrosDut success", t, func() {
		expected := &labapi.IpEndpoint{Address: "localhost", Port: 44355}
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		cmd := getCmd(exec)
		exec.Container = &mockContainerApi{
			process: func(ctx context.Context, template *api.Template) (string, error) {
				So(template, ShouldNotBeNil)
				return "localhost:44355", nil
			},
		}

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldBeNil)
		So(cmd.DutServerAddress, ShouldResemble, expected)
	})

	Convey("StartCrosDut error", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.Container = &mockContainerApi{
			process: func(ctx context.Context, template *api.Template) (string, error) {
				return "", fmt.Errorf("ctr error")
			},
		}
		cmd := getCmd(exec)

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldNotBeNil)
	})
}

type mockContainerApi struct {
	interfaces.ContainerInterface
	process func(context.Context, *api.Template) (string, error)
}

func (m *mockContainerApi) ProcessContainer(ctx context.Context, t *api.Template) (string, error) {
	return m.process(ctx, t)
}

func buildCrosDutVmExecutor() *CrosDutVmExecutor {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	cont := containers.NewCrosDutTemplatedContainer("container/image/path", ctr)
	exec := NewCrosDutVmExecutor(cont)
	return exec
}
