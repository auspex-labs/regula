// Copyright 2021 Fugue, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

var defaultGitIgnorePatterns = []gitignore.Pattern{
	gitignore.ParsePattern("..", []string{}),
	gitignore.ParsePattern(".git", []string{}),
}

type Repo interface {
	IsPathIgnored(path string, isDir bool) bool
}

type repo struct {
	path          string
	ignoreMatcher gitignore.Matcher
}

func NewRepo(path string, inputTree *InputTreeNode) (Repo, error) {
	// fsys := osfs.New(path)
	patterns, err := ReadPatterns(path, inputTree, nil, nil, None)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, defaultGitIgnorePatterns...)
	return &repo{
		path:          path,
		ignoreMatcher: gitignore.NewMatcher(patterns),
	}, nil
}

func (r *repo) IsPathIgnored(path string, isDir bool) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	relPath, err := filepath.Rel(r.path, absPath)
	if err != nil {
		return false
	}
	splitPath := strings.Split(relPath, string(os.PathSeparator))
	return r.ignoreMatcher.Match(splitPath, isDir)
}

// RepoFinder finds the git repository for a given directory.
type RepoFinder struct {
	inputTree *InputTreeNode
	cache     map[string]Repo
}

// NewRepoFinder returns a new RepoFinder instance
func NewRepoFinder(inputPaths []string) *RepoFinder {
	inputTree := NewInputTree(inputPaths)
	return &RepoFinder{
		inputTree: inputTree,
		cache:     map[string]Repo{},
	}
}

// FindRepo takes a directory path and finds the git repository for it if one exists.
// It works by searching within the given directory, followed by searching in parent
// directories until it either reaches the top-level directory or encounters an error.
func (s *RepoFinder) FindRepo(path string) Repo {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	lastPath := ""
	traversedPaths := []string{}
	for absPath != lastPath {
		if foundRepo, ok := s.cache[absPath]; ok {
			return foundRepo
		}
		entries, err := os.ReadDir(absPath)
		if err != nil {
			// Store nil so that we don't retry this operation when child dirs are
			// passed in.
			s.cache[absPath] = nil
			return nil
		}
		for _, e := range entries {
			if e.Name() == ".git" {
				r, err := NewRepo(absPath, s.inputTree)
				if err != nil {
					s.cache[absPath] = nil
					return nil
				}
				s.cache[absPath] = r
				return s.cache[absPath]
			}
		}
		traversedPaths = append(traversedPaths, absPath)
		lastPath = absPath
		absPath = filepath.Dir(absPath)
	}
	// At this point we've traversed to the top of the tree and haven't found
	// anything. We'll cache all traversed paths so that we don't repeat the
	// list operations for child dirs.
	for _, p := range traversedPaths {
		s.cache[p] = nil
	}
	return nil
}

// Vendored from go-git so that we can fix their behavior
// readIgnoreFile reads a specific git ignore file.
func readIgnoreFile(prefix string, path []string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	absPath := append(append([]string{prefix}, path...), ignoreFile)
	joinPath := filepath.Join(absPath...)
	// fmt.Println("Trying to read .gitignore file", joinPath)
	f, err := os.Open(joinPath)
	if err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
				ps = append(ps, gitignore.ParsePattern(s, path))
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

// ReadPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher). This
// function has been modified to respect gitignore patterns while it's traversing. This
// has a big impact for larger repositories.
func ReadPatterns(prefix string, inputTree *InputTreeNode, path []string, accumulator []gitignore.Pattern, lastRelation Relation) ([]gitignore.Pattern, error) {
	ps, _ := readIgnoreFile(prefix, path, ".gitignore")
	accumulator = append(accumulator, ps...)

	absPath := append([]string{prefix}, path...)
	joinPath := filepath.Join(absPath...)
	// fmt.Println(joinPath)
	var fis []os.DirEntry
	fis, err := os.ReadDir(joinPath)
	if err != nil {
		return nil, err
	}

	matcher := gitignore.NewMatcher(accumulator)
	for _, fi := range fis {
		name := fi.Name()
		// fmt.Println(name)
		subPath := append(path, name)
		isDir := fi.IsDir()
		if matcher.Match(subPath, isDir) {
			continue
		}
		if isDir && name != ".git" {
			// fmt.Println("In iteration", joinPath, name)
			if lastRelation != IsChild {
				splitPath := strings.Split(prefix, string(os.PathSeparator))
				splitPath = append(splitPath, path...)
				splitPath = append(splitPath, name)
				lastRelation = inputTree.Relation(splitPath)
			}
			if lastRelation == None {
				// fmt.Println("lastRelation was None for", joinPath, name)
				continue
			}
			var subps []gitignore.Pattern
			subps, err = ReadPatterns(prefix, inputTree, subPath, accumulator, lastRelation)
			if err != nil {
				return nil, err
			}

			if len(subps) > 0 {
				ps = append(ps, subps...)
			}
		}
	}

	return ps, nil
}
