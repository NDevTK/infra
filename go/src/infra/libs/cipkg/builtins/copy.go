package builtins

import (
	"archive/tar"
	"context"
	"crypto"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"infra/libs/cipkg"
)

// TODO: copy can be merged into import as a special mode after fs.FS adding
// support for readlink: https://github.com/golang/go/issues/49580
// Otherwise we can't support copying from embed.FS requiring fs.FS interface
// and a normal directory which may contain symbolic link at same time.
const CopyFilesBuilder = BuiltinBuilderPrefix + "copyFiles"

var (
	copyFilesHashMap       = make(map[string]*CopyFiles)
	copyFilesHashAlgorithm = crypto.SHA256
)

type CopyFiles struct {
	Name  string
	Files fs.FS

	// By default, hash will be calculated for files as version. Manually
	// assigning the version means users are responsible for updating the
	// version when files changed.
	Version string
}

func (cf *CopyFiles) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	version := cf.Version
	if version == "" {
		h := copyFilesHashAlgorithm.New()
		if err := WalkDir(cf.Files, ".", h, func(path string, d fs.DirEntry, err error) error { return err }); err != nil {
			return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
		}
		version = fmt.Sprintf("%s:%x", copyFilesHashAlgorithm, h.Sum(nil))
	}
	copyFilesHashMap[version] = cf

	return cipkg.Derivation{
		Name:    cf.Name,
		Builder: CopyFilesBuilder,
		Args:    []string{version},
	}, cipkg.PackageMetadata{}, nil
}

func copyFiles(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:copyFiles", version]
	if len(cmd.Args) != 2 {
		return fmt.Errorf("invalid arguments: %v", cmd.Args)
	}
	out := GetEnv("out", cmd.Env)

	h := copyFilesHashAlgorithm.New()
	cf := copyFilesHashMap[cmd.Args[1]]
	if err := WalkDir(cf.Files, ".", h, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dst := filepath.Join(out, path)

		// Create path and return if it's directory
		if d.IsDir() {
			if err := os.MkdirAll(dst, os.ModePerm); err != nil {
				return fmt.Errorf("create dir failed: %s: %w", dst, err)
			}
			return nil
		}

		// Copy file content
		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("create dst file failed: %s: %w", dst, err)
		}
		defer dstFile.Close()

		srcFile, err := cf.Files.Open(path)
		if err != nil {
			return fmt.Errorf("open src file failed: %s: %w", dst, err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("copy file failed: %s: %w", dst, err)
		}

		// Change file mode
		info, err := fs.Stat(cf.Files, path)
		if err != nil {
			return fmt.Errorf("get file mode failed: %s: %w", path, err)
		}
		if err := dstFile.Chmod(info.Mode()); err != nil {
			return fmt.Errorf("chmod failed: %s: %w", dst, err)
		}

		return nil
	}); err != nil {
		return err
	}

	version := cf.Version
	if version == "" {
		version = fmt.Sprintf("%s:%x", copyFilesHashAlgorithm, h.Sum(nil))
	}
	if version != cmd.Args[1] {
		return fmt.Errorf("hash value mismatch: expected: %s, result: %s", cmd.Args[1], version)
	}
	return nil
}

// WalkDir behave almost same as fs.WalkDir, with extra hash.Hash argument used
// to calculate hash value from all the files in a directory.
func WalkDir(src fs.FS, root string, h hash.Hash, fn fs.WalkDirFunc) error {
	// Tar is used for calculating hash from files - including metadata - in a
	// simple way.
	tw := tar.NewWriter(h)
	defer tw.Close()

	return fs.WalkDir(src, root, func(name string, d fs.DirEntry, err error) error {
		if err := fn(name, d, err); err != nil {
			return err
		}

		info, err := fs.Stat(src, name)
		if err != nil {
			return fmt.Errorf("failed to stat file: %s: %w", name, err)
		}

		switch d.Type() {
		case fs.ModeSymlink:
			// We have to copy the file before fs.FS support readlink:
			// https://github.com/golang/go/issues/49580
			fallthrough
		case 0: // Regular File
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeReg,
				Name:     name,
				Mode:     int64(info.Mode()),
				Size:     info.Size(),
			}); err != nil {
				return fmt.Errorf("failed to write header: %s: %w", name, err)
			}
			f, err := src.Open(name)
			if err != nil {
				return fmt.Errorf("failed to open file: %s: %w", name, err)
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("failed to write file: %s: %w", name, err)
			}
		case fs.ModeDir:
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     name,
				Mode:     int64(info.Mode()),
			}); err != nil {
				return fmt.Errorf("failed to write header: %s: %w", name, err)
			}
		default:
			return fmt.Errorf("unsupported file type: %s: %s", name, d.Type())
		}
		return nil
	})
}

// Override the default FileMode. This can be used for embedded files since
// golang's embed strips all the mode bits from files.
type FSWithMode struct {
	fs.FS
	ModeOverride func(info fs.FileInfo) (fs.FileMode, error)
}

type fileInfoWithMode struct {
	fs.FileInfo
	mode fs.FileMode
}

func (fi *fileInfoWithMode) Mode() fs.FileMode {
	return fi.mode
}

func (ef *FSWithMode) Stat(name string) (fs.FileInfo, error) {
	f, err := ef.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	mode, err := ef.ModeOverride(info)
	if err != nil {
		return nil, err
	}
	return &fileInfoWithMode{info, mode}, nil
}
