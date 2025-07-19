package usage

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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

// RunCurDir 获取当前路径
func RunCurDir() string {
	dir, err := filepath.Abs("./") //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

// RunParentDir 获取上层目录
func RunParentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败：%v\n", err)
		return ""
	}
	// 获取上层目录（父目录）
	return filepath.Dir(currentDir)
}
