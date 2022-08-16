package utils

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// 获取全部目录
func GetDirList(dir string, dirs []string) ([]string, error) {
	fs1, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range fs1 {
		if f.Type()&(fs.ModeSymlink|fs.ModeDir) != 0 {
			dirs = append(dirs, f.Name())
		}
	}

	return dirs, nil
}

// 创建目录
func CreateDir(path string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}

	return nil
}

// 是否是目录
func IsDirectory(path string) (bool, error) {
	if fi, err := os.Stat(path); err != nil && IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		if fi.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

func FileExists(filename string) (bool, error) {
	if fi, err := os.Stat(filename); err != nil && IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		if !fi.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

// 路径是否存在
func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil && !IsNotExist(err) {
		return false, err
	} else if err != nil {
		return false, nil
	}

	return true, nil
}

// 获取指定扩展名的全部文件
func GetFileList(dir, ext string, files []string) ([]string, error) {
	fs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range fs {
		if !f.IsDir() {
			if strings.HasSuffix(f.Name(), ext) {
				files = append(files, dir+"/"+f.Name())
			}
		}
	}

	return files, nil
}

func IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, fs.ErrNotExist)
}

func IsExist(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, fs.ErrExist)
}

func RemoveFile(path string) bool {
	if len(path) != 0 {
		err := os.Remove(path)
		if err != nil {
			return false
		}
	}

	return true
}
