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

func InitCfg(path ...string) {
	GetDefalutServerCfgMgr().LoadConf(path...)
}

func GetDefalutServerCfgMgr() *ServerCfgMgr {
	cfgOnce.Do(func() {
		defaultServerCfgMgr = NewServerCfgMgr()
	})
	if defaultServerCfgMgr == nil {
		panic("init config failed")
	}
	return defaultServerCfgMgr
}

func GetDefaultLogConf() *LogConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetLogCfg("default")
}

func GetLogCfg(logName string) *LogConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetLogCfg(logName)
}

func GetSrvCfg(srvName string) *ServerConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetSrvCfg(srvName)
}

func GetDBCfg(srvName string) *DBConfig {
	if defaultServerCfgMgr == nil {
		return nil
	}
	return defaultServerCfgMgr.GetDBCfg(srvName)
}

func GetServerName() string {
	if defaultServerCfgMgr == nil {
		return ""
	}
	return defaultServerCfgMgr.GetServerName()
}

func GetServerId() string {
	if defaultServerCfgMgr == nil {
		return ""
	}
	return defaultServerCfgMgr.GetServerId()
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
	LogFile		 string	`json:"logFile"`
	MaxAge       int    `json:"maxAge"`       //minute, -1 unlimited
	RotationTime int    `json:"rotationTime"` //minute
}

type ServerConfig struct {
	Name          string   `json:"name"`
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	WebHost       string   `json:"webHost"`
	WebPort       int      `json:"webPort"`
	MaxConn       int      `json:"maxConn"`
	MaxPacketSize int      `json:"maxPacketSize"`
	Remotes       []string `json:"remotes"`
	HostList      []string `json:"hostList"`
	Timeout       int64    `json:"timeout"`
	Group         string   `json:"group"`
	WatchGroups   []string `json:"watchGroups"`
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
	ServerId 	string				 `json:"server_id"`
	ServerName 	string				 `json:"server_name"`
	Servers map[string]*ServerConfig `json:"servers"`
	Db      map[string]*DBConfig     `json:"db"`
	Log     map[string]*LogConfig    `json:"log"`
}

func NewServerCfgMgr() *ServerCfgMgr {
	return &ServerCfgMgr{
		Servers: make(map[string]*ServerConfig),
		Db:      make(map[string]*DBConfig),
		Log:     make(map[string]*LogConfig),
	}
}

func (scm *ServerCfgMgr) LoadConf(path ...string) {
	scm.mux.Lock()
	defer scm.mux.Unlock()

	for _, file := range path {
		data, err := LoadFile(file)
		if err != nil {
			panic(err)
		}
		err = jsoniter.Unmarshal(data, scm)
		if err != nil {
			panic(err)
		}
	}
}

func (scm *ServerCfgMgr) GetLogCfg(logName string) *LogConfig {
	scm.mux.RLock()
	defer scm.mux.RUnlock()

	d, ok := scm.Log[logName]
	if !ok || d == nil {
		return nil
	}
	return d
}

func (scm *ServerCfgMgr) GetSrvCfg(key string) *ServerConfig {
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

func (scm *ServerCfgMgr) GetDBCfg(key string) *DBConfig {
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

func (scm *ServerCfgMgr) GetServerName() string {
	if scm.Db == nil {
		return ""
	}
	return scm.ServerName
}


func (scm *ServerCfgMgr) GetServerId() string {
	if scm.Db == nil {
		return ""
	}
	return scm.ServerId
}

