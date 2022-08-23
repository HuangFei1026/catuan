package test

import (
	"catuan/web"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"testing"
)

func TestYamlConf(t *testing.T) {
	f, err := os.OpenFile("./conf.yaml", os.O_RDONLY, 0666)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	defer f.Close()

	bodyData, err := io.ReadAll(f)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	conf := web.AppConfInfo{}
	err = yaml.Unmarshal(bodyData, &conf)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	fmt.Println(conf)
}
