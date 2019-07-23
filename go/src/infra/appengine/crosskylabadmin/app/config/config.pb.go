// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/crosskylabadmin/app/config/config.proto

package config

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Config is the configuration data served by luci-config for this app.
type Config struct {
	// AccessGroup is the luci-auth group controlling access to admin app APIs.
	AccessGroup string `protobuf:"bytes,1,opt,name=access_group,json=accessGroup,proto3" json:"access_group,omitempty"`
	// Swarming contains information about the Swarming instance that hosts the
	// bots managed by this app.
	Swarming *Swarming `protobuf:"bytes,2,opt,name=swarming,proto3" json:"swarming,omitempty"`
	// Tasker contains configuration data specific to the Tasker API endpoints.
	Tasker *Tasker `protobuf:"bytes,3,opt,name=tasker,proto3" json:"tasker,omitempty"`
	// Cron contains the configuration data specific to cron jobs on this app.
	Cron *Cron `protobuf:"bytes,4,opt,name=cron,proto3" json:"cron,omitempty"`
	// Inventory contains configuration information about skylab inventory
	// repo.
	Inventory *Inventory `protobuf:"bytes,5,opt,name=inventory,proto3" json:"inventory,omitempty"`
	// endpoint contains configuration of specific API endpoints.
	Endpoint *Endpoint `protobuf:"bytes,6,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	// RPCcontrol controls rpc traffic.
	RpcControl           *RPCControl `protobuf:"bytes,7,opt,name=rpc_control,json=rpcControl,proto3" json:"rpc_control,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{0}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetAccessGroup() string {
	if m != nil {
		return m.AccessGroup
	}
	return ""
}

func (m *Config) GetSwarming() *Swarming {
	if m != nil {
		return m.Swarming
	}
	return nil
}

func (m *Config) GetTasker() *Tasker {
	if m != nil {
		return m.Tasker
	}
	return nil
}

func (m *Config) GetCron() *Cron {
	if m != nil {
		return m.Cron
	}
	return nil
}

func (m *Config) GetInventory() *Inventory {
	if m != nil {
		return m.Inventory
	}
	return nil
}

func (m *Config) GetEndpoint() *Endpoint {
	if m != nil {
		return m.Endpoint
	}
	return nil
}

func (m *Config) GetRpcControl() *RPCControl {
	if m != nil {
		return m.RpcControl
	}
	return nil
}

