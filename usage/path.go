package usage

import (
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

func HomeDir() string {
	u, e := user.Current()
	if e != nil {
		return "./"
	}
	return u.HomeDir
}

func ExeName() string {
	var en string
	//获取执行文件名字
	exeFile, err := exec.LookPath(os.Args[0])
	if err != nil {
		return time.Now().String()
	} else {
		begin := strings.LastIndex(exeFile, "/") + 1
		end := len(exeFile)
		en = exeFile[begin:end]
	}
	return en
}

func CfgFile(suffix string) string {
	return ExeNameWithPath() + suffix
}

func ExeNameWithPath() string {
	//获取执行文件名字
	exeWithPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return time.Now().String()
	}
	return exeWithPath
}
