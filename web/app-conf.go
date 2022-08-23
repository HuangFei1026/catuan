package web

type AppConfInfo struct {
	Version   string            `yaml:"version"`
	ActiveEnv string            `yaml:"activeEnv"`
	Web       *WebConfInfo      `yaml:"web"`
	Mysql     *MysqlConfInfo    `yaml:"mysql"`
	Redis     []*RedisConfInfo  `yaml:"redis"`
	Log       *LogConfInfo      `yaml:"log"`
	Env       map[string]string `yaml:"env"`
}

type WebConfInfo struct {
	WriteTimeout int       `yaml:"writeTimeout"` // 写入超时时间 5 秒
	Http         *HttpInfo `yaml:"http"`
	Https        *HttpInfo `yaml:"https"`
}

type HttpInfo struct {
	Port     string `yaml:"port"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type MysqlConfInfo struct {
	Debug    int8            `yaml:"debug"`
	Database []*DatabaseInfo `yaml:"database"`
}

type DatabaseInfo struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
}

type RedisConfInfo struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Db       int    `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type LogConfInfo struct {
	Level    string `yaml:"level"`    //级别
	FilePath string `yaml:"filePath"` //文件路径
}
