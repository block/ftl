package watch

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"github.com/bmatcuk/doublestar/v4"
)

type FileChangeType rune

func (f FileChangeType) String() string { return string(f) }
func (f FileChangeType) GoString() string {
	switch f {
	case FileAdded:
		return "buildengine.FileAdded"
	case FileRemoved:
		return "buildengine.FileRemoved"
	case FileChanged:
		return "buildengine.FileChanged"
	default:
		panic("unknown file change type")
	}
}

const (
	FileAdded   FileChangeType = '+'
	FileRemoved FileChangeType = '-'
	FileChanged FileChangeType = '*'
)

type FileHashes map[string][]byte

// CompareFileHashes compares the hashes of the files in the oldFiles and newFiles maps.
//
// Returns all file changes
func CompareFileHashes(oldFiles, newFiles FileHashes) []FileChange {
	changes := []FileChange{}
	for key, hash1 := range oldFiles {
		hash2, exists := newFiles[key]
		if !exists {
			changes = append(changes, FileChange{Change: FileRemoved, Path: key})
		}
		if !bytes.Equal(hash1, hash2) {
			changes = append(changes, FileChange{Change: FileChanged, Path: key})
		}
	}

	for key := range newFiles {
		if _, exists := oldFiles[key]; !exists {
			changes = append(changes, FileChange{Change: FileAdded, Path: key})
		}
	}

	return changes
}

// ComputeFileHashes computes the SHA256 hash of all files in the given directory.
//
// If skipGitIgnoredFiles is true, files that are ignored by git will be skipped.
func ComputeFileHashes(dir string, skipGitIgnoredFiles bool, patterns []string) (FileHashes, error) {
	// Watch paths are allowed to be outside the deploy directory.
	fileHashes := make(FileHashes)
	rootDirs := computeRootDirs(dir, patterns)

	for _, rootDir := range rootDirs {
		err := WalkDir(rootDir, skipGitIgnoredFiles, func(srcPath string, entry fs.DirEntry) error {
			if entry.IsDir() {
				return nil
			}
			hash, matched, err := computeFileHash(rootDir, srcPath, patterns)
			if err != nil {
				return errors.WithStack(err)
			}
			if !matched {
				if patterns[0] == "*" {
					return errors.Errorf("file %s:%s does not match any: %s", rootDir, srcPath, patterns)
				}
				return nil
			}
			fileHashes[srcPath] = hash
			return nil
		})

		if err != nil {
			return nil, errors.Wrapf(err, "could not compute file hashes for %s", rootDir)
		}
	}

	return fileHashes, nil
}

func computeFileHash(baseDir, srcPath string, patterns []string) (hash []byte, matched bool, err error) {
	relativePath, err := filepath.Rel(baseDir, srcPath)
	if err != nil {
		return nil, false, errors.Wrap(err, "could not calculate relative path while computing filehash")
	}
	for _, pattern := range patterns {
		match, err := doublestar.PathMatch(pattern, relativePath)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		if !match {
			continue
		}

		file, err := os.Open(srcPath)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}

		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			_ = file.Close()
			return nil, false, errors.Wrap(err, "could not hash file")
		}

		hash := hasher.Sum(nil)

		if err := file.Close(); err != nil {
			return nil, false, errors.Wrap(err, "could not close file after hashing")
		}
		return hash, true, nil
	}
	return nil, false, nil
}

// computeRootDirs computes the unique root directories for the given baseDir and patterns.
func computeRootDirs(baseDir string, patterns []string) []string {
	uniqueRoots := make(map[string]struct{})
	uniqueRoots[baseDir] = struct{}{}

	for _, pattern := range patterns {
		fullPath := filepath.Join(baseDir, pattern)
		dirPath, _ := doublestar.SplitPattern(fullPath)
		cleanedPath := filepath.Clean(dirPath)

		if _, err := os.Stat(cleanedPath); err == nil {
			uniqueRoots[cleanedPath] = struct{}{}
		}
	}

	roots := make([]string, 0, len(uniqueRoots))
	for root := range uniqueRoots {
		roots = append(roots, root)
	}

	return roots
}
