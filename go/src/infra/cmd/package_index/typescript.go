// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/common/logging"

	kpb "infra/cmd/package_index/kythe/proto"
)

type tsTarget struct {
	imports     []string
	args        []string
	rootDir     string
	outDir      string
	corpus      string
	buildConfig string
	targetName  string
	tsconfig    string
	target      gnTargetInfo
	hashMap     *FileHashMap
	ctx         context.Context
}

func getTSConfig(target gnTargetInfo, ctx context.Context, rootDir, outDir string) (string, error) {
	for i := 0; i < len(target.Args)-1; i++ {
		if target.Args[i] == "--tsconfig_output_location" {
			tsconfig := target.Args[i+1]
			if strings.HasPrefix(tsconfig, "//") {
				gn, err := convertGnPath(ctx, tsconfig, outDir)
				if err != nil {
					return "", err
				}
				return filepath.Join(rootDir, outDir, gn), nil
			} else {
				return filepath.Join(rootDir, outDir, tsconfig), nil
			}
		}
	}
	return "", errNotSupported
}

func newTSTarget(ctx context.Context, gnTargetDict map[string]gnTargetInfo,
	targetName string, hashMap *FileHashMap, rootDir, outDir, corpus,
	buildConfig string) (*tsTarget, error) {
	target := gnTargetDict[targetName]
	tsconfig, err := getTSConfig(target, ctx, rootDir, outDir)
	if err != nil {
		return nil, err
	}
	m := &tsTarget{
		ctx:         ctx,
		targetName:  targetName,
		rootDir:     rootDir,
		outDir:      outDir,
		corpus:      corpus,
		buildConfig: buildConfig,
		hashMap:     hashMap,
		tsconfig:    tsconfig,
	}
	m.target = target
	m.args = gnTargetDict[m.targetName].Args
	m.imports = []string{}
	return m, nil
}

func (m *tsTarget) getUnit() (*kpb.CompilationUnit, error) {
	unitProto := &kpb.CompilationUnit{}
	var sourceFiles []string
	tsconfig, err := filepath.Rel(m.rootDir, m.tsconfig)
	if err != nil {
		return nil, err
	}
	sourceFiles = append(sourceFiles, convertPathToForwardSlashes(tsconfig))

	for _, src := range m.target.Sources {
		gn, err := convertGnPath(m.ctx, src, m.outDir)
		if err != nil {
			return nil, err
		}
		p := filepath.Join(m.outDir, gn)
		sourceFiles = append(sourceFiles, convertPathToForwardSlashes(p))
	}

	unitProto.Argument = append(unitProto.Argument, "@@"+convertPathToForwardSlashes(tsconfig))

	unitProto.SourceFile = sourceFiles
	unitProto.VName = &kpb.VName{Corpus: m.corpus, Language: "typescript"}

	for _, requiredFile := range sourceFiles {
		p, err := filepath.Abs(filepath.Join(m.rootDir, requiredFile))
		if err != nil {
			return nil, err
		}
		h, ok := m.hashMap.Filehash(p)
		if !ok {
			logging.Warningf(m.ctx, "Missing from filehashes %s\n", p)
			continue
		}

		vname := &kpb.VName{}
		setVnameForFile(vname, convertPathToForwardSlashes(requiredFile), m.corpus)
		requiredInput := &kpb.CompilationUnit_FileInput{
			VName: vname,
			Info: &kpb.FileInfo{
				Digest: h,
				Path:   convertPathToForwardSlashes(requiredFile),
			},
		}
		unitProto.RequiredInput = append(unitProto.GetRequiredInput(),
			requiredInput)
	}

	return unitProto, nil
}

func (m *tsTarget) getFiles() ([]string, error) {
	var dataFiles []string
	for _, src := range m.target.Sources {
		gn, err := convertGnPath(m.ctx, src, m.outDir)
		if err != nil {
			return nil, err
		}
		dataFiles = append(dataFiles, filepath.Join(m.rootDir, m.outDir, gn))
	}
	dataFiles = append(dataFiles, m.tsconfig)

	return dataFiles, nil
}

func tsTargetProcessor(ctx context.Context, rootPath, outDir, corpus,
	buildConfig string, hashMaps *FileHashMap, t *gnTarget) (
	GnTargetInterface, error) {

	if !isTSTarget(t) {
		return nil, errNotSupported
	}

	return newTSTarget(ctx, gnTargetsMap, t.targetName, hashMaps,
		rootPath, outDir, corpus, buildConfig)
}

func isTSTarget(t *gnTarget) bool {
	return t.targetInfo.Script == "//third_party/devtools-frontend/src/third_party/typescript/ts_library.py"
}
