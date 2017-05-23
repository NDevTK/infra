// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/maruel/subcommands"
	"golang.org/x/net/context"

	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/common/errors"
	log "github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/system/environ"
	"github.com/luci/luci-go/common/system/exitcode"
	"github.com/luci/luci-go/common/system/filesystem"

	"infra/tools/kitchen/cookflags"
	"infra/tools/kitchen/migration"
	"infra/tools/kitchen/proto"
	"infra/tools/kitchen/third_party/recipe_engine"
)

// BootstrapStepName is the name of kitchen's step where it makes preparations
// for running a recipe, e.g. fetches a repository.
const BootstrapStepName = "recipe bootstrap"

// cmdCook checks out a repository at a revision and runs a recipe.
var cmdCook = &subcommands.Command{
	UsageLine: "cook -repository <repository URL> -recipe <recipe>",
	ShortDesc: "bootstraps a LUCI job.",
	LongDesc:  "Bootstraps a LUCI job.",
	CommandRun: func() subcommands.CommandRun {
		var c cookRun

		// Initialize our AnnotationFlags operational argument.
		c.rr.opArgs.AnnotationFlags = &recipe_engine.Arguments_AnnotationFlags{}
		c.rr.opArgs.EngineFlags = &recipe_engine.Arguments_EngineFlags{}

		c.CookFlags.Register(&c.Flags)

		return &c
	},
}

type cookRun struct {
	subcommands.CommandRunBase

	cookflags.CookFlags

	mode cookMode
	rr   recipeRun
}

// normalizeFlags validates and normalizes flags.
func (c *cookRun) normalizeFlags() error {
	if err := c.CookFlags.Normalize(); err != nil {
		return err
	}

	c.mode = cookModeSelector[c.Mode]
	c.rr.workDir = c.WorkDir
	c.rr.recipeName = c.RecipeName

	return nil
}

