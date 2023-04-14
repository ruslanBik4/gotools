/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package storable

import (
	"os"
	"path"

	"github.com/pkg/errors"
)

// MakeSrcDir create with destination directory 'dst'
func MakeSrcDir(dst string) error {
	err := os.MkdirAll(dst, os.ModePerm)

	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "mkDirAll")
	}

	return nil
}

// MakeSrcFile create/open file '{name}.go' inside 'dst'
func MakeSrcFile(dst, name string) (*os.File, error) {
	f, err := os.Create(path.Join(dst, name) + ".go")
	if err != nil && !os.IsExist(err) {
		// err.(*os.PathError).Err
		return nil, errors.Wrap(err, "creator")
	}

	return f, nil
}

// CreateFile create/open file 'name.ext' inside 'dst'
func CreateFile(dst, name, ext string) (*os.File, error) {
	f, err := os.Create(path.Join(dst, name) + "." + ext)
	if err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "creator")
	}

	return f, nil
}

// CreateDirPathAndWriteFile make dir on path 'dst' & write slice b in file 'name.ext' inside
func CreateDirPathAndWriteFile(dst string, name, ext string, b []byte) error {
	err := MakeSrcDir(dst)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(dst, name)+"."+ext, b, os.ModePerm)
}

// CreateEmptyDirAndFile  make empty dir on path 'dst' & create file 'name.ext' inside
func CreateEmptyDirAndFile(dst string, name, ext string) (*os.File, error) {

	if err := os.RemoveAll(dst); err != nil {
		return nil, err
	}

	if err := MakeSrcDir(dst); err != nil {
		return nil, err
	}

	return CreateFile(dst, name, ext)

}

// CreateDirPathAndFile make dir on path 'dst' & create file 'name.ext' inside
func CreateDirPathAndFile(dst string, name, ext string) (*os.File, error) {
	err := MakeSrcDir(dst)
	if err != nil {
		return nil, err
	}

	return CreateFile(dst, name, ext)
}
