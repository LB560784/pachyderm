/*

directory structure

  .
  |-- repositoryName
      |-- __root__
	      |-- shardNum // the writeable root created from btrfs snapshot create on InitRepository
	  |-- scratch
		  |-- shardNum // the read-only read created on InitRepository, this is where to start branching
      |-- commitID
	      |-- shardNum // this is where subvolumes are

*/
package btrfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/executil"
)

type driver struct {
	rootDir string
}

func newDriver(rootDir string) *driver {
	return &driver{rootDir}
}

func (d *driver) Init() error {
	return nil
}

func (d *driver) InitRepository(repository *pfs.Repository, shards map[int]bool) error {
	// syscall.Mkdir (which os.Mkdir directly calls) is atomic across processes, and since
	// we do not include shards as part of the repository path, this guaranteees
	// only one initialization will complete successfully
	if err := os.Mkdir(d.repositoryPath(repository), 0700); err != nil {
		return err
	}
	systemRootCommit := &pfs.Commit{
		Repository: repository,
		Id:         drive.SystemRootCommitID,
	}
	if err := os.Mkdir(d.commitPathNoShard(systemRootCommit), 0700); err != nil {
		return err
	}
	initialCommit := &pfs.Commit{
		Repository: repository,
		Id:         drive.InitialCommitID,
	}
	if err := os.Mkdir(d.commitPathNoShard(initialCommit), 0700); err != nil {
		return err
	}
	for shard := range shards {
		if err := subvolumeCreate(d.commitPath(systemRootCommit, shard)); err != nil {
			return err
		}
		if err := subvolumeSnapshotReadonly(d.commitPath(systemRootCommit, shard), d.commitPath(initialCommit, shard)); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) GetFile(path *pfs.Path, shard int) (io.ReadCloser, error) {
	return os.Open(d.filePath(path, shard))
}

func (d *driver) MakeDirectory(path *pfs.Path, shard int) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	return os.MkdirAll(d.filePath(path, shard), 0700)
}

func (d *driver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	file, err := os.Create(d.filePath(path, shard))
	if err != nil {
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(file)
	return err
}

func (d *driver) ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error) {
	return nil, nil
}

func (d *driver) GetParent(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error) {
	if newCommit == nil {
		newCommit = &pfs.Commit{
			Repository: commit.Repository,
			Id:         drive.NewCommitID(),
		}
	}
	if err := os.Mkdir(d.commitPathNoShard(newCommit), 0700); err != nil {
		return nil, err
	}
	for shard := range shards {
		commitPath := d.commitPath(commit, shard)
		readOnly, err := isReadOnly(commitPath)
		if err != nil {
			return nil, err
		}
		if !readOnly {
			return nil, fmt.Errorf("%+v is not a read-only snapshot", commit)
		}
		newCommitPath := d.commitPath(newCommit, shard)
		if err := subvolumeSnapshot(commitPath, newCommitPath); err != nil {
			return nil, err
		}
	}
	return newCommit, nil
}

func (d *driver) Commit(commit *pfs.Commit, shards map[int]bool) error {
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (d *driver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (d *driver) GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) repositoryPath(repository *pfs.Repository) string {
	return filepath.Join(d.rootDir, repository.Name)
}

func (d *driver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repositoryPath(commit.Repository), commit.Id)
}

func (d *driver) commitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprintf("%d", shard))
}

func (d *driver) filePath(path *pfs.Path, shard int) string {
	return filepath.Join(d.commitPath(path.Commit, shard), path.Path)
}

func isReadOnly(path string) (bool, error) {
	reader, err := snapshotPropertyGet(path)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "ro=true") {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func snapshotPropertyGet(path string) (io.Reader, error) {
	return executil.RunStdout("btrfs", "property", "get", "-t", "s", path)
}

func subvolumeCreate(path string) error {
	return executil.Run("btrfs", "subvolume", "create", path)
}

func subvolumeSnapshot(src string, dest string) error {
	return executil.Run("btrfs", "subvolume", "snapshot", src, dest)
}

func subvolumeSnapshotReadonly(src string, dest string) error {
	return executil.Run("btrfs", "subvolume", "snapshot", "-r", src, dest)
}