// Swarming contains information about the Swarming instance that hosts the bots
// managed by this app.
type Swarming struct {
	// Host is the swarming instance hosting skylab bots.
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// BotPool is the swarming pool containing skylab bots.
	BotPool string `protobuf:"bytes,2,opt,name=bot_pool,json=botPool,proto3" json:"bot_pool,omitempty"`
	// FleetAdminTaskTag identifies all tasks created by the fleet admin app.
	FleetAdminTaskTag string `protobuf:"bytes,3,opt,name=fleet_admin_task_tag,json=fleetAdminTaskTag,proto3" json:"fleet_admin_task_tag,omitempty"`
	// LuciProjectTag is the swarming tag that associates the task with a
	// luci project, allowing milo to work with the swarming UI.
	LuciProjectTag       string   `protobuf:"bytes,4,opt,name=luci_project_tag,json=luciProjectTag,proto3" json:"luci_project_tag,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Swarming) Reset()         { *m = Swarming{} }
func (m *Swarming) String() string { return proto.CompactTextString(m) }
func (*Swarming) ProtoMessage()    {}
func (*Swarming) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{1}
}

func (m *Swarming) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Swarming.Unmarshal(m, b)
}
func (m *Swarming) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Swarming.Marshal(b, m, deterministic)
}
func (m *Swarming) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Swarming.Merge(m, src)
}
func (m *Swarming) XXX_Size() int {
	return xxx_messageInfo_Swarming.Size(m)
}
func (m *Swarming) XXX_DiscardUnknown() {
	xxx_messageInfo_Swarming.DiscardUnknown(m)
}

var xxx_messageInfo_Swarming proto.InternalMessageInfo

func (m *Swarming) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *Swarming) GetBotPool() string {
	if m != nil {
		return m.BotPool
	}
	return ""
}

func (m *Swarming) GetFleetAdminTaskTag() string {
	if m != nil {
		return m.FleetAdminTaskTag
	}
	return ""
}

func (m *Swarming) GetLuciProjectTag() string {
	if m != nil {
		return m.LuciProjectTag
	}
	return ""
}

// Tasker contains configuration data specific to the Tasker API endpoints.
type Tasker struct {
	// BackgroundTaskExecutionTimeoutSecs is the execution timeout (in
	// seconds) for background tasks created by tasker.
	BackgroundTaskExecutionTimeoutSecs int64 `protobuf:"varint,1,opt,name=background_task_execution_timeout_secs,json=backgroundTaskExecutionTimeoutSecs,proto3" json:"background_task_execution_timeout_secs,omitempty"`
	// BackgroundTaskExpirationSecs is the expiration time (in seconds) for
	// background tasks created by tasker.
	BackgroundTaskExpirationSecs int64 `protobuf:"varint,2,opt,name=background_task_expiration_secs,json=backgroundTaskExpirationSecs,proto3" json:"background_task_expiration_secs,omitempty"`
	// LogdogHost is the Logdog host to use for logging from the created tasks.
	LogdogHost string `protobuf:"bytes,3,opt,name=logdog_host,json=logdogHost,proto3" json:"logdog_host,omitempty"`
	// AdminTaskServiceAccount is the name of the service account to use for admin
	// tasks.
	AdminTaskServiceAccount string   `protobuf:"bytes,4,opt,name=admin_task_service_account,json=adminTaskServiceAccount,proto3" json:"admin_task_service_account,omitempty"`
	XXX_NoUnkeyedLiteral    struct{} `json:"-"`
	XXX_unrecognized        []byte   `json:"-"`
	XXX_sizecache           int32    `json:"-"`
}

func (m *Tasker) Reset()         { *m = Tasker{} }
func (m *Tasker) String() string { return proto.CompactTextString(m) }
func (*Tasker) ProtoMessage()    {}
func (*Tasker) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{2}
}

func (m *Tasker) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Tasker.Unmarshal(m, b)
}
func (m *Tasker) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Tasker.Marshal(b, m, deterministic)
}
func (m *Tasker) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Tasker.Merge(m, src)
}
func (m *Tasker) XXX_Size() int {
	return xxx_messageInfo_Tasker.Size(m)
}
func (m *Tasker) XXX_DiscardUnknown() {
	xxx_messageInfo_Tasker.DiscardUnknown(m)
}

var xxx_messageInfo_Tasker proto.InternalMessageInfo

func (m *Tasker) GetBackgroundTaskExecutionTimeoutSecs() int64 {
	if m != nil {
		return m.BackgroundTaskExecutionTimeoutSecs
	}
	return 0
}

func (m *Tasker) GetBackgroundTaskExpirationSecs() int64 {
	if m != nil {
		return m.BackgroundTaskExpirationSecs
	}
	return 0
}

func (m *Tasker) GetLogdogHost() string {
	if m != nil {
		return m.LogdogHost
	}
	return ""
}

func (m *Tasker) GetAdminTaskServiceAccount() string {
	if m != nil {
		return m.AdminTaskServiceAccount
	}
	return ""
}

// Cron contains the configuration data specific to cron jobs on this app.
type Cron struct {
	// FleetAdminTaskPriority is the swarming task priority of created tasks.
	//
	// This must be numerically smaller (i.e. more important) than Skylab's test
	// task priority range [49-255] and numerically larger than the minimum
	// allowed Swarming priority (20) for non administrator users.
	FleetAdminTaskPriority int64 `protobuf:"varint,1,opt,name=fleet_admin_task_priority,json=fleetAdminTaskPriority,proto3" json:"fleet_admin_task_priority,omitempty"`
	// EnsureTasksCount is the number of background tasks maintained against
	// each bot.
	EnsureTasksCount int32 `protobuf:"varint,2,opt,name=ensure_tasks_count,json=ensureTasksCount,proto3" json:"ensure_tasks_count,omitempty"`
	// RepairIdleDuration is the duration for which a bot in the fleet must have
	// been idle for a repair task to be created against it.
	RepairIdleDuration *duration.Duration `protobuf:"bytes,3,opt,name=repair_idle_duration,json=repairIdleDuration,proto3" json:"repair_idle_duration,omitempty"`
	// RepairAttemptDelayDuration is the time between successive attempts at
	// repairing repair failed bots in the fleet.
	RepairAttemptDelayDuration *duration.Duration `protobuf:"bytes,4,opt,name=repair_attempt_delay_duration,json=repairAttemptDelayDuration,proto3" json:"repair_attempt_delay_duration,omitempty"`
	// Configuration of automatic pool balancing to keep critical pools healthy.
	PoolBalancer         *PoolBalancer `protobuf:"bytes,5,opt,name=pool_balancer,json=poolBalancer,proto3" json:"pool_balancer,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Cron) Reset()         { *m = Cron{} }
func (m *Cron) String() string { return proto.CompactTextString(m) }
func (*Cron) ProtoMessage()    {}
func (*Cron) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{3}
}

func (m *Cron) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Cron.Unmarshal(m, b)
}
func (m *Cron) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Cron.Marshal(b, m, deterministic)
}
func (m *Cron) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Cron.Merge(m, src)
}
func (m *Cron) XXX_Size() int {
	return xxx_messageInfo_Cron.Size(m)
}
func (m *Cron) XXX_DiscardUnknown() {
	xxx_messageInfo_Cron.DiscardUnknown(m)
}

var xxx_messageInfo_Cron proto.InternalMessageInfo

func (m *Cron) GetFleetAdminTaskPriority() int64 {
	if m != nil {
		return m.FleetAdminTaskPriority
	}
	return 0
}

func (m *Cron) GetEnsureTasksCount() int32 {
	if m != nil {
		return m.EnsureTasksCount
	}
	return 0
}

func (m *Cron) GetRepairIdleDuration() *duration.Duration {
	if m != nil {
		return m.RepairIdleDuration
	}
	return nil
}

func (m *Cron) GetRepairAttemptDelayDuration() *duration.Duration {
	if m != nil {
		return m.RepairAttemptDelayDuration
	}
	return nil
}

func (m *Cron) GetPoolBalancer() *PoolBalancer {
	if m != nil {
		return m.PoolBalancer
	}
	return nil
}

