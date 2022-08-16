// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	cipd "go.chromium.org/luci/cipd/client/cipd/builder"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// defaultExcludePrefixes excludes everything not packaged in "ios" or "mac"
// packages. In the past this has excluded parts of Xcode.app that
// are not necessary when uploading Xcode contents, but as each release of Xcode
// may change what is required, proceed with caution before adding new folders
// to exclude.
var defaultExcludePrefixes = []string{}

// iosPrefixes excludes parts of Xcode.app not required for building
// Chrome on Mac OS, but is useful for iOS.
var iosPrefixes = []string{
	"Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator",
	"Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/SDKs",
}

// Packages is the set of CIPD package definitions. The key is a convenience
// package name for direct reference.
type Packages map[string]cipd.PackageDef

// PackageSpec bundles the package name with a path to its YAML definition file.
type PackageSpec struct {
	Name     string
	YamlPath string
}

func isUnderPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		p := filepath.Join(strings.Split(prefix, "/")...)
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// MakePackageArgs are the parameters for makePackage() to keep them manageable.
type MakePackageArgs struct {
	cipdPackageName   string
	cipdPackagePrefix string
	rootPath          string
	includePrefixes   []string
	excludePrefixes   []string
}

// Makes a CIPD PackageDef using |MakePackageArgs|. Only files in |rootPath|,
// and meanwhile under any of |includePrefixes| relative path prefixes (if
// provided), and not under any of |excludePrefixes| will be included. All paths
// in |rootPath| are first filtered by |includePrefixes| (if provided), then
// tested to ensure it's not in |excludePrefixes|, to be included in the
// package.
func makePackage(args MakePackageArgs) (packageDef cipd.PackageDef, err error) {
	absRootPath, err := filepath.Abs(args.rootPath)
	if err != nil {
		err = errors.Annotate(err, "failed to create an absolute root path from %s", args.rootPath).Err()
		return
	}
	packageDef = cipd.PackageDef{
		Root:             absRootPath,
		InstallMode:      "copy",
		PreserveModTime:  true,
		PreserveWritable: true,
		Package:          args.cipdPackagePrefix + "/" + args.cipdPackageName,
		Data: []cipd.PackageChunkDef{
			{VersionFile: ".xcode_versions/" + args.cipdPackageName + ".cipd_version"},
		},
	}

	err = filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsDir() {
			if !strings.HasPrefix(path, absRootPath+string(os.PathSeparator)) {
				return errors.Reason("file is not in the source folder: %s", path).Err()
			}
			relPath := path[len(absRootPath)+1:]

			if len(args.includePrefixes) > 0 && !isUnderPrefix(relPath, args.includePrefixes) {
				return nil
			}
			if len(args.excludePrefixes) > 0 && isUnderPrefix(relPath, args.excludePrefixes) {
				return nil
			}

			packageDef.Data = append(packageDef.Data, cipd.PackageChunkDef{File: relPath})
		}
		return nil
	})
	return packageDef, err
}

// Makes Xcode's CIPD package definitions, including "mac" and "ios" package
// types.
// Legacy iOS package contains all runtimes in Xcode (default before spring
// 2021), while new iOS package contains no runtimes.
func makeXcodePackages(xcodeAppPath string, cipdPackagePrefix string, legacyIOSPackage bool) (p Packages, err error) {
	absXcodeAppPath, err := filepath.Abs(xcodeAppPath)
	if err != nil {
		err = errors.Annotate(err, "failed to create an absolute path from %s", xcodeAppPath).Err()
		return
	}

	// Mac package exclude prefixes include prefixes in |defaultExcludePrefixes|
	// and |iosPrefixes|. Use |make|, |copy| and |append| functions to ensure
	// slices won't be accidentally changed.
	excludePrefixesForMacPackage := make([]string, len(defaultExcludePrefixes))
	copy(excludePrefixesForMacPackage, defaultExcludePrefixes)
	excludePrefixesForMacPackage = append(excludePrefixesForMacPackage, iosPrefixes...)

	macMakePackageArgs := MakePackageArgs{
		cipdPackageName:   MacPackageName,
		cipdPackagePrefix: cipdPackagePrefix,
		rootPath:          absXcodeAppPath,
		includePrefixes:   []string{},
		excludePrefixes:   excludePrefixesForMacPackage,
	}
	mac, err := makePackage(macMakePackageArgs)
	if err != nil {
		err = errors.Annotate(err, "failed to create mac cipd pakcage").Err()
	}

	excludePrefixesForiOSPackage := make([]string, len(defaultExcludePrefixes))
	copy(excludePrefixesForiOSPackage, defaultExcludePrefixes)
	if !legacyIOSPackage {
		excludePrefixesForiOSPackage = append(excludePrefixesForMacPackage, XcodeIOSSimulatorRuntimeRelPath)
	}

	iosMakePackageArgs := MakePackageArgs{
		cipdPackageName:   IosPackageName,
		cipdPackagePrefix: cipdPackagePrefix,
		rootPath:          absXcodeAppPath,
		includePrefixes:   iosPrefixes,
		excludePrefixes:   excludePrefixesForiOSPackage,
	}
	ios, err := makePackage(iosMakePackageArgs)
	if err != nil {
		err = errors.Annotate(err, "failed to create ios cipd pakcage").Err()
	}

	p = Packages{"mac": mac, "ios": ios}
	return
}