// ensureAndRunRecipe ensures that we have the recipe (according to -repository,
// -revision and -checkout-dir) and runs it.
func (c *cookRun) ensureAndRunRecipe(ctx context.Context, env environ.Env) *kitchen.CookResult {
	result := &kitchen.CookResult{}

	fail := func(err error) *kitchen.CookResult {
		if err == nil {
			panic("do not call fail with nil err")
		}
		if result.KitchenError != nil {
			panic("bug! forgot to return the result on previous error")
		}
		result.KitchenError = kitchenError(err)
		return result
	}

	if c.RepositoryURL == "" {
		// The ready-to-run recipe is already present on the file system.
		recipesPath, err := exec.LookPath(filepath.Join(c.CheckoutDir, "recipes"))
		if err != nil {
			return fail(errors.Annotate(err).Reason("could not find bundled recipes").Err())
		}
		c.rr.cmdPrefix = []string{recipesPath}
	} else {
		// Fetch the recipe.
		if err := checkoutRepository(ctx, c.CheckoutDir, c.RepositoryURL, c.Revision); err != nil {
			return fail(errors.Annotate(err).Reason("could not checkout %(repo)q at %(rev)q to %(path)q").
				D("repo", c.RepositoryURL).
				D("rev", c.Revision).
				D("path", c.CheckoutDir).
				Err())
		}
		// Read the path to the recipes.py within the fetched repo.
		recipesPath, err := getRecipesPath(c.CheckoutDir)
		if err != nil {
			return fail(errors.Annotate(err).Reason("could not read recipes.cfg").Err())
		}
		c.rr.cmdPrefix = []string{
			"python",
			filepath.Join(c.CheckoutDir, filepath.FromSlash(recipesPath), "recipes.py"),
		}
	}

	// Setup our working directory.
	var err error
	c.rr.workDir, err = prepareRecipeRunWorkDir(c.rr.workDir)
	if err != nil {
		return fail(errors.Annotate(err).Reason("failed to prepare workdir").Err())
	}

	// Tell the recipe to write the result protobuf message to a file and read
	// it below.
	c.rr.opArgs.EngineFlags.UseResultProto = true
	c.rr.outputResultJSONFile = filepath.Join(c.TempDir, "recipe-result.json")

	rv := 0
	if c.CookFlags.LogDogFlags.Active() {
		result.AnnotationUrl = c.CookFlags.LogDogFlags.AnnotationURL.String()
		rv, result.Annotations, err = c.runWithLogdogButler(ctx, &c.rr, env)
		if err != nil {
			return fail(errors.Annotate(err).Reason("failed to run recipe").Err())
		}
	} else {
		// This code is reachable only in buildbot mode.
		recipeCmd, err := c.rr.command(ctx, filepath.Join(c.TempDir, "rr"), env)
		if err != nil {
			return fail(errors.Annotate(err).Reason("failed to build recipe command").Err())
		}
		printCommand(ctx, recipeCmd)

		recipeCmd.Stdout = os.Stdout
		recipeCmd.Stderr = os.Stderr

		err = recipeCmd.Run()
		var hasRV bool
		rv, hasRV = exitcode.Get(err)
		if !hasRV {
			return fail(errors.Annotate(err).Reason("failed to run recipe").Err())
		}
	}
	result.RecipeExitCode = &kitchen.OptionalInt32{Value: int32(rv)}

	// Now read the recipe result file.
	recipeResultFile, err := os.Open(c.rr.outputResultJSONFile)
	if err != nil {
		// The recipe result file must exist and be readable.
		// If it is not, it is a fatal error.
		return fail(errors.Annotate(err).Reason("could not read recipe result file at %(path)q").
			D("path", c.rr.outputResultJSONFile).
			Err())
	}
	defer recipeResultFile.Close()

	if c.RecipeResultByteLimit > 0 {
		st, err := recipeResultFile.Stat()
		if err != nil {
			return fail(errors.Annotate(err).Reason("could not stat recipe result file at %(path)q").
				D("path", c.rr.outputResultJSONFile).
				Err())
		}

		if sz := st.Size(); sz > int64(c.RecipeResultByteLimit) {
			return fail(errors.Reason("recipe result file is %(size)d bytes which is more than %(limit)d").
				D("size", sz).
				D("limit", c.RecipeResultByteLimit).
				Err())
		}
	}

	result.RecipeResult = &recipe_engine.Result{}
	err = (&jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}).Unmarshal(recipeResultFile, result.RecipeResult)
	if err != nil {
		return fail(errors.Annotate(err).Reason("could not parse recipe result").Err())
	}

	// TODO(nodir): verify consistency between result.Build and result.RecipeResult.

	if result.RecipeResult.GetFailure() != nil && result.RecipeResult.GetFailure().GetFailure() == nil {
		// The recipe run has failed and the failure type is not step failure.
		result.KitchenError = &kitchen.KitchenError{
			Text: fmt.Sprintf("recipe infra failure: %s", result.RecipeResult.GetFailure().HumanReason),
			Type: kitchen.KitchenError_RECIPE_INFRA_FAILURE,
		}
		return result
	}

	return result
}

// stepModuleProperties are constructed properties for the "recipe_engine/step"
// recipe module.
type stepModuleProperties struct {
	PrefixPATH []string `json:"prefix_path,omitempty"`
}

// stepModuleProperties returns properties for the "recipe_engine/step" module.
func (c *cookRun) stepModuleProperties() *stepModuleProperties {
	if len(c.PrefixPathENV) == 0 {
		return nil
	}

	return &stepModuleProperties{
		PrefixPATH: []string(c.PrefixPathENV),
	}
}

// pathModuleProperties returns properties for the "recipe_engine/path" module.
func (c *cookRun) pathModuleProperties() (map[string]string, error) {
	recipeTempDir := filepath.Join(c.TempDir, "rt")
	if err := ensureDir(recipeTempDir); err != nil {
		return nil, err
	}
	paths := []struct{ name, path string }{
		{"cache_dir", c.CacheDir},
		{"temp_dir", recipeTempDir},
	}

	props := make(map[string]string, len(paths))
	for _, p := range paths {
		if p.path == "" {
			continue
		}
		native := filepath.FromSlash(p.path)
		if err := filesystem.AbsPath(&native); err != nil {
			return nil, err
		}
		props[p.name] = native
	}

	return props, nil
}

