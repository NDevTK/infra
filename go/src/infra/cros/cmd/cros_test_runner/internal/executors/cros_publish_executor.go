package executors

import (
	"context"
	"fmt"
	"strings"

	_go "go.chromium.org/chromiumos/config/go"
	"go.chromium.org/chromiumos/config/go/test/api"
	test_api "go.chromium.org/chromiumos/config/go/test/api"
	test_api_metadata "go.chromium.org/chromiumos/config/go/test/api/metadata"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// CrosPublishExecutor represents executor for all cros-publish related commands.
type CrosPublishExecutor struct {
	*interfaces.AbstractExecutor

	Container               interfaces.ContainerInterface
	GcsPublishServiceClient test_api.GenericPublishServiceClient
	TkoPublishServiceClient test_api.GenericPublishServiceClient
	RdbPublishServiceClient test_api.GenericPublishServiceClient
	ServerAddress           string
}

func NewCrosPublishExecutor(container interfaces.ContainerInterface, execType interfaces.ExecutorType) *CrosPublishExecutor {
	if execType != CrosGcsPublishExecutorType && execType != CrosTkoPublishExecutorType && execType != CrosRdbPublishExecutorType {
		return nil
	}
	return &CrosPublishExecutor{AbstractExecutor: interfaces.NewAbstractExecutor(execType), Container: container}
}

func (ex *CrosPublishExecutor) ExecuteCommand(ctx context.Context, cmdInterface interfaces.CommandInterface) error {
	switch cmd := cmdInterface.(type) {
	case *commands.GcsPublishServiceStartCmd:
		return ex.gcsPublishStartCommandExecution(ctx, cmd)
	case *commands.GcsPublishUploadCmd:
		return ex.gcsPublishUploadCommandExecution(ctx, cmd)
	case *commands.RdbPublishServiceStartCmd:
		return ex.rdbPublishStartCommandExecution(ctx, cmd)
	case *commands.RdbPublishUploadCmd:
		return ex.rdbPublishUploadCommandExecution(ctx, cmd)
	case *commands.TkoPublishServiceStartCmd:
		return ex.tkoPublishStartCommandExecution(ctx, cmd)
	case *commands.TkoPublishUploadCmd:
		return ex.tkoPublishUploadCommandExecution(ctx, cmd)
	default:
		return fmt.Errorf("Command type %s is not supported by %s executor type!", cmd.GetCommandType(), ex.GetExecutorType())
	}
}

// -- GCS Commands --

// gcsPublishStartCommandExecution executes the gcs-publish start command.
func (ex *CrosPublishExecutor) gcsPublishStartCommandExecution(ctx context.Context, cmd *commands.GcsPublishServiceStartCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "gcs-publish service start")
	defer func() { step.End(err) }()

	gcsPublishTemplate := &test_api.CrosPublishTemplate{PublishType: test_api.CrosPublishTemplate_PUBLISH_GCS, PublishSrcDir: cmd.GcsPublishSrcDir}
	publishClient, err := ex.Start(ctx, &api.Template{Container: &api.Template_CrosPublish{CrosPublish: gcsPublishTemplate}})
	if err != nil {
		return errors.Annotate(err, "Start gcs-publish cmd err: ").Err()
	}

	ex.GcsPublishServiceClient = publishClient

	return err
}

// gcsPublishUploadCommandExecution executes the gcs-publish upload command.
func (ex *CrosPublishExecutor) gcsPublishUploadCommandExecution(ctx context.Context, cmd *commands.GcsPublishUploadCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "gcs-publish upload")
	defer func() { step.End(err) }()
	step.SetSummaryMarkdown(fmt.Sprintf("* [GCS Link](%s)", getGcsClickableLink(cmd.GcsUrl)))

	// Create request.
	artifactDirPath := &_go.StoragePath{HostType: _go.StoragePath_LOCAL, Path: common.GcsPublishTestArtifactsDir}
	gcsPath := &_go.StoragePath{HostType: _go.StoragePath_GS, Path: cmd.GcsUrl}
	gcsMetadata, err := anypb.New(&test_api.PublishGcsMetadata{GcsPath: gcsPath})
	if err != nil {
		return errors.Annotate(err, "Creating publish gcs metadata err: ").Err()
	}

	gcsPublishReq := &test_api.PublishRequest{ArtifactDirPath: artifactDirPath, TestResponse: nil, Metadata: gcsMetadata}
	return ex.InvokePublishWithAsyncLogging(ctx, "gcs-publish", gcsPublishReq, ex.GcsPublishServiceClient, step)
}