// buildCipdPackages builds and optionally uploads CIPD packages to the
// server. `buildFn` callback takes a PackageSpec for each package in `packages`
// and is expected to call `cipd pkg-build` or `cipd create` on it.
func buildCipdPackages(packages Packages, buildFn func(PackageSpec) error) error {
	tmpDir, err := ioutil.TempDir("", "mac_toolchain_")
	if err != nil {
		return errors.Annotate(err, "cannot create a temporary folder for CIPD package configuration files in %s", os.TempDir()).Err()
	}
	defer os.RemoveAll(tmpDir)

	// Iterate deterministically (for testability).
	names := make([]string, 0, len(packages))
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		p := packages[name]
		yamlBytes, err := yaml.Marshal(p)
		if err != nil {
			return errors.Annotate(err, "failed to serialize %s.yaml", name).Err()
		}
		yamlPath := filepath.Join(tmpDir, name+".yaml")
		if err = ioutil.WriteFile(yamlPath, yamlBytes, 0600); err != nil {
			return errors.Annotate(err, "failed to write package definition file %s", yamlPath).Err()
		}
		if err = buildFn(PackageSpec{Name: p.Package, YamlPath: yamlPath}); err != nil {
			return err
		}
	}
	return nil
}

func createBuilder(ctx context.Context, tags []string, refs []string, serviceAccountJSON, outputDir string) func(PackageSpec) error {
	builder := func(p PackageSpec) error {
		args := []string{}
		if outputDir != "" {
			pkgParts := strings.Split(p.Name, "/")
			fileName := pkgParts[len(pkgParts)-1] + ".cipd"
			args = append(args, "pkg-build",
				"-out", filepath.Join(outputDir, fileName),
			)
			// Ensure outputDir exists. MkdirAll returns nil if path already exists.
			if err := os.MkdirAll(outputDir, 0777); err != nil {
				return errors.Annotate(err, "failed to create output directory %s", outputDir).Err()
			}
		} else {
			args = append(args,
				"create", "-verification-timeout", "60m",
			)
			for _, tag := range tags {
				args = append(args, "-tag", tag)
			}
			for _, ref := range refs {
				args = append(args, "-ref", strings.ToLower(ref))
			}
		}
		args = append(args, "-pkg-def", p.YamlPath)
		if serviceAccountJSON != "" {
			args = append(args, "-service-account-json", serviceAccountJSON)
		}

		logging.Infof(ctx, "Creating a CIPD package %s", p.Name)
		logging.Debugf(ctx, "Running cipd %s", strings.Join(args, " "))
		if err := RunCommand(ctx, "cipd", args...); err != nil {
			return errors.Annotate(err, "creating a CIPD package failed.").Err()
		}
		return nil
	}
	return builder
}

// PackageXcodeArgs are the parameters for PackageXcode() to keep them
// manageable.
type PackageXcodeArgs struct {
	xcodeAppPath       string
	cipdPackagePrefix  string
	serviceAccountJSON string
	outputDir          string
	skipRefTag         bool
	legacyIOSPackage   bool
}

func packageXcode(ctx context.Context, args PackageXcodeArgs) error {
	xcodeVersion, buildVersion, err := getXcodeVersion(filepath.Join(args.xcodeAppPath, "Contents", "version.plist"))
	if err != nil {
		return errors.Annotate(err, "this doesn't look like a valid Xcode.app folder: %s", args.xcodeAppPath).Err()
	}

	packages, err := makeXcodePackages(args.xcodeAppPath, args.cipdPackagePrefix, args.legacyIOSPackage)
	if err != nil {
		return err
	}
	tags := []string{
		"xcode_version:" + xcodeVersion,
		"build_version:" + buildVersion,
	}
	refs := []string{
		strings.ToLower(buildVersion), // Refs must match [a-z0-9_-]*
		"latest",
	}

	if args.skipRefTag {
		tags = []string{}
		refs = []string{}
	}

	buildFn := createBuilder(ctx, tags, refs, args.serviceAccountJSON, args.outputDir)

	if err = buildCipdPackages(packages, buildFn); err != nil {
		return err
	}

	fmt.Printf("\nCIPD packages:\n")
	for _, p := range packages {
		fmt.Printf("  %s  %s\n", p.Package, strings.ToLower(buildVersion))
	}

	return nil
}

// PackageRuntimeAndXcodeArgs are the parameters for packageRuntimeAndXcode() to
// keep them manageable.
type PackageRuntimeAndXcodeArgs struct {
	xcodeAppPath       string
	cipdPackagePrefix  string
	serviceAccountJSON string
	outputDir          string
	skipRefTag         bool
	legacyIOSPackage   bool
}