type RPCControl struct {
	// Configuration of if disabling some rpc calls. It's used in experimental stage.
	// Once an RPC call is verified to be working/useless, it will be added/deleted.
	DisableEnsureBackgroundTasks       bool     `protobuf:"varint,1,opt,name=disable_ensure_background_tasks,json=disableEnsureBackgroundTasks,proto3" json:"disable_ensure_background_tasks,omitempty"`
	DisableEnsureCriticalPoolsHealthy  bool     `protobuf:"varint,2,opt,name=disable_ensure_critical_pools_healthy,json=disableEnsureCriticalPoolsHealthy,proto3" json:"disable_ensure_critical_pools_healthy,omitempty"`
	DisablePushBotsForAdminTasks       bool     `protobuf:"varint,3,opt,name=disable_push_bots_for_admin_tasks,json=disablePushBotsForAdminTasks,proto3" json:"disable_push_bots_for_admin_tasks,omitempty"`
	DisableRefreshBots                 bool     `protobuf:"varint,4,opt,name=disable_refresh_bots,json=disableRefreshBots,proto3" json:"disable_refresh_bots,omitempty"`
	DisableRefreshInventory            bool     `protobuf:"varint,5,opt,name=disable_refresh_inventory,json=disableRefreshInventory,proto3" json:"disable_refresh_inventory,omitempty"`
	DisableTriggerRepairOnIdle         bool     `protobuf:"varint,6,opt,name=disable_trigger_repair_on_idle,json=disableTriggerRepairOnIdle,proto3" json:"disable_trigger_repair_on_idle,omitempty"`
	DisableTriggerRepairOnRepairFailed bool     `protobuf:"varint,7,opt,name=disable_trigger_repair_on_repair_failed,json=disableTriggerRepairOnRepairFailed,proto3" json:"disable_trigger_repair_on_repair_failed,omitempty"`
	XXX_NoUnkeyedLiteral               struct{} `json:"-"`
	XXX_unrecognized                   []byte   `json:"-"`
	XXX_sizecache                      int32    `json:"-"`
}

func (m *RPCControl) Reset()         { *m = RPCControl{} }
func (m *RPCControl) String() string { return proto.CompactTextString(m) }
func (*RPCControl) ProtoMessage()    {}
func (*RPCControl) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{4}
}

func (m *RPCControl) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RPCControl.Unmarshal(m, b)
}
func (m *RPCControl) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RPCControl.Marshal(b, m, deterministic)
}
func (m *RPCControl) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RPCControl.Merge(m, src)
}
func (m *RPCControl) XXX_Size() int {
	return xxx_messageInfo_RPCControl.Size(m)
}
func (m *RPCControl) XXX_DiscardUnknown() {
	xxx_messageInfo_RPCControl.DiscardUnknown(m)
}

var xxx_messageInfo_RPCControl proto.InternalMessageInfo

func (m *RPCControl) GetDisableEnsureBackgroundTasks() bool {
	if m != nil {
		return m.DisableEnsureBackgroundTasks
	}
	return false
}

func (m *RPCControl) GetDisableEnsureCriticalPoolsHealthy() bool {
	if m != nil {
		return m.DisableEnsureCriticalPoolsHealthy
	}
	return false
}

func (m *RPCControl) GetDisablePushBotsForAdminTasks() bool {
	if m != nil {
		return m.DisablePushBotsForAdminTasks
	}
	return false
}

func (m *RPCControl) GetDisableRefreshBots() bool {
	if m != nil {
		return m.DisableRefreshBots
	}
	return false
}

func (m *RPCControl) GetDisableRefreshInventory() bool {
	if m != nil {
		return m.DisableRefreshInventory
	}
	return false
}

func (m *RPCControl) GetDisableTriggerRepairOnIdle() bool {
	if m != nil {
		return m.DisableTriggerRepairOnIdle
	}
	return false
}

func (m *RPCControl) GetDisableTriggerRepairOnRepairFailed() bool {
	if m != nil {
		return m.DisableTriggerRepairOnRepairFailed
	}
	return false
}

