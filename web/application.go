package web

import (
	"catuan/comm"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"os"
	"time"
)

type Application struct {
	*gin.Engine
	version    string
	activeEnv  string // active environment
	configPath string
	runPath    string

	roles  map[string]RoleInf
	groups map[string]GroupInf

	envPropertyHooks []AppPropertyHook
	envProperties    map[string]string

	cdbChain    []*gorm.DB
	credisChain []*redis.Client

	appConf *AppConfInfo
}

type AppPropertyHook func(envProperty map[string]string)

// Default args: 0 - env, 1 - config file path
func Default(args ...string) *Application {
	activeEnv := "dev"
	configPath := ""
	if len(args) == 1 {
		activeEnv = args[0]
	}
	if len(args) == 2 {
		activeEnv = args[0]
		configPath = args[1]
	}
	app := &Application{
		Engine:     gin.Default(),
		version:    "1.0.0",
		activeEnv:  activeEnv,
		roles:      make(map[string]RoleInf),
		groups:     make(map[string]GroupInf),
		configPath: configPath,

		cdbChain:    make([]*gorm.DB, 0),
		credisChain: make([]*redis.Client, 0),

		envPropertyHooks: make([]AppPropertyHook, 0),
	}
	app.loadConfigFile() //加载配置文件
	return app
}

func New(args ...string) *Application {
	activeEnv := "dev"
	configPath := ""
	if len(args) == 1 {
		activeEnv = args[0]
	}
	if len(args) == 2 {
		activeEnv = args[0]
		configPath = args[1]
	}
	app := &Application{
		Engine:     gin.Default(),
		version:    "1.0.0",
		activeEnv:  activeEnv,
		roles:      make(map[string]RoleInf),
		groups:     make(map[string]GroupInf),
		configPath: configPath,

		cdbChain:    make([]*gorm.DB, 0),
		credisChain: make([]*redis.Client, 0),

		envPropertyHooks: make([]AppPropertyHook, 0),
	}
	app.loadConfigFile() //加载配置文件
	return app
}

func (a *Application) Version() string {
	return a.version
}

func (a *Application) ActiveEnv() string {
	return a.activeEnv
}

// loadConfigFile 加载配置文件
func (a *Application) loadConfigFile() {
	rootPath, _ := os.Getwd()
	a.runPath = rootPath
	fileName := "application.yaml"
	fileInfo := rootPath + "/resources/" + fileName
	if a.configPath != "" {
		fileInfo = a.configPath + "/" + fileName
	}
	f, err := os.OpenFile(fileInfo, os.O_RDONLY, 0666)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	content, err := io.ReadAll(f)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	a.appConf = &AppConfInfo{}
	err = yaml.Unmarshal(content, a.appConf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file_content": string(content),
			"tip":          "配置文件解析失败",
			"fileInfo":     fileInfo,
		}).Error(err.Error())
		return
	}
}

// EnvProperty 获取环境变量
func (a *Application) EnvProperty(key string) (string, bool) {
	if a.appConf != nil && a.appConf.Env != nil {
		val, ok := a.appConf.Env[key]
		return val, ok
	}
	return "", false
}

func (a *Application) InitDB() {
	if a.appConf != nil && a.appConf.Mysql != nil {
		for _, info := range a.appConf.Mysql.Database {
			db, err := a.mysqlConnect(*info)
			if err != nil {
				logrus.Error(err.Error())
				continue
			}
			if a.appConf.Mysql.Debug == 1 {
				db = db.Debug()
			}
			a.cdbChain = append(a.cdbChain, db)
		}
	}
}

func (a *Application) InitRedis() {
	if a.appConf != nil && a.appConf.Redis != nil {
		for _, info := range a.appConf.Redis {
			rdb := redis.NewClient(&redis.Options{
				Addr:     info.Host + ":" + info.Port,
				Password: info.Password, // no password set
				DB:       info.Db,       // use default DB
			})
			a.credisChain = append(a.credisChain, rdb)
		}
	}
}

func (a *Application) SetCDB(cdb *gorm.DB, index int) {
	if len(a.cdbChain) == 0 || index >= len(a.cdbChain) {
		a.cdbChain = append(a.cdbChain, cdb)
	} else {
		a.cdbChain[index] = cdb
	}
}

func (a *Application) Init() {
	a.runEnvPropertyHook()
	a.InitDB()
	a.InitRedis()
}

