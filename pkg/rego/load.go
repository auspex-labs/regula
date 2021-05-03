package rego

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

//go:embed lib
var regulaLib embed.FS

//go:embed rules
var regulaRules embed.FS

var loadExts map[string]bool = map[string]bool{
	".rego": true,
	// TODO: We should evaluate how useful it is for end-users to load non-rego files
	// in their rules. We'll need to change how these files get loaded into OPA in
	// order to support these other extensions.
	// ".yaml": true,
	// ".yml":  true,
	// ".json": true,
}

type regoFile struct {
	path     string
	contents []byte
}

func (r *regoFile) Raw() []byte {
	return r.contents
}

func (r *regoFile) String() string {
	return string(r.contents)
}

func (r *regoFile) AstModule() (*ast.Module, error) {
	return ast.ParseModule(r.Path(), r.String())
}

func (r *regoFile) RegoModule() func(r *rego.Rego) {
	return rego.Module(r.Path(), r.String())
}

func (r *regoFile) Path() string {
	return r.path
}

func newRegoFile(fsys fs.FS, path string) (RegoFile, error) {
	contents, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	return &regoFile{
		path:     path,
		contents: contents,
	}, nil
}

func loadDirectory(fsys fs.FS, path string, cb func(r RegoFile) error) error {
	walkDirFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); !loadExts[ext] {
			return nil
		}
		regoFile, err := newRegoFile(fsys, path)
		if err != nil {
			return err
		}
		if err := cb(regoFile); err != nil {
			return err
		}
		return nil
	}

	if err := fs.WalkDir(fsys, path, walkDirFunc); err != nil {
		return err
	}

	return nil
}

func loadOsFiles(paths []string, cb func(r RegoFile) error) error {
	fsys := &osFs{}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			err := loadDirectory(fsys, path, cb)
			if err != nil {
				return err
			}
			continue
		}
		file, err := newRegoFile(fsys, path)
		if err != nil {
			return err
		}
		if err := cb(file); err != nil {
			return err
		}
	}
	return nil
}

func loadRegula(userOnly bool, cb func(r RegoFile) error) error {
	if err := loadDirectory(regulaLib, "lib", cb); err != nil {
		return err
	}
	if !userOnly {
		if err := loadDirectory(regulaRules, "rules", cb); err != nil {
			return err
		}
	}

	return nil
}

// I might be missing something, but it looks like the only fs.FS implementation
// with os methods is os.DirFS, which has behavior that we don't want.
type osFs struct {
	fs.FS
	fs.GlobFS
	fs.ReadDirFS
	fs.ReadFileFS
	fs.StatFS
}

func (o *osFs) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (o *osFs) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (o *osFs) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (o *osFs) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *osFs) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}
