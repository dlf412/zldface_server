package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			err = os.MkdirAll(v, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func SaveBytes(src []byte, dst string) error {
	if err := CreateDir(path.Dir(dst)); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(src)
	return err
}

func SaveReader(src io.Reader, dst string) (int64, error) {
	if err := CreateDir(path.Dir(dst)); err != nil {
		return 0, err
	}
	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	return io.Copy(out, src)
}

func SaveFile(src io.ReadSeeker, dst string) error {
	_, err := SaveReader(src, dst)
	src.Seek(0, io.SeekStart)
	return err
}

func MD5RelativePath(r io.ReadSeeker) string {
	md5 := MD5sum(r)
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	return fmt.Sprintf("%s/%s/%s/%s.jpg", year, month, day, md5)
}