// -- RDB Commands --

// rdbPublishStartCommandExecution executes the rdb-publish start command.
func (ex *CrosPublishExecutor) rdbPublishStartCommandExecution(ctx context.Context, cmd *commands.RdbPublishServiceStartCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "rdb-publish service start")
	defer func() { step.End(err) }()

	rdbPublishTemplate := &test_api.CrosPublishTemplate{PublishType: test_api.CrosPublishTemplate_PUBLISH_RDB}
	publishClient, err := ex.Start(ctx, &api.Template{Container: &api.Template_CrosPublish{CrosPublish: rdbPublishTemplate}})
	if err != nil {
		return errors.Annotate(err, "Start rdb-publish cmd err: ").Err()
	}

	ex.RdbPublishServiceClient = publishClient

	return err
}

// rdbPublishUploadCommandExecution executes the rdb-publish upload command.
func (ex *CrosPublishExecutor) rdbPublishUploadCommandExecution(ctx context.Context, cmd *commands.RdbPublishUploadCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "rdb-publish upload")
	defer func() { step.End(err) }()
	step.SetSummaryMarkdown(fmt.Sprintf("* [Stainless Link](%s)", cmd.StainlessUrl))

	// Create request.
	rdbMetadata, err := anypb.New(&test_api_metadata.PublishRdbMetadata{CurrentInvocationId: cmd.CurrentInvocationId, TestResult: cmd.TestResultForRdb, StainlessUrl: cmd.StainlessUrl})
	if err != nil {
		return errors.Annotate(err, "Creating publish rdb metadata err: ").Err()
	}

	// TODO (azrhaman): remove artifactDirPath after unnecessary rdb validation is removed.
	artifactDirPath := &_go.StoragePath{HostType: _go.StoragePath_LOCAL, Path: "/tmp/rdb-publish-test-artifacts/"}
	rdbPublishReq := &test_api.PublishRequest{ArtifactDirPath: artifactDirPath, TestResponse: nil, Metadata: rdbMetadata}
	return ex.InvokePublishWithAsyncLogging(ctx, "rdb-publish", rdbPublishReq, ex.RdbPublishServiceClient, step)
}

// -- TKO Commands --

// tkoPublishStartCommandExecution executes the tko-publish start command.
func (ex *CrosPublishExecutor) tkoPublishStartCommandExecution(ctx context.Context, cmd *commands.TkoPublishServiceStartCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "tko-publish service start")
	defer func() { step.End(err) }()

	tkoPublishTemplate := &test_api.CrosPublishTemplate{PublishType: test_api.CrosPublishTemplate_PUBLISH_TKO, PublishSrcDir: cmd.TkoPublishSrcDir}
	publishClient, err := ex.Start(ctx, &api.Template{Container: &api.Template_CrosPublish{CrosPublish: tkoPublishTemplate}})
	if err != nil {
		return errors.Annotate(err, "Start tko-publish cmd err: ").Err()
	}

	ex.TkoPublishServiceClient = publishClient

	return err
}

// tkoPublishUploadCommandExecution executes the tko-publish upload command.
func (ex *CrosPublishExecutor) tkoPublishUploadCommandExecution(ctx context.Context, cmd *commands.TkoPublishUploadCmd) error {
	var err error
	step, ctx := build.StartStep(ctx, "tko-publish upload")
	defer func() { step.End(err) }()

	// Create request.
	artifactDirPath := &_go.StoragePath{HostType: _go.StoragePath_LOCAL, Path: common.TKOPublishTestArtifactsDir}
	tkoMetadata, err := anypb.New(&test_api.PublishTkoMetadata{JobName: cmd.TkoJobName})
	if err != nil {
		return errors.Annotate(err, "Creating publish tko metadata err: ").Err()
	}

	tkoPublishReq := &test_api.PublishRequest{ArtifactDirPath: artifactDirPath, TestResponse: nil, Metadata: tkoMetadata}
	return ex.InvokePublishWithAsyncLogging(ctx, "tko-publish", tkoPublishReq, ex.TkoPublishServiceClient, step)
}

