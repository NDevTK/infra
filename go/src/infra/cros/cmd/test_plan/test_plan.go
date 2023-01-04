package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/maruel/subcommands"
	"google.golang.org/protobuf/encoding/prototext"
	protov2 "google.golang.org/protobuf/proto"

	igerrit "infra/cros/internal/gerrit"
	"infra/cros/internal/manifestutil"
	"infra/cros/internal/shared"
	"infra/cros/internal/testplan"
	"infra/cros/internal/testplan/migrationstatus"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"

	"go.chromium.org/chromiumos/config/go/test/plan"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	luciflag "go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	cvpb "go.chromium.org/luci/cv/api/config/v2"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

var logCfg = gologger.LoggerConfig{
	Out: os.Stderr,
}

func errToCode(a subcommands.Application, err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", a.GetName(), err)
		return 1
	}

	return 0
}

func unmarshalTextproto(path string, m protov2.Message) error {
	protoBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return prototext.Unmarshal(protoBytes, m)
}

// baseTestPlanRun embeds subcommands.CommandRunBase and implements flags shared
// across commands, such as logging and auth flags. It should be embedded in
// another struct that implements Run() for a specific command. baseTestPlanRun
// implements cli.ContextModificator, to set the log level based on flags.
type baseTestPlanRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	logLevel  logging.Level
}

// addSharedFlags adds shared auth and logging flags.
func (r *baseTestPlanRun) addSharedFlags(authOpts auth.Options) {
	r.authFlags = authcli.Flags{}
	r.authFlags.Register(r.GetFlags(), authOpts)

	r.logLevel = logging.Info
	r.Flags.Var(&r.logLevel, "loglevel", text.Doc(`
	Log level, valid options are "debug", "info", "warning", "error". Default is "info".
	`))
}

// ModifyContext returns a new Context with the log level set in the flags.
func (r *baseTestPlanRun) ModifyContext(ctx context.Context) context.Context {
	return logging.SetLevel(ctx, r.logLevel)
}

func app(authOpts auth.Options) *cli.Application {
	return &cli.Application{
		Name:    "test_plan",
		Title:   "A tool to work with SourceTestPlan protos in DIR_METADATA files.",
		Context: logCfg.Use,
		Commands: []*subcommands.Command{
			cmdRelevantPlans(authOpts),
			cmdValidate(authOpts),
			cmdMigrationStatus(authOpts),

			authcli.SubcommandInfo(authOpts, "auth-info", false),
			authcli.SubcommandLogin(authOpts, "auth-login", false),
			authcli.SubcommandLogout(authOpts, "auth-logout", false),

			subcommands.CmdHelp,
		},
	}
}

func cmdRelevantPlans(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "relevant-plans -cl CL1 [-cl CL2] -out OUTPUT",
		ShortDesc: "Find SourceTestPlans relevant to a set of CLs",
		LongDesc: text.Doc(`
		Find SourceTestPlans relevant to a set of CLs.

		Computes SourceTestPlans from "DIR_METADATA" files and returns plans
		relevant to the files changed by a CL.
	`),
		CommandRun: func() subcommands.CommandRun {
			r := &relevantPlansRun{}
			r.addSharedFlags(authOpts)

			r.Flags.Var(luciflag.StringSlice(&r.cls), "cl", text.Doc(`
			CL URL for the patchsets being tested. Must be specified at least once.
			Changes will be merged in the order they are passed on the command line.

			Example: https://chromium-review.googlesource.com/c/chromiumos/platform2/+/123456
		`))
			r.Flags.StringVar(&r.out, "out", "", "Path to the output test plan")

			return r
		},
	}
}

type relevantPlansRun struct {
	baseTestPlanRun
	cls []string
	out string
}

func (r *relevantPlansRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)
	return errToCode(a, r.run(ctx))
}

// getChangeRevs parses each of rawCLURLs and returns a ChangeRev.
func getChangeRevs(ctx context.Context, authedClient *http.Client, rawCLURLs []string) ([]*igerrit.ChangeRev, error) {
	changeRevs := make([]*igerrit.ChangeRev, len(rawCLURLs))

	for i, cl := range rawCLURLs {
		changeRevKey, err := igerrit.ParseCLURL(cl)
		if err != nil {
			return nil, err
		}

		changeRev, err := igerrit.GetChangeRev(
			ctx, authedClient, changeRevKey.ChangeNum, changeRevKey.Revision, changeRevKey.Host, shared.DefaultOpts,
		)
		if err != nil {
			return nil, err
		}

		changeRevs[i] = changeRev
	}

	return changeRevs, nil
}

// writePlans writes each of plans to a textproto file. The first plan is in a
// file named "relevant_plan_1.textpb", the second is in
// "relevant_plan_2.textpb", etc.
//
// TODO(b/182898188): Consider making a message to hold multiple SourceTestPlans
// instead of writing multiple files.
func writePlans(ctx context.Context, plans []*plan.SourceTestPlan, outPath string) error {
	logging.Infof(ctx, "writing output to %s", outPath)

	err := os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		return err
	}

	for i, plan := range plans {
		outFile, err := os.Create(path.Join(outPath, fmt.Sprintf("relevant_plan_%d.textpb", i)))
		if err != nil {
			return err
		}
		defer outFile.Close()

		err = proto.MarshalText(outFile, plan)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *relevantPlansRun) validateFlags() error {
	if len(r.cls) == 0 {
		return errors.New("-cl must be specified at least once")
	}

	if r.out == "" {
		return errors.New("-out is required")
	}

	return nil
}

