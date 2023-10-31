package global

import (
	"flag"
	"os"
	"path/filepath"
	"sync"
)

var conf = flag.String("conf", "", "config path contain config and template path")

func init() {
	flag.Parse()
	Init()
}

var RootDir string

var once = new(sync.Once)

func Init() {
	once.Do(func() {
		inferRootDir()
		initConfig()
	})
}

// inferRootDir 推断出项目根目录
func inferRootDir() {
	if exists(*conf+"/template") && exists(*conf+"/config") {
		RootDir = *conf
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		// 这里要确保项目根目录下存在 template 目录
		if exists(d + "/template") {
			return d
		}
		if exists(d + "/config") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	RootDir = infer(cwd)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