// Skylab inventory is stored in a git project. A Gitiles server as well as
// Gerrit review server are used by this app to view and update the inventory
// data.
type Inventory struct {
	// Gitiles server hosting inventory project.
	// e.g. chromium.googlesource.com
	GitilesHost string `protobuf:"bytes,1,opt,name=gitiles_host,json=gitilesHost,proto3" json:"gitiles_host,omitempty"`
	// Gerrit code review server hosting inventory project.
	// e.g. chromium-review.googlesource.com
	GerritHost string `protobuf:"bytes,2,opt,name=gerrit_host,json=gerritHost,proto3" json:"gerrit_host,omitempty"`
	// Git project containing the inventory data.
	Project string `protobuf:"bytes,3,opt,name=project,proto3" json:"project,omitempty"`
	// Git branch from the inventory project to be used.
	Branch   string `protobuf:"bytes,4,opt,name=branch,proto3" json:"branch,omitempty"`
	DataPath string `protobuf:"bytes,5,opt,name=data_path,json=dataPath,proto3" json:"data_path,omitempty"` // Deprecated: Do not use.
	// Inventory environment managed by this instance of the app.
	// e.g. ENVIRONMENT_STAGING
	Environment string `protobuf:"bytes,6,opt,name=environment,proto3" json:"environment,omitempty"`
	// Path to the infrastructure inventory data file within the git project.
	// e.g. data/skylab/server_db.textpb
	InfrastructureDataPath string `protobuf:"bytes,7,opt,name=infrastructure_data_path,json=infrastructureDataPath,proto3" json:"infrastructure_data_path,omitempty"`
	// Path to the lab inventory data file within the git project.
	// e.g. data/skylab/lab.textpb
	LabDataPath string `protobuf:"bytes,8,opt,name=lab_data_path,json=labDataPath,proto3" json:"lab_data_path,omitempty"`
	// dut_info_cache_validty is the amount of time cached inventory information
	// about a DUT is valid after being refreshed.
	//
	// This duration should be long enough to
	// (1) smooth over any refresh failures due to backing gitiles flake or quota
	//     issues.
	// (2) Allow a human to interfere and fix corrupt inventory data about (some)
	//     DUTs.
	//
	// A DUT will continue to live in the cache (and hence be served via various
	// RPCs) for dut_info_cache_validity after it has been deleted from the
	// inventory.
	DutInfoCacheValidity *duration.Duration `protobuf:"bytes,9,opt,name=dut_info_cache_validity,json=dutInfoCacheValidity,proto3" json:"dut_info_cache_validity,omitempty"`
	// update_limit_per_minute is used to rate limit some inventory updates.
	UpdateLimitPerMinute int32 `protobuf:"varint,10,opt,name=update_limit_per_minute,json=updateLimitPerMinute,proto3" json:"update_limit_per_minute,omitempty"`
	// Queen service to push inventory to.
	QueenService string `protobuf:"bytes,11,opt,name=queen_service,json=queenService,proto3" json:"queen_service,omitempty"`
	// Git project containing the device config.
	DeviceConfigProject string `protobuf:"bytes,12,opt,name=device_config_project,json=deviceConfigProject,proto3" json:"device_config_project,omitempty"`
	// Git branch from the device config project to be used.
	DeviceConfigBranch string `protobuf:"bytes,13,opt,name=device_config_branch,json=deviceConfigBranch,proto3" json:"device_config_branch,omitempty"`
	// The device config file path.
	// e.g. deviceconfig/generated/device_configs.cfg
	DeviceConfigPath     string   `protobuf:"bytes,14,opt,name=device_config_path,json=deviceConfigPath,proto3" json:"device_config_path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Inventory) Reset()         { *m = Inventory{} }
func (m *Inventory) String() string { return proto.CompactTextString(m) }
func (*Inventory) ProtoMessage()    {}
func (*Inventory) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{5}
}

func (m *Inventory) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Inventory.Unmarshal(m, b)
}
func (m *Inventory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Inventory.Marshal(b, m, deterministic)
}
func (m *Inventory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Inventory.Merge(m, src)
}
func (m *Inventory) XXX_Size() int {
	return xxx_messageInfo_Inventory.Size(m)
}
func (m *Inventory) XXX_DiscardUnknown() {
	xxx_messageInfo_Inventory.DiscardUnknown(m)
}

var xxx_messageInfo_Inventory proto.InternalMessageInfo

func (m *Inventory) GetGitilesHost() string {
	if m != nil {
		return m.GitilesHost
	}
	return ""
}

func (m *Inventory) GetGerritHost() string {
	if m != nil {
		return m.GerritHost
	}
	return ""
}

func (m *Inventory) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *Inventory) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

// Deprecated: Do not use.
func (m *Inventory) GetDataPath() string {
	if m != nil {
		return m.DataPath
	}
	return ""
}

func (m *Inventory) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

func (m *Inventory) GetInfrastructureDataPath() string {
	if m != nil {
		return m.InfrastructureDataPath
	}
	return ""
}

func (m *Inventory) GetLabDataPath() string {
	if m != nil {
		return m.LabDataPath
	}
	return ""
}

func (m *Inventory) GetDutInfoCacheValidity() *duration.Duration {
	if m != nil {
		return m.DutInfoCacheValidity
	}
	return nil
}

func (m *Inventory) GetUpdateLimitPerMinute() int32 {
	if m != nil {
		return m.UpdateLimitPerMinute
	}
	return 0
}

func (m *Inventory) GetQueenService() string {
	if m != nil {
		return m.QueenService
	}
	return ""
}

func (m *Inventory) GetDeviceConfigProject() string {
	if m != nil {
		return m.DeviceConfigProject
	}
	return ""
}

func (m *Inventory) GetDeviceConfigBranch() string {
	if m != nil {
		return m.DeviceConfigBranch
	}
	return ""
}

func (m *Inventory) GetDeviceConfigPath() string {
	if m != nil {
		return m.DeviceConfigPath
	}
	return ""
}

type PoolBalancer struct {
	// Names of the pools to keep healthy automatically via pool balancing.
	TargetPools []string `protobuf:"bytes,1,rep,name=target_pools,json=targetPools,proto3" json:"target_pools,omitempty"`
	// Name of the pool to use as the spare pool for pool balancing.
	SparePool string `protobuf:"bytes,2,opt,name=spare_pool,json=sparePool,proto3" json:"spare_pool,omitempty"`
	// Maximum number of unhealthy DUTs per model that can be balanced away from
	// a single target pool.
	MaxUnhealthyDuts     int32    `protobuf:"varint,3,opt,name=max_unhealthy_duts,json=maxUnhealthyDuts,proto3" json:"max_unhealthy_duts,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PoolBalancer) Reset()         { *m = PoolBalancer{} }
func (m *PoolBalancer) String() string { return proto.CompactTextString(m) }
func (*PoolBalancer) ProtoMessage()    {}
func (*PoolBalancer) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{6}
}

func (m *PoolBalancer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PoolBalancer.Unmarshal(m, b)
}
func (m *PoolBalancer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PoolBalancer.Marshal(b, m, deterministic)
}
func (m *PoolBalancer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoolBalancer.Merge(m, src)
}
func (m *PoolBalancer) XXX_Size() int {
	return xxx_messageInfo_PoolBalancer.Size(m)
}
func (m *PoolBalancer) XXX_DiscardUnknown() {
	xxx_messageInfo_PoolBalancer.DiscardUnknown(m)
}

var xxx_messageInfo_PoolBalancer proto.InternalMessageInfo

func (m *PoolBalancer) GetTargetPools() []string {
	if m != nil {
		return m.TargetPools
	}
	return nil
}