// prepareProperties parses the properties specified by flags,
// validates them and add some extra properties to describe current build
// environment.
// May mutate some properties.
func (c *cookRun) prepareProperties(env environ.Env) (map[string]interface{}, error) {
	props, err := parseProperties(c.Properties, c.PropertiesFile)
	if err != nil {
		return nil, errors.Annotate(err).Reason("could not parse properties").Err()
	}
	if props == nil {
		props = map[string]interface{}{}
	}

	// Reject reserved properties.
	rejectProperties := []string{
		"$recipe_engine/path",
		"$recipe_engine/step",
		"bot_id",
		// not specifying path_config means that all paths must be passed
		// explicitly. We do that below.
		"path_config",
	}
	for _, p := range rejectProperties {
		if _, ok := props[p]; ok {
			return nil, inputError("%s property must not be set", p)
		}
	}

	// Configure paths that the recipe will use.
	pathProps, err := c.pathModuleProperties()
	if err != nil {
		return nil, err
	}
	props["$recipe_engine/path"] = pathProps

	if p := c.stepModuleProperties(); p != nil {
		props["$recipe_engine/step"] = p
	}

	// Use "generic" infra path config. See
	// https://chromium.googlesource.com/chromium/tools/depot_tools/+/master/recipes/recipe_modules/infra_paths/
	props["path_config"] = "generic"

	if err := c.mode.addProperties(props, env); err != nil {
		return nil, errors.Annotate(err).Reason("chosen mode could not add properties").Err()
	}
	if _, ok := props[PropertyBotID]; !ok {
		return nil, errors.Reason("chosen mode didn't add %(p)s property").
			D("p", PropertyBotID).
			Err()
	}

	err = migration.TransformProperties(props)
	if err != nil {
		return nil, err
	}

	return props, nil
}

func (c *cookRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// The first thing we do is write a result file in case we crash or get killed.
	// Note that this code is not reachable if subcommands package could not
	// parse flags.
	result := &kitchen.CookResult{
		KitchenError: &kitchen.KitchenError{
			Type: kitchen.KitchenError_INTERNAL_ERROR,
			Text: "kitchen crashed or got killed",
		},
	}
	if err := c.flushResult(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	ctx := cli.GetContext(a, c, env)
	sysEnv := environ.System()

	result = c.run(ctx, args, sysEnv)
	fmt.Println(strings.Repeat("-", 35), "RESULTS", strings.Repeat("-", 36))
	proto.MarshalText(os.Stdout, result)
	fmt.Println(strings.Repeat("-", 80))

	if err := c.flushResult(result); err != nil {
		fmt.Fprintf(os.Stderr, "could not flush result to a file: %s\n", err)
		return 1
	}

	if result.KitchenError != nil {
		fmt.Fprintln(os.Stderr, "run failed because of the kitchen error")
		return 1
	}
	if result.RecipeExitCode == nil {
		panic("impossible: KitchenError is not nil, but there is no recipe exit code")
	}
	return int(result.RecipeExitCode.Value)
}

// run runs the cook subcommmand and returns cook result.
func (c *cookRun) run(ctx context.Context, args []string, env environ.Env) *kitchen.CookResult {
	fail := func(err error) *kitchen.CookResult {
		return &kitchen.CookResult{KitchenError: kitchenError(err)}
	}
	// Process input.
	if len(args) != 0 {
		return fail(inputError("unexpected arguments: %v", args))
	}
	if _, err := os.Getwd(); err != nil {
		return fail(inputError("failed to resolve CWD: %s", err))
	}
	if err := c.normalizeFlags(); err != nil {
		return fail(err)
	}

	// initialize temp dir.
	if c.TempDir == "" {
		tdir, err := ioutil.TempDir("", "kitchen")
		if err != nil {
			return fail(errors.Annotate(err).Reason("failed to create temporary directory").Err())
		}
		c.TempDir = tdir
		defer func() {
			if rmErr := os.RemoveAll(tdir); rmErr != nil {
				log.Warningf(ctx, "Failed to clean up temporary directory at [%s]: %s", tdir, rmErr)
			}
		}()
	}

	// Prepare recipe properties. Print them too.
	var err error
	if c.rr.properties, err = c.prepareProperties(env); err != nil {
		return fail(err)
	}
	propsJSON, err := json.MarshalIndent(c.rr.properties, "", "  ")
	if err != nil {
		return fail(errors.Annotate(err).Reason("could not marshal properties to JSON").Err())
	}
	log.Infof(ctx, "using properties:\n%s", propsJSON)

	// If we're not using LogDog, send out annotations.
	bootstrapSuccess := false
	if c.CookFlags.LogDogFlags.ShouldEmitAnnotations() {
		// This code is reachable only in buildbot mode.

		annotate := func(args ...string) {
			fmt.Printf("@@@%s@@@\n", strings.Join(args, "@"))
		}
		annotate("SEED_STEP", BootstrapStepName)
		annotate("STEP_CURSOR", BootstrapStepName)
		annotate("STEP_STARTED")
		defer func() {
			annotate("STEP_CURSOR", BootstrapStepName)
			if bootstrapSuccess {
				annotate("STEP_CLOSED")
			} else {
				annotate("STEP_EXCEPTION")
			}
		}()

		for k, v := range c.rr.properties {
			// Order is not stable, but that is okay.
			jv, err := json.Marshal(v)
			if err != nil {
				return fail(errors.Annotate(err).Err())
			}
			annotate("SET_BUILD_PROPERTY", k, string(jv))
		}
	}

	c.updateEnv(env)
	// Make kitchen use the new $PATH too.
	// In practice, we do it so that kitchen uses the installed git wrapper.
	path, _ := env.Get("PATH")
	if err := os.Setenv("PATH", path); err != nil {
		return fail(errors.Annotate(err).Reason("failed to update process PATH").Err())
	}

	// Run the recipe.
	result := c.ensureAndRunRecipe(ctx, env)
	bootstrapSuccess = result.KitchenError == nil
	return result
}

// flushResult writes the result to c.OutputResultJSOPath file
// if the path is specified.
func (c *cookRun) flushResult(result *kitchen.CookResult) (err error) {
	if c.OutputResultJSONPath == "" {
		return nil
	}

	defer func() {
		if err != nil {
			err = errors.Annotate(err).Reason("could not write result file at %(path)q").
				D("path", c.OutputResultJSONPath).
				Err()
		}
	}()

	f, err := os.Create(c.OutputResultJSONPath)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			err = errors.Annotate(err).Reason("could not close file").Err()
		}
	}()
	m := jsonpb.Marshaler{EmitDefaults: true}
	return m.Marshal(f, result)
}

