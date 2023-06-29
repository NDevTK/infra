// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"infra/libs/cipkg_new/core"

	"github.com/spf13/afero"
	"go.chromium.org/luci/common/system/environ"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Executor[M proto.Message] func(ctx context.Context, msg M, dstFS afero.Fs) error

const envCipkgExec = "_CIPKG_EXEC_CMD"

// actions.Main adds _CIPKG_EXEC_CMD to main binary for reexec.
// If _CIPKG_EXEC_CMD is specified, an executor registered to the main module
// will be called.
type Main struct {
	execs map[protoreflect.FullName]Executor[proto.Message]

	mu     sync.Mutex
	sealed bool
}

// TODO(fancl): allow custom extension.
func NewMain() *Main {
	m := &Main{
		execs: make(map[protoreflect.FullName]Executor[proto.Message]),
	}
	MustSetExecutor[*core.ActionReexec](m, ActionReexecExecutor)
	MustSetExecutor[*core.ActionURLFetch](m, ActionURLFetchExecutor)
	MustSetExecutor[*core.ActionFilesCopy](m, defaultFilesCopyExecutor.Execute)
	MustSetExecutor[*core.ActionCIPDExport](m, ActionCIPDExportExecutor)
	return m
}

var (
	ErrExecutorExisted  = errors.New("executor for the message type already existed")
	ErrReexecMainSealed = errors.New("executor can't be set after main being used")
)

// SetExecutor set the executor for the action specification M.
func SetExecutor[M proto.Message](m *Main, execFunc Executor[M]) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sealed {
		return ErrExecutorExisted
	}

	var msg M
	name := proto.MessageName(msg)
	if _, ok := m.execs[name]; ok {
		return ErrTransformerExisted
	}
	m.execs[name] = func(ctx context.Context, msg proto.Message, dstFS afero.Fs) error {
		return execFunc(ctx, msg.(M), dstFS)
	}
	return nil
}

// MustSetExecutor set the executor for the action specification M and
// panic if any error happened.
func MustSetExecutor[M proto.Message](m *Main, execFunc Executor[M]) {
	if err := SetExecutor[M](m, execFunc); err != nil {
		panic(err)
	}
}

// Execute the main function. This is REQUIRED for reexec to function properly
// and need to be executed after init() because embed fs or other resources may
// be registered in init().
// Any application using the framework should use Main.Run(...) to wrap around
// the real main function.
func (m *Main) Run(main func()) {
	m.RunWithArgs(afero.NewOsFs(), environ.System(), os.Args, main)
}

// Execute the main function with environment and args.
func (m *Main) RunWithArgs(dstFS afero.Fs, env environ.Env, args []string, main func()) {
	m.mu.Lock()
	if !m.sealed {
		m.sealed = true
	}
	m.mu.Unlock()
	if _, ok := env.Lookup(envCipkgExec); !ok {
		main()
	}

	if len(args) < 2 {
		panic(fmt.Sprintf("usage: cipkg-exec <proto>: insufficient args: %s", args))
	}

	var any anypb.Any
	if err := protojson.Unmarshal([]byte(args[1]), &any); err != nil {
		panic(fmt.Sprintf("failed to unmarshal anypb: %s, %s", err, args))
	}
	msg, err := any.UnmarshalNew()
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal proto from any: %s, %s", err, args))
	}
	f := m.execs[proto.MessageName(msg)]
	if f == nil {
		panic(fmt.Sprintf("unknown cipkg-exec command: %s", args))
	}

	ctx := env.SetInCtx(context.Background())
	out := NewBasePathFs(dstFS, env.Get("out"))
	if err := f(ctx, msg, out); err != nil {
		log.Fatalln(err)
	}
}

// TODO(fancl): Use runtime/debug for git commit/module version if possible.
var reexecFixed = fmt.Sprintf("now:%s", time.Now().UTC())

// ActionReexecTransformer is the default transformer for reexec action.
func ActionReexecTransformer(a *core.ActionReexec, deps []PackageDependency) (*core.Derivation, error) {
	m, err := anypb.New(a)
	if err != nil {
		return nil, err
	}

	b, err := protojson.Marshal(m)
	if err != nil {
		return nil, err
	}

	self, err := os.Executable()
	if err != nil {
		return nil, err
	}

	env := environ.New(nil)
	env.Set(envCipkgExec, "1")

	return &core.Derivation{
		Args:        []string{self, string(b)},
		FixedOutput: reexecFixed,
		Env:         env.Sorted(),
	}, nil
}

// ActionReexecExecutor is the default executor for reexec action.
func ActionReexecExecutor(ctx context.Context, a *core.ActionReexec, dstFS afero.Fs) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	src, err := os.Open(self)
	if err != nil {
		return err
	}
	defer src.Close()

	dstFile := "cipkg_exec"
	if runtime.GOOS == "windows" {
		dstFile += ".exe"
	}
	dst, err := dstFS.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	if err := dstFS.Chmod(dstFile, os.ModePerm); err != nil {
		return err
	}

	return nil
}

const reexecActionName = "_cipkg_exec"

// ReexecDependency returns a default action dependency for any action require
// re-executing the binary.
func ReexecDependency() *core.Action_Dependency {
	return &core.Action_Dependency{
		Action: &core.Action{
			Metadata: &core.Action_Metadata{Name: reexecActionName},
			Spec:     &core.Action_Reexec{},
		},
		Runtime: false,
	}
}

var (
	ErrReexecNotFound = errors.New("reexec not found in dependencies")
)

// ReexecDerivation returns a derivation for re-executing the self-duplicated
// binary. It sets the FixedOutput using hash generated from action spec.
func ReexecDerivation(m proto.Message, deps []PackageDependency, hostEnv bool) (*core.Derivation, error) {
	if len(deps) == 0 {
		return nil, ErrReexecNotFound
	}
	var cipkgExec string
	for _, d := range deps {
		if d.Package.Action.Metadata.Name == reexecActionName {
			cipkgExec = filepath.Join(d.Package.Handler.OutputDirectory(), "cipkg_exec")
			if runtime.GOOS == "windows" {
				cipkgExec += ".exe"
			}
		}
	}
	if cipkgExec == "" {
		return nil, ErrReexecNotFound
	}

	m, err := anypb.New(m)
	if err != nil {
		return nil, err
	}
	b, err := protojson.Marshal(m)
	if err != nil {
		return nil, err
	}

	fixed, err := sha256String(m)
	if err != nil {
		return nil, err
	}

	env := environ.New(nil)
	if hostEnv {
		env = environ.System()
	}
	env.Set(envCipkgExec, "1")

	return &core.Derivation{
		Args:        []string{cipkgExec, string(b)},
		Env:         env.Sorted(),
		FixedOutput: fixed,
	}, nil
}

func sha256String(m proto.Message) (string, error) {
	const algo = crypto.SHA256
	h := algo.New()
	if err := core.StableHash(h, m); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%x", algo, h.Sum(nil)), nil
}