func (m *PoolBalancer) GetSparePool() string {
	if m != nil {
		return m.SparePool
	}
	return ""
}

func (m *PoolBalancer) GetMaxUnhealthyDuts() int32 {
	if m != nil {
		return m.MaxUnhealthyDuts
	}
	return 0
}

type Endpoint struct {
	DeployDut            *DeployDut `protobuf:"bytes,1,opt,name=deploy_dut,json=deployDut,proto3" json:"deploy_dut,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Endpoint) Reset()         { *m = Endpoint{} }
func (m *Endpoint) String() string { return proto.CompactTextString(m) }
func (*Endpoint) ProtoMessage()    {}
func (*Endpoint) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{7}
}

func (m *Endpoint) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Endpoint.Unmarshal(m, b)
}
func (m *Endpoint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Endpoint.Marshal(b, m, deterministic)
}
func (m *Endpoint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Endpoint.Merge(m, src)
}
func (m *Endpoint) XXX_Size() int {
	return xxx_messageInfo_Endpoint.Size(m)
}
func (m *Endpoint) XXX_DiscardUnknown() {
	xxx_messageInfo_Endpoint.DiscardUnknown(m)
}

var xxx_messageInfo_Endpoint proto.InternalMessageInfo

func (m *Endpoint) GetDeployDut() *DeployDut {
	if m != nil {
		return m.DeployDut
	}
	return nil
}

type DeployDut struct {
	// Amount of time the deploy Skylab task can be PENDING.
	//
	// This should be long enough for the newly updated inventory information to
	// propagate to the Swarming bots.
	TaskExpirationTimeout *duration.Duration `protobuf:"bytes,1,opt,name=task_expiration_timeout,json=taskExpirationTimeout,proto3" json:"task_expiration_timeout,omitempty"`
	// Amount of time the deploy Skylab task is allowed to run.
	//
	// This should be enough for possibly installing firmware and test image on
	// the DUT.
	TaskExecutionTimeout *duration.Duration `protobuf:"bytes,2,opt,name=task_execution_timeout,json=taskExecutionTimeout,proto3" json:"task_execution_timeout,omitempty"`
	// Priority of the deploy Skylab task.
	//
	// This should be the same as, or higher priority (i.e., numerically lower)
	// than other admin tasks.
	TaskPriority         int64    `protobuf:"varint,3,opt,name=task_priority,json=taskPriority,proto3" json:"task_priority,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeployDut) Reset()         { *m = DeployDut{} }
func (m *DeployDut) String() string { return proto.CompactTextString(m) }
func (*DeployDut) ProtoMessage()    {}
func (*DeployDut) Descriptor() ([]byte, []int) {
	return fileDescriptor_d85dab011e4afc14, []int{8}
}

func (m *DeployDut) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeployDut.Unmarshal(m, b)
}
func (m *DeployDut) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeployDut.Marshal(b, m, deterministic)
}
func (m *DeployDut) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeployDut.Merge(m, src)
}
func (m *DeployDut) XXX_Size() int {
	return xxx_messageInfo_DeployDut.Size(m)
}
func (m *DeployDut) XXX_DiscardUnknown() {
	xxx_messageInfo_DeployDut.DiscardUnknown(m)
}

var xxx_messageInfo_DeployDut proto.InternalMessageInfo

func (m *DeployDut) GetTaskExpirationTimeout() *duration.Duration {
	if m != nil {
		return m.TaskExpirationTimeout
	}
	return nil
}

func (m *DeployDut) GetTaskExecutionTimeout() *duration.Duration {
	if m != nil {
		return m.TaskExecutionTimeout
	}
	return nil
}

func (m *DeployDut) GetTaskPriority() int64 {
	if m != nil {
		return m.TaskPriority
	}
	return 0
}

func init() {
	proto.RegisterType((*Config)(nil), "crosskylabadmin.config.Config")
	proto.RegisterType((*Swarming)(nil), "crosskylabadmin.config.Swarming")
	proto.RegisterType((*Tasker)(nil), "crosskylabadmin.config.Tasker")
	proto.RegisterType((*Cron)(nil), "crosskylabadmin.config.Cron")
	proto.RegisterType((*RPCControl)(nil), "crosskylabadmin.config.RPCControl")
	proto.RegisterType((*Inventory)(nil), "crosskylabadmin.config.Inventory")
	proto.RegisterType((*PoolBalancer)(nil), "crosskylabadmin.config.PoolBalancer")
	proto.RegisterType((*Endpoint)(nil), "crosskylabadmin.config.Endpoint")
	proto.RegisterType((*DeployDut)(nil), "crosskylabadmin.config.DeployDut")
}

func init() {
	proto.RegisterFile("infra/appengine/crosskylabadmin/app/config/config.proto", fileDescriptor_d85dab011e4afc14)
}