func (a *Application) runEnvPropertyHook() {
	if a.envPropertyHooks != nil && len(a.envPropertyHooks) > 0 {
		for _, hook := range a.envPropertyHooks {
			if a.appConf.Env != nil {
				hook(a.appConf.Env)
			}
		}
	}
}

// 数据库链接
func (a *Application) mysqlConnect(info DatabaseInfo) (*gorm.DB, error) {
	charset := "utf8mb4"
	if info.Charset != "" {
		charset = info.Charset
	}
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", info.User, info.Password, info.Host, info.Port, info.Name, charset)
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (a *Application) FindRole(roleLabel string) (RoleInf, bool) {
	roleInf, ok := a.roles[roleLabel]
	return roleInf, ok
}

func (a *Application) UseRole(roles ...RoleInf) {
	for _, role := range roles {
		_, ok := a.roles[role.RoleLabel()]
		if ok {
			panic("role already exists")
		}
		a.roles[role.RoleLabel()] = role
	}
}

func (a *Application) FindGroup(roleLabel, groupLabel string) (GroupInf, bool) {
	key := roleLabel + "-" + groupLabel
	groupInf, ok := a.groups[key]
	return groupInf, ok
}

func (a *Application) UseGroup(groups ...GroupInf) {
	for _, group := range groups {
		key := group.RoleLabel() + "-" + group.GroupLabel()
		_, ok := a.groups[key]
		if ok {
			panic("group already exists")
		}
		a.groups[group.GroupLabel()] = group
	}
}

func (a *Application) HttpServerRun() error {
	defaultPort := "8080"
	if a.appConf != nil && a.appConf.Web != nil {
		if a.appConf.Web.Http.Port != "" {
			defaultPort = a.appConf.Web.Http.Port
		}
	}
	err := a.Run(":" + defaultPort)
	return err
}

func (a *Application) HttpsServerRun() error {
	defaultPort := "8443"
	defaultCertFile := a.runPath + "/resources/cert/server.crt"
	defaultKeyFile := a.runPath + "/resources/cert/server.key"
	if a.appConf != nil && a.appConf.Web != nil {
		if a.appConf.Web.Https.Port != "" {
			defaultPort = a.appConf.Web.Https.Port
		}
		if a.appConf.Web.Https.CertFile != "" {
			defaultCertFile = a.appConf.Web.Https.CertFile
		}
		if a.appConf.Web.Https.KeyFile != "" {
			defaultKeyFile = a.appConf.Web.Https.KeyFile
		}
	}
	err := a.RunTLS(":"+defaultPort, defaultCertFile, defaultKeyFile)
	return err
}

func (a *Application) Router(ctx context.Context, c Context) {
	defaultTimeout := time.Second * 5
	go func() {
		defer func() {
			// recover from panic
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"role":   c.RoleLabel(),
					"group":  c.GroupLabel(),
					"action": c.ActionLabel(),
				}).Error(err)
			}
		}()
		a.router(c)
	}()
	select {
	case <-time.After(defaultTimeout):
		c.JsonResponse(&comm.RespResult{ErrCode: -1, ErrMsg: "timeout"})
		return
	case resp := <-c.RespChannel():
		c.JsonResponse(resp)
	case <-ctx.Done():
		return
	}
}

func (a *Application) router(c Context) {
	role, ok := a.FindRole(c.RoleLabel())
	if !ok {
		c.Result(-1, "access denied,role not found")
		return
	}
	group, ok := a.FindGroup(c.RoleLabel(), c.GroupLabel())
	if !ok {
		c.Result(-1, "access denied,group not found")
		return
	}
	if !ok {
		c.Result(-1, "access denied,action not found")
		return
	}
	role.Call(c)
	if !c.IsNext() {
		return
	}
	group.Call(c)
	if !c.IsNext() {
		return
	}
}

func (a *Application) DBDefault() *gorm.DB {
	if a.cdbChain == nil || len(a.cdbChain) == 0 {
		return nil
	}
	return a.cdbChain[0]
}

func (a *Application) GetDB(i int) *gorm.DB {
	if a.cdbChain == nil || len(a.cdbChain)-1 <= i {
		return nil
	}
	return a.cdbChain[i]
}

func (a *Application) RedisDefault() *redis.Client {
	if a.credisChain == nil || len(a.credisChain) == 0 {
		return nil
	}
	return a.credisChain[0]
}

func (a *Application) GetRedis(i int) *redis.Client {
	if a.credisChain == nil || len(a.credisChain)-1 <= i {
		return nil
	}
	return a.credisChain[i]
}