// Start starts the cros-publish server.
func (ex *CrosPublishExecutor) Start(ctx context.Context, template *api.Template) (test_api.GenericPublishServiceClient, error) {
	if template == nil {
		return nil, fmt.Errorf("Cannot start publish service with empty template.")
	}

	// Process container.
	serverAddress, err := ex.Container.ProcessContainer(ctx, template)
	if err != nil {
		return nil, errors.Annotate(err, "error processing container: ").Err()
	}

	ex.ServerAddress = serverAddress

	// Connect with the service.
	conn, err := common.ConnectWithService(ctx, serverAddress)
	if err != nil {
		logging.Infof(ctx, "error during connecting server at %s: %s", serverAddress, err.Error())
		return nil, err
	}
	logging.Infof(ctx, "Connected with service.")

	// Create new client.
	publishClient := api.NewGenericPublishServiceClient(conn)
	if publishClient == nil {
		return nil, fmt.Errorf("GenericPublishServiceClient is nil")
	}

	return publishClient, nil
}

// Publish invokes the publish endpoint of cros-publish.
func (ex *CrosPublishExecutor) Publish(ctx context.Context, publishReq *test_api.PublishRequest, publishClient test_api.GenericPublishServiceClient) (*test_api.PublishResponse, error) {
	if publishClient == nil {
		return nil, fmt.Errorf("CrosPublishServiceClient is nil in CrosPublishExecutor")
	}
	if publishReq == nil {
		return nil, fmt.Errorf("Cannot publish results with empty publish request.")
	}

	publishOp, err := publishClient.Publish(ctx, publishReq, grpc.EmptyCallOption{})
	if err != nil {
		return nil, errors.Annotate(err, "publish failure: ").Err()
	}

	opResp, err := common.ProcessLro(ctx, publishOp)
	if err != nil {
		return nil, errors.Annotate(err, "publish lro failure: ").Err()
	}

	publishResp := &test_api.PublishResponse{}
	if err := opResp.UnmarshalTo(publishResp); err != nil {
		logging.Infof(ctx, "publish lro response unmarshalling failed: %s", err.Error())
		return nil, errors.Annotate(err, "publish lro response unmarshalling failed: ").Err()
	}

	return publishResp, nil
}

// InvokePublishWithAsyncLogging invokes publish endpoint of the service with async logging.
func (ex *CrosPublishExecutor) InvokePublishWithAsyncLogging(ctx context.Context, publishType string, request *api.PublishRequest, client api.GenericPublishServiceClient, step *build.Step) error {
	if request == nil {
		return fmt.Errorf("Cannot publish result for %s with empty publish request.", publishType)
	}
	if client == nil {
		return fmt.Errorf("Cannot publish result for %s with no established publish client.", publishType)
	}
	if ex.Container == nil {
		return fmt.Errorf("Cannot publish result for %s with empty publish container.", publishType)
	}

	// Write request.
	common.WriteProtoToStepLog(ctx, step, request, fmt.Sprintf("%s request", publishType))

	// Get container logs location.
	logsLoc, err := ex.Container.GetLogsLocation()
	if err != nil {
		logging.Infof(ctx, "error during getting %s container log location: %s", publishType, err)
		return err
	}

	// Stream logs async to build step.
	containerLog := step.Log(fmt.Sprintf("%s Log", publishType))
	taskDone, wg, err := common.StreamLogAsync(ctx, logsLoc, containerLog)
	if err != nil {
		logging.Infof(ctx, "Warning: error during reading %s container log: %s", publishType, err)
	}

	// Publish.
	resp, err := ex.Publish(ctx, request, client)
	taskDone <- true // Notify logging process that main task is done
	wg.Wait()        // wait for the logging to complete

	if err != nil {
		return errors.Annotate(err, fmt.Sprintf("%s publish cmd err: ", publishType)).Err()
	}

	common.WriteProtoToStepLog(ctx, step, resp, fmt.Sprintf("%s response", publishType))

	return nil
}

// getGcsClickableLink constructs the gcs cliclable link from provided gs url.
func getGcsClickableLink(gsUrl string) string {
	gsPrefix := "gs://"
	urlSuffix := gsUrl
	if strings.HasPrefix(gsUrl, gsPrefix) {
		urlSuffix = gsUrl[len(gsPrefix):]
	}
	return fmt.Sprintf("%s%s", common.GcsUrlPrefix, urlSuffix)
}