var fileDescriptor_d85dab011e4afc14 = []byte{
	// 1248 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x56, 0x5b, 0x6f, 0xdb, 0xb6,
	0x17, 0x47, 0x12, 0xd7, 0x91, 0x8f, 0x9d, 0x22, 0x7f, 0xfe, 0xd3, 0x44, 0x09, 0xda, 0x26, 0xd1,
	0x6e, 0x7d, 0x28, 0x9c, 0xa2, 0xc3, 0xee, 0x03, 0xb6, 0xda, 0x49, 0xdb, 0x60, 0x1d, 0xea, 0x31,
	0xd9, 0x1e, 0x86, 0x01, 0x04, 0x2d, 0xd1, 0x32, 0x57, 0x99, 0xd4, 0x48, 0x2a, 0x6b, 0x5e, 0x06,
	0xec, 0x3b, 0x0c, 0xd8, 0x17, 0xdb, 0x47, 0xd8, 0x3e, 0xc2, 0xde, 0x07, 0x5e, 0xe4, 0x5b, 0x93,
	0xf4, 0x49, 0xd2, 0x39, 0xbf, 0xdf, 0x8f, 0xe4, 0xe1, 0xb9, 0x08, 0x3e, 0xe1, 0x62, 0xa4, 0xe8,
	0x11, 0x2d, 0x4b, 0x26, 0x72, 0x2e, 0xd8, 0x51, 0xaa, 0xa4, 0xd6, 0xaf, 0x2e, 0x0b, 0x3a, 0xa4,
	0xd9, 0x84, 0x0b, 0xeb, 0x39, 0x4a, 0xa5, 0x18, 0xf1, 0x3c, 0x3c, 0xba, 0xa5, 0x92, 0x46, 0xa2,
	0xed, 0x25, 0x60, 0xd7, 0x7b, 0xf7, 0xee, 0xe7, 0x52, 0xe6, 0x05, 0x3b, 0x72, 0xa8, 0x61, 0x35,
	0x3a, 0xca, 0x2a, 0x45, 0x0d, 0x97, 0xc2, 0xf3, 0x92, 0x3f, 0xd7, 0xa0, 0xd9, 0x77, 0x50, 0x74,
	0x08, 0x1d, 0x9a, 0xa6, 0x4c, 0x6b, 0x92, 0x2b, 0x59, 0x95, 0xf1, 0xca, 0xc1, 0xca, 0x83, 0x16,
	0x6e, 0x7b, 0xdb, 0x33, 0x6b, 0x42, 0x5f, 0x42, 0xa4, 0x7f, 0xa5, 0x6a, 0xc2, 0x45, 0x1e, 0xaf,
	0x1e, 0xac, 0x3c, 0x68, 0x3f, 0x3e, 0xe8, 0x5e, 0xbd, 0x70, 0xf7, 0x2c, 0xe0, 0xf0, 0x94, 0x81,
	0x3e, 0x86, 0xa6, 0xa1, 0xfa, 0x15, 0x53, 0xf1, 0x9a, 0xe3, 0xde, 0xbf, 0x8e, 0x7b, 0xee, 0x50,
	0x38, 0xa0, 0xd1, 0x23, 0x68, 0xa4, 0x4a, 0x8a, 0xb8, 0xe1, 0x58, 0x77, 0xaf, 0x63, 0xf5, 0x95,
	0x14, 0xd8, 0x21, 0xd1, 0x57, 0xd0, 0xe2, 0xe2, 0x82, 0x09, 0x23, 0xd5, 0x65, 0x7c, 0xcb, 0xd1,
	0x0e, 0xaf, 0xa3, 0x9d, 0xd6, 0x40, 0x3c, 0xe3, 0xd8, 0x83, 0x32, 0x91, 0x95, 0x92, 0x0b, 0x13,
	0x37, 0x6f, 0x3e, 0xe8, 0x49, 0xc0, 0xe1, 0x29, 0x03, 0xf5, 0xa1, 0xad, 0xca, 0x94, 0xa4, 0x52,
	0x18, 0x25, 0x8b, 0x78, 0xdd, 0x09, 0x24, 0xd7, 0x09, 0xe0, 0x41, 0xbf, 0xef, 0x91, 0x18, 0x54,
	0x99, 0x86, 0xf7, 0xe4, 0x8f, 0x15, 0x88, 0xea, 0x20, 0x22, 0x04, 0x8d, 0xb1, 0xd4, 0x26, 0xdc,
	0x89, 0x7b, 0x47, 0xbb, 0x10, 0x0d, 0xa5, 0x21, 0xa5, 0x94, 0x85, 0xbb, 0x8c, 0x16, 0x5e, 0x1f,
	0x4a, 0x33, 0x90, 0xb2, 0x40, 0x47, 0xb0, 0x35, 0x2a, 0x18, 0x33, 0xc4, 0x2d, 0x44, 0x6c, 0x1c,
	0x89, 0xa1, 0xb9, 0x8b, 0x7b, 0x0b, 0xff, 0xcf, 0xf9, 0x9e, 0x58, 0x97, 0x8d, 0xf4, 0x39, 0xcd,
	0xd1, 0x03, 0xd8, 0x2c, 0xaa, 0x94, 0x93, 0x52, 0xc9, 0x9f, 0x59, 0x6a, 0x1c, 0xb8, 0xe1, 0xc0,
	0xb7, 0xad, 0x7d, 0xe0, 0xcd, 0xe7, 0x34, 0x4f, 0x7e, 0x5f, 0x85, 0xa6, 0xbf, 0x1f, 0x84, 0xe1,
	0xfd, 0x21, 0x4d, 0x5f, 0xd9, 0x6c, 0x11, 0x99, 0x5f, 0x84, 0xbd, 0x66, 0x69, 0x65, 0xd3, 0x8b,
	0x18, 0x3e, 0x61, 0xb2, 0x32, 0x44, 0xb3, 0x54, 0xbb, 0x6d, 0xaf, 0xe1, 0x64, 0x86, 0xb6, 0x0a,
	0x27, 0x35, 0xf6, 0xdc, 0x43, 0xcf, 0x58, 0xaa, 0xd1, 0x09, 0xec, 0xbf, 0xa9, 0x59, 0x72, 0x9f,
	0xb3, 0x5e, 0x6c, 0xd5, 0x89, 0xdd, 0x5d, 0x16, 0xab, 0x41, 0x4e, 0x66, 0x1f, 0xda, 0x85, 0xcc,
	0x33, 0x99, 0x13, 0x17, 0x36, 0x7f, 0x6e, 0xf0, 0xa6, 0xe7, 0x36, 0x78, 0x5f, 0xc0, 0xde, 0x5c,
	0x6c, 0x34, 0x53, 0x17, 0x3c, 0x65, 0x84, 0xa6, 0xa9, 0xac, 0x84, 0x09, 0x47, 0xdf, 0xa1, 0x75,
	0x88, 0xce, 0xbc, 0xff, 0x89, 0x77, 0x27, 0xff, 0xac, 0x42, 0xc3, 0x66, 0x1b, 0xfa, 0x0c, 0x76,
	0xdf, 0x88, 0x73, 0xa9, 0xb8, 0x54, 0xdc, 0x5c, 0x86, 0x43, 0x6f, 0x2f, 0x06, 0x7b, 0x10, 0xbc,
	0xe8, 0x21, 0x20, 0x26, 0x74, 0xa5, 0x98, 0x63, 0x69, 0xe2, 0x17, 0xb6, 0x67, 0xbb, 0x85, 0x37,
	0xbd, 0xc7, 0xe2, 0x75, 0xdf, 0xda, 0xd1, 0x37, 0xb0, 0xa5, 0x58, 0x49, 0xb9, 0x22, 0x3c, 0x2b,
	0x18, 0xa9, 0x8b, 0x38, 0x14, 0xd2, 0x6e, 0xd7, 0x57, 0x79, 0xb7, 0xae, 0xf2, 0xee, 0x71, 0x00,
	0x60, 0xe4, 0x69, 0xa7, 0x59, 0xc1, 0x6a, 0x1b, 0xfa, 0x09, 0xee, 0x05, 0x31, 0x6a, 0x0c, 0x9b,
	0x94, 0x86, 0x64, 0xac, 0xa0, 0x97, 0x33, 0xd5, 0xc6, 0xdb, 0x54, 0xf7, 0x3c, 0xff, 0x89, 0xa7,
	0x1f, 0x5b, 0xf6, 0x54, 0xfd, 0x14, 0x36, 0x6c, 0x4a, 0x92, 0x21, 0x2d, 0xa8, 0x48, 0x99, 0x0a,
	0xf5, 0xf7, 0xee, 0x75, 0xe9, 0x6f, 0x13, 0xb6, 0x17, 0xb0, 0xb8, 0x53, 0xce, 0x7d, 0x25, 0xff,
	0xae, 0x01, 0xcc, 0xaa, 0xc3, 0xe6, 0x46, 0xc6, 0x35, 0x1d, 0x16, 0x8c, 0x84, 0xd0, 0x2d, 0xa5,
	0x8a, 0x4f, 0xb4, 0x08, 0xdf, 0x0d, 0xb0, 0x13, 0x87, 0xea, 0x2d, 0x24, 0x8a, 0x46, 0x03, 0x78,
	0x6f, 0x49, 0x26, 0x55, 0xdc, 0xf0, 0x94, 0x16, 0xae, 0x96, 0x34, 0x19, 0x33, 0x5a, 0x98, 0xf1,
	0xa5, 0xbb, 0x8c, 0x08, 0x1f, 0x2e, 0x88, 0xf5, 0x03, 0xd4, 0xee, 0x5a, 0x3f, 0xf7, 0x40, 0xf4,
	0x0c, 0x6a, 0x10, 0x29, 0x2b, 0x3d, 0x26, 0x43, 0x69, 0x34, 0x19, 0x49, 0x35, 0x97, 0x16, 0xda,
	0x5d, 0xd5, 0x6c, 0x6b, 0x83, 0x4a, 0x8f, 0x7b, 0xd2, 0xe8, 0xa7, 0x52, 0x4d, 0x73, 0x43, 0xa3,
	0x47, 0xb0, 0x55, 0x0b, 0x29, 0x36, 0x52, 0x2c, 0x68, 0xb9, 0x0b, 0x89, 0x30, 0x0a, 0x3e, 0xec,
	0x5d, 0x96, 0x8e, 0x3e, 0x87, 0xdd, 0x65, 0xc6, 0x62, 0xe7, 0x8b, 0xf0, 0xce, 0x22, 0x6d, 0xda,
	0xef, 0x50, 0x0f, 0xee, 0xd7, 0x5c, 0xa3, 0x78, 0x9e, 0x33, 0x45, 0x42, 0x5e, 0x48, 0xe1, 0xf2,
	0xcc, 0xb5, 0xbe, 0x08, 0xef, 0x05, 0xd4, 0xb9, 0x07, 0x61, 0x87, 0x79, 0x29, 0x6c, 0x4e, 0xa1,
	0x33, 0xf8, 0xe0, 0x7a, 0x8d, 0xf0, 0x36, 0xa2, 0xbc, 0x60, 0x99, 0x6b, 0x83, 0x11, 0x4e, 0xae,
	0x16, 0xf3, 0xcf, 0xa7, 0x0e, 0x99, 0xfc, 0xdd, 0x80, 0xd6, 0x6c, 0x9b, 0x87, 0xd0, 0xc9, 0xb9,
	0xe1, 0x05, 0xd3, 0x64, 0xae, 0x07, 0xb6, 0x83, 0xcd, 0x55, 0xf3, 0x3e, 0xb4, 0x73, 0xa6, 0x14,
	0x37, 0x1e, 0xe1, 0xbb, 0x21, 0x78, 0x93, 0x03, 0xc4, 0xb0, 0x1e, 0x5a, 0x5b, 0xe8, 0x05, 0xf5,
	0x27, 0xda, 0x86, 0xe6, 0x50, 0x51, 0x91, 0x8e, 0x43, 0xd1, 0x87, 0x2f, 0xb4, 0x0f, 0xad, 0x8c,
	0x1a, 0x4a, 0x4a, 0x6a, 0xc6, 0x2e, 0x90, 0xad, 0xde, 0x6a, 0xbc, 0x82, 0x23, 0x6b, 0x1c, 0x50,
	0x33, 0x46, 0x07, 0xd0, 0x66, 0xe2, 0x82, 0x2b, 0x29, 0x26, 0x2c, 0x4c, 0x89, 0x16, 0x9e, 0x37,
	0xa1, 0x4f, 0x21, 0x76, 0xe3, 0x5c, 0x1b, 0x55, 0xa5, 0xc6, 0x26, 0xda, 0x4c, 0x71, 0xdd, 0xc1,
	0xb7, 0x17, 0xfd, 0xc7, 0xb5, 0x76, 0x02, 0x1b, 0x05, 0x1d, 0xce, 0xc1, 0x23, 0xaf, 0x5e, 0xd0,
	0xe1, 0x14, 0x33, 0x80, 0x9d, 0xac, 0x32, 0x84, 0x8b, 0x91, 0x24, 0x29, 0x4d, 0xc7, 0x8c, 0x5c,
	0xd0, 0x82, 0x67, 0xb6, 0xf3, 0xb4, 0xde, 0x56, 0xbf, 0x5b, 0x59, 0x65, 0x4e, 0xc5, 0x48, 0xf6,
	0x2d, 0xef, 0x87, 0x40, 0x43, 0x1f, 0xc1, 0x4e, 0x55, 0x66, 0xd4, 0x30, 0x52, 0xf0, 0x09, 0x37,
	0xa4, 0x64, 0x8a, 0x4c, 0xb8, 0xa8, 0x0c, 0x8b, 0xc1, 0xf5, 0xa5, 0x2d, 0xef, 0x7e, 0x61, 0xbd,
	0x03, 0xa6, 0xbe, 0x75, 0x3e, 0xf4, 0x0e, 0x6c, 0xfc, 0x52, 0x31, 0x26, 0xea, 0x2e, 0x1a, 0xb7,
	0xdd, 0x66, 0x3b, 0xce, 0x18, 0x3a, 0x27, 0x7a, 0x0c, 0x77, 0x32, 0xe6, 0x7a, 0xac, 0x2f, 0xfb,
	0x7a, 0xd2, 0xc4, 0x1d, 0x07, 0xfe, 0xbf, 0x77, 0xfa, 0x3f, 0x91, 0x30, 0x6d, 0x5c, 0x35, 0x2c,
	0x70, 0xc2, 0x45, 0x6d, 0x38, 0x0a, 0x9a, 0xa7, 0xf4, 0xfc, 0xa5, 0x3d, 0x04, 0xb4, 0xb4, 0x8a,
	0x0d, 0xde, 0x6d, 0x87, 0xdf, 0x5c, 0x58, 0x82, 0x9a, 0x71, 0xf2, 0x1b, 0x74, 0xe6, 0x9b, 0x8f,
	0x4d, 0x34, 0x43, 0x55, 0xce, 0xfc, 0x4c, 0xb5, 0xcd, 0x64, 0xcd, 0x06, 0xdd, 0xdb, 0x5c, 0xc1,
	0xa3, 0x7b, 0x00, 0xba, 0xa4, 0x8a, 0xcd, 0x4f, 0xdd, 0x96, 0xb3, 0xb8, 0xb9, 0xfb, 0x10, 0xd0,
	0x84, 0xbe, 0x26, 0x95, 0x08, 0x2d, 0x84, 0x64, 0x95, 0xf1, 0x95, 0x7f, 0x0b, 0x6f, 0x4e, 0xe8,
	0xeb, 0xef, 0x6b, 0xc7, 0x71, 0x65, 0x74, 0xf2, 0x02, 0xa2, 0xfa, 0xe7, 0x01, 0x7d, 0x0d, 0x90,
	0xb1, 0xb2, 0x90, 0x8e, 0xe2, 0x52, 0xfc, 0x86, 0x5f, 0x96, 0x63, 0x87, 0x3c, 0xae, 0x0c, 0x6e,
	0x65, 0xf5, 0x6b, 0xf2, 0xd7, 0x0a, 0xb4, 0xa6, 0x0e, 0xf4, 0x1d, 0xec, 0x2c, 0x0f, 0xcf, 0x30,
	0x91, 0x83, 0xf8, 0x0d, 0xd9, 0x71, 0xc7, 0x2c, 0x0c, 0xd4, 0x30, 0x9e, 0xd1, 0x4b, 0xd8, 0xbe,
	0x7a, 0xc6, 0x87, 0x5f, 0xc1, 0x9b, 0xf2, 0xcd, 0x5c, 0x31, 0xef, 0x6d, 0xe2, 0x2c, 0x4e, 0xcc,
	0x35, 0x37, 0x31, 0x3b, 0x66, 0x6e, 0x4e, 0xf6, 0xa2, 0x1f, 0x9b, 0xfe, 0xd4, 0xc3, 0xa6, 0xd3,
	0xfd, 0xf0, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x02, 0xb4, 0x0c, 0xcc, 0x24, 0x0b, 0x00, 0x00,
}