// Packages runtime & rest of Xcode.
func packageRuntimeAndXcode(ctx context.Context, args PackageRuntimeAndXcodeArgs) error {
	if !args.legacyIOSPackage {
		runtimePath := filepath.Join(args.xcodeAppPath, XcodeIOSSimulatorRuntimeRelPath, XcodeIOSSimulatorRuntimeFilename)
		packageRuntimeArgs := PackageRuntimeArgs{
			xcodeAppPath:       args.xcodeAppPath,
			runtimePath:        runtimePath,
			cipdPackagePrefix:  args.cipdPackagePrefix,
			serviceAccountJSON: args.serviceAccountJSON,
			outputDir:          args.outputDir,
			skipRefTag:         args.skipRefTag,
		}
		if err := packageRuntime(ctx, packageRuntimeArgs); err != nil {
			return errors.Annotate(err, "Error when packaging runtime.").Err()
		}
	}

	packageXcodeArgs := PackageXcodeArgs{
		xcodeAppPath:       args.xcodeAppPath,
		cipdPackagePrefix:  args.cipdPackagePrefix,
		serviceAccountJSON: args.serviceAccountJSON,
		outputDir:          args.outputDir,
		skipRefTag:         args.skipRefTag,
		legacyIOSPackage:   args.legacyIOSPackage,
	}
	if err := packageXcode(ctx, packageXcodeArgs); err != nil {
		return errors.Annotate(err, "Error when packaging rest of Xcode.").Err()
	}
	return nil
}

// PackageRuntimeArgs are the parameters for packageRuntime() to keep them
// manageable.
type PackageRuntimeArgs struct {
	xcodeAppPath       string
	runtimePath        string
	cipdPackagePrefix  string
	serviceAccountJSON string
	outputDir          string
	skipRefTag         bool
}

// Packages the iOS runtime named |runtimeFileName|(e.g. iOS.simruntime) under
// |runtimeDir|. |xcodeAppPath| is required when packaging a runtime that comes
// within Xcode package to properly set CIPD refs & tags.
func packageRuntime(ctx context.Context, args PackageRuntimeArgs) error {
	runtimeDir := filepath.Dir(args.runtimePath)
	runtimeFileName := args.runtimePath[strings.LastIndex(args.runtimePath, string(os.PathSeparator))+1:]

	xcodeBuildVersion := ""
	if args.xcodeAppPath != "" {
		var err error
		_, xcodeBuildVersion, err = getXcodeVersion(filepath.Join(args.xcodeAppPath, "Contents", "version.plist"))
		if err != nil {
			return errors.Annotate(err, "this doesn't look like a valid Xcode.app folder: %s", args.xcodeAppPath).Err()
		}
	}

	runtimeMakePackageArgs := MakePackageArgs{
		cipdPackageName:   IosRuntimePackageName,
		cipdPackagePrefix: args.cipdPackagePrefix,
		rootPath:          runtimeDir,
		includePrefixes:   []string{runtimeFileName},
		excludePrefixes:   []string{},
	}
	pkg, err := makePackage(runtimeMakePackageArgs)
	if err != nil {
		return errors.Annotate(err, "failed to create cipd package definition for %s/%s", runtimeDir, runtimeFileName).Err()
	}

	runtimeName, runtimeID, err := getSimulatorVersion(filepath.Join(runtimeDir, runtimeFileName, "Contents", "Info.plist"))
	if err != nil {
		return errors.Annotate(err, "failed to get simulator info from %s/%s/Contents/Info.plist", runtimeDir, runtimeFileName).Err()
	}

	tags := []string{
		"ios_runtime_version:" + runtimeName,
	}
	refs := []string{
		runtimeID + "_latest",
	}

	// Sets CIPD refs & tags according to whether the runtime is a default one
	// within Xcode package.
	if xcodeBuildVersion == "" {
		tags = append(tags,
			"type:manually_uploaded",
		)
		refs = append(refs,
			runtimeID,
		)
	} else {
		xcodeBuildVersion = strings.ToLower(xcodeBuildVersion)
		tags = append(tags,
			"xcode_build_version:"+xcodeBuildVersion,
			"type:xcode_default",
		)
		refs = append(refs,
			xcodeBuildVersion,
			runtimeID+"_"+xcodeBuildVersion,
		)
	}

	if args.skipRefTag {
		tags = []string{}
		refs = []string{}
	}

	buildFn := createBuilder(ctx, tags, refs, args.serviceAccountJSON, args.outputDir)

	if err = buildCipdPackages(Packages{runtimeID: pkg}, buildFn); err != nil {
		return err
	}

	fmt.Printf("\nCIPD package for simulator runtime:\n")
	fmt.Printf("  %s  %s %s\n", pkg.Package, runtimeID, xcodeBuildVersion)

	return nil
}
