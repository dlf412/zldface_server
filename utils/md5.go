package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: MD5V
//@description: md5加密
//@param: str []byte
//@return: string

func MD5V(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func MD5sum(r io.ReadSeeker) string {
	br := bufio.NewReader(r)
	h := md5.New()
	_, err := io.Copy(h, br)
	if err != nil {
		return ""
	}
	r.Seek(0, io.SeekStart)
	return hex.EncodeToString(h.Sum(nil))
}
