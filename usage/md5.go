package usage

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
)

func Md5(input []byte) string {
	hash := md5.Sum(input)
	md5str := fmt.Sprintf("%x", hash)
	return md5str
}

func Md5Str(input string) string {
	data := []byte(input)
	return Md5(data)
}

func FileMd5(fn string) string {
	file, e := os.Open(fn)
	if e != nil {
		fmt.Errorf("CheckMd5, open(%s) err:%v", fn, e)
		return ""
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	return Md5(data)
}
