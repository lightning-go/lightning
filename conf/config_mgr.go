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

var defaultCfgPath = "config/srvConf.json"

var defaultServerCfgMgr *ServerCfgMgr
var cfgOnce sync.Once

func InitCfg(path string) {
	cfgOnce.Do(func() {
		defaultServerCfgMgr = newServerCfgMgr()
		defaultServerCfgMgr.LoadConf(path)
	})
	if defaultServerCfgMgr == nil {
		panic("init config failed")
	}
}

func GetServer(srvName string) *ServerConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetServer(srvName)
}

func GetDB(srvName string) *DBConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetDB(srvName)
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
	Timeout       int64                 `json:"timeout"`
	Group         string                `json:"group"`
	WatchGroup    string                `json:"watchGroup"`
}

func (sc *ServerConfig) GetDefaultLogConf() *LogConfig {
	d, ok := sc.Log["default"]
	if !ok {
		return nil
	}
	return d
}

type DBConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}

type ServerCfgMgr struct {
	mux     sync.RWMutex
	Servers map[string]*ServerConfig
	Db      map[string]*DBConfig
}

func newServerCfgMgr() *ServerCfgMgr {
	return &ServerCfgMgr{
		Servers: make(map[string]*ServerConfig),
		Db:      make(map[string]*DBConfig),
	}
}

func (scm *ServerCfgMgr) LoadConf(path string) {
	data, err := LoadFile(path)
	if err != nil {
		panic(err)
	}

	scm.mux.Lock()
	defer scm.mux.Unlock()

	err = jsoniter.Unmarshal(data, scm)
	if err != nil {
		panic(err)
	}
}

func (scm *ServerCfgMgr) GetServer(key string) *ServerConfig {
	if scm.Servers == nil {
		return nil
	}

	scm.mux.RLock()
	v, ok := scm.Servers[key]
	scm.mux.RUnlock()

	if !ok {
		return nil
	}
	return v
}

func (scm *ServerCfgMgr) GetDB(key string) *DBConfig {
	if scm.Db == nil {
		return nil
	}

	scm.mux.RLock()
	v, ok := scm.Db[key]
	scm.mux.RUnlock()

	if !ok {
		return nil
	}
	return v
}
