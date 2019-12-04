// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"path/filepath"
	"testing"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type TestDummySuite struct {
	test.Test
}

func TestDummy(t *testing.T) {
	suite.Run(t, &TestDummySuite{})
}

func (s *TestDummySuite) TestCp() {
	s.Run("should copy the file correctly", func() {
		fs = afero.NewMemMapFs()
		src := "/test/subdir/1.txt"
		dst := "/minio/bucket-1"

		s.Require().NoError(fs.MkdirAll(dst, 0755))
		s.Require().NoError(afero.WriteFile(fs, src, []byte("ok boomer"), 0644))

		s.NoError(cp(src, dst))

		b, err := afero.ReadFile(fs, filepath.Join(dst, src))
		s.NoError(err)
		s.Equal([]byte("ok boomer"), b)
	})

	s.Run("should copy a directory correctly", func() {
		fs = afero.NewMemMapFs()
		src := "/test/subdir"
		dst := "/minio/bucket-1"

		s.Require().NoError(fs.MkdirAll(dst, 0755))
		s.Require().NoError(afero.WriteFile(fs, filepath.Join(src, "1.txt"), []byte("ok boomer"), 0644))
		s.Require().NoError(afero.WriteFile(fs, filepath.Join(src, "2.txt"), []byte("ok b00mer"), 0644))

		s.NoError(cp(filepath.Dir(src), dst))

		b, err := afero.ReadFile(fs, filepath.Join(dst, src, "1.txt"))
		s.NoError(err)
		s.Equal([]byte("ok boomer"), b)

		b, err = afero.ReadFile(fs, filepath.Join(dst, src, "2.txt"))
		s.NoError(err)
		s.Equal([]byte("ok b00mer"), b)
	})

	s.Run("should return an error if the src file doesn't exist", func() {
		fs = afero.NewMemMapFs()
		src := "/test/subdir/1.txt"
		dst := "/minio/bucket-1"

		s.EqualError(cp(src, dst), "open /test/subdir/1.txt: file does not exist")
	})
}