func (r *relevantPlansRun) run(ctx context.Context) error {
	if err := r.validateFlags(); err != nil {
		return err
	}

	authOpts, err := r.authFlags.Options()
	if err != nil {
		return err
	}

	authedClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).Client()
	if err != nil {
		return err
	}

	var changeRevs []*igerrit.ChangeRev

	logging.Infof(ctx, "fetching metadata for CLs")

	changeRevs, err = getChangeRevs(ctx, authedClient, r.cls)
	if err != nil {
		return err
	}

	for i, changeRev := range changeRevs {
		logging.Debugf(ctx, "change rev %d: %q", i, changeRev)
	}

	// Use a workdir creation function that returns a tempdir, and removes the
	// entire tempdir on cleanup.
	workdirFn := func() (string, func() error, error) {
		workdir, err := ioutil.TempDir("", "")
		if err != nil {
			return "", nil, err
		}

		return workdir, func() error { return os.RemoveAll((workdir)) }, nil
	}

	plans, err := testplan.FindRelevantPlans(ctx, changeRevs, workdirFn)
	if err != nil {
		return err
	}

	return writePlans(ctx, plans, r.out)
}

func cmdValidate(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "validate DIR1 [DIR2...]",
		ShortDesc: "validate metadata files",
		LongDesc: text.Doc(`
		Validate metadata files.

		Validation logic on "DIR_METADATA" files specific to ChromeOS test planning.

		Each positional argument should be a path to a directory to compute and validate
		metadata for.

		The subcommand returns a non-zero exit code if any of the files is invalid.
	`),
		CommandRun: func() subcommands.CommandRun {
			r := &validateRun{}
			r.addSharedFlags(authOpts)
			return r
		},
	}
}

type validateRun struct {
	baseTestPlanRun
}

func (r *validateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	return errToCode(a, r.run(a, args, env))
}

func (r *validateRun) validateFlags(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one directory must be specified as a positional argument")
	}

	return nil
}

func (r *validateRun) run(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, r, env)

	if err := r.validateFlags(args); err != nil {
		return err
	}

	authOpts, err := r.authFlags.Options()
	if err != nil {
		return err
	}

	authedClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).Client()
	if err != nil {
		return err
	}

	gerritClient, err := igerrit.NewClient(authedClient)
	if err != nil {
		return err
	}

	mapping, err := dirmd.ReadMapping(ctx, dirmdpb.MappingForm_ORIGINAL, true, args...)
	if err != nil {
		return err
	}

	return testplan.ValidateMapping(ctx, gerritClient, mapping)
}

func cmdMigrationStatus(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "migration-status -crossrcroot ~/chromiumos [-project PROJECT1 -project PROJECT2...]",
		ShortDesc: "summarize the migration status of projects",
		LongDesc: text.Doc(`
		Summarize the migration status of projects in the manifest.

		Reads the default manifest, Buildbucket config, and CV config from
		-crossrcroot, and for each project in the manifest checks if it has a
		matching CrosTestPlanV2Properties.ProjectMigrationConfig in the input
		properties of the CQ orchestrators. Prints a summary of the number of
		projects migrated.

		Projects that are not in the "ToT" ConfigGroup of cvConfig or are
		excluded from the CQ orchestrator by a LocationFilter are skipped.

		Optionally takes multiple -project arguments, and prints whether those
		specific projects are migrated. If one of these projects does not exist
		in the manifest, an error is returned.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &migrationStatusRun{}
			r.addSharedFlags(authOpts)

			r.Flags.StringVar(&r.crosSrcRoot, "crossrcroot", "", text.Doc(`
			Required, path to the root of a ChromeOS checkout. The manifest and
			generated Buildbucket config found in this checkout will be used.
			`))
			r.Flags.Var(luciflag.StringSlice(&r.projects), "project", text.Doc(`
			Projects to check the specific migration status of.
			`))
			return r
		},
	}
}

type migrationStatusRun struct {
	baseTestPlanRun
	crosSrcRoot string
	projects    []string
}

func (r *migrationStatusRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)
	return errToCode(a, r.run(ctx))
}

func (r *migrationStatusRun) validateFlagsAndSetDefaults() error {
	if r.crosSrcRoot == "" {
		return fmt.Errorf("-crossrcroot must be set")
	}

	return nil
}

func (r *migrationStatusRun) run(ctx context.Context) error {
	if err := r.validateFlagsAndSetDefaults(); err != nil {
		return err
	}

	manifestPath := filepath.Join(r.crosSrcRoot, "manifest-internal", "default.xml")
	logging.Debugf(ctx, "reading manifest from %q", manifestPath)
	manifest, err := manifestutil.LoadManifestFromFileWithIncludes(manifestPath)

	infraCfgPath := filepath.Join(r.crosSrcRoot, "infra", "config", "generated")

	cvCfgPath := filepath.Join(infraCfgPath, "commit-queue.cfg")
	bbCfgPath := filepath.Join(infraCfgPath, "cr-buildbucket.cfg")

	logging.Debugf(ctx, "reading CV config from %q", cvCfgPath)
	cvConfig := &cvpb.Config{}
	if err := unmarshalTextproto(cvCfgPath, cvConfig); err != nil {
		return err
	}

	logging.Debugf(ctx, "reading Buildbucket config from %q", bbCfgPath)
	bbCfg := &bbpb.BuildbucketCfg{}
	if err := unmarshalTextproto(bbCfgPath, bbCfg); err != nil {
		return err
	}

	textSummary, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, r.projects)
	if err != nil {
		return err
	}

	fmt.Print(textSummary)

	return nil
}

func main() {
	opts := chromeinfra.DefaultAuthOptions()
	opts.Scopes = append(opts.Scopes, gerrit.OAuthScope)
	os.Exit(subcommands.Run(app(opts), nil))
}
