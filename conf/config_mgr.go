/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package conf

import (
	"io/ioutil"
	"path/filepath"
	"github.com/json-iterator/go"
	"sync"
)

var defaultCfgPath = "./config/srvConf.json"

type LogConfig struct {
	LogLevel     string `json:"logLevel"`
	LogPath      string `json:"logPath"`
	MaxAge       int    `json:"maxAge"`       //分钟, -1 无限制
	RotationTime int    `json:"rotationTime"` //分钟
}

type ServerConfig struct {
	Name          string                `json:"name"`
	Host          string                `json:"host"`
	Port          int                   `json:"port"`
	WebPort       int                   `json:"webPort"`
	MaxConn       int                   `json:"maxConn"`
	MaxPacketSize int                   `json:"maxPacketSize"`
	Log           map[string]*LogConfig `json:"log"`
	Remotes       []string              `json:"remotes"`
	HostList      []string              `json:"hostList"`
	Timeout       int                   `json:"timeout"`
	WatchServerId string                `json:"watchServerId"`
}

func (sc *ServerConfig) GetDefaultLogConf() *LogConfig {
	d, ok := sc.Log["default"]
	if !ok {
		return nil
	}
	return d
}

type ServerCfgMgr struct {
	conf map[string]*ServerConfig
}

var defaultServerCfgMgr *ServerCfgMgr
var cfgOnce sync.Once

func InitCfg(path string) {
	cfgOnce.Do(func() {
		defaultServerCfgMgr = newServerCfgMgr()
	})
	if defaultServerCfgMgr == nil {
		panic("init config failed")
	}
	defaultServerCfgMgr.LoadConf(path)
}

func Get(srvName string) *ServerConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.Get(srvName)
}

func newServerCfgMgr() *ServerCfgMgr {
	return &ServerCfgMgr{
		conf: make(map[string]*ServerConfig),
	}
}

func (scm *ServerCfgMgr) LoadConf(path string) {
	if scm.conf == nil {
		scm.conf = make(map[string]*ServerConfig)
	}
	data, err := LoadFile(path)
	if err != nil {
		panic(err)
	}
	err = jsoniter.Unmarshal(data, &scm.conf)
	if err != nil {
		panic(err)
	}
}

func (scm *ServerCfgMgr) Get(key string) *ServerConfig {
	if scm.conf == nil {
		return nil
	}
	v, ok := scm.conf[key]
	if !ok {
		return nil
	}
	return v
}

func LoadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func GetConfPath(path ...string) string {
	dir, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return ""
	}
	if len(path) > 0 {
		return filepath.Join(dir, path[0])
	}
	return filepath.Join(dir, defaultCfgPath)
}