// updateEnv updates $PATH, $PYTHONPATH and temp path env variables in env.
func (c *cookRun) updateEnv(env environ.Env) {
	addPaths := func(key string, paths []string) {
		if len(paths) == 0 {
			return
		}
		cur, _ := env.Get(key)
		all := make([]string, 0, len(paths)+1)
		all = append(all, paths...)
		if len(cur) > 0 {
			all = append(all, cur)
		}
		env.Set(key, strings.Join(all, string(os.PathListSeparator)))
	}

	addPaths("PATH", c.PrefixPathENV)
	addPaths("PYTHONPATH", c.PythonPaths)

	for _, entry := range c.SetEnvAbspath {
		env.SetEntry(entry)
	}

	// Tell subprocesses to use Kitchen's temp dir.
	if c.TempDir == "" {
		// It should have been initialized in c.run.
		panic("TempDir was not initialzied earlier")
	}
	for _, v := range []string{"TEMPDIR", "TMPDIR", "TEMP", "TMP", "MAC_CHROMIUM_TMPDIR"} {
		env.Set(v, c.TempDir)
	}
}

func parseProperties(properties, propertiesFile string) (result map[string]interface{}, err error) {
	if properties != "" {
		err = unmarshalJSONWithNumber([]byte(properties), &result)
		if err != nil {
			err = inputError("could not parse properties %s\n%s", properties, err)
		}
		return
	}
	if propertiesFile != "" {
		b, err := ioutil.ReadFile(propertiesFile)
		if err != nil {
			err = inputError("could not read properties file %s\n%s", propertiesFile, err)
			return nil, err
		}
		err = unmarshalJSONWithNumber(b, &result)
		if err != nil {
			err = inputError("could not parse JSON from file %s\n%s\n%s",
				propertiesFile, b, err)
		}
	}
	return
}
