package lib

import (
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/uber-go/zap"
)

func getEnv(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

// GetGoPath ...
func GetGoPath() string {
	if goPath := getEnv("GOPATH"); goPath != "" {
		return goPath
	}
	return ""
}

// Config ...
type Config struct {
	Etcd     *EtcdConfig
	Server *ServerConfig
}

// ServerConfig ...
type ServerConfig struct {
	EnvPrefix        string `toml:"env_prefix"`
	APIUser          string `toml:"api_user"`
	APIPassword      string `toml:"api_password"`
	Region           string `toml:"region"`
	CertPath         string `toml:"cert_path"`
	VersionPath      string `toml:"version_path"`
	ClusterQuota     int    `toml:"cluster_quota"`
	EnableClusterCap bool   `toml:"enable_cluster_cap"`
	ClusterCap       int    `toml:"cluster_cap"`
	SoftlayerURL     string `toml:"softlayer_url"`
	AdminOrg         string `toml:"admin_org"`
	FreeDataCenter   string `toml:"free_datacenter"`
	FreePublicVlan   string `toml:"free_public_vlan"`
	FreePrivateVlan  string `toml:"free_private_vlan"`
}

// EtcdConfig ...
type EtcdConfig struct {
	Endpoints string `toml:"etcd_endpoints"`
	Auth      bool   `toml:"etcd_auth"`
	Secret    string `toml:"etcd_secret"`
	Username  string `toml:"etcd_user"`
	Password  string `toml:"etcd_password"`
}

// ParseConfig ...
//func ParseConfig(filePath string, conf interface{}, logger zap.Logger) {
func ParseConfig(filePath string, conf interface{}) {
	if _, err := toml.DecodeFile(filePath, conf); err != nil {
		//logger.Fatal("error parsing config file", zap.Error(err))
	}
}

// GetConfigString ...
func GetConfigString(envKey, defaultConf string) string {
	if val := getEnv(envKey); val != "" {
		return val
	}
	return defaultConf
}

// GetConfigInt ...
func GetConfigInt(envKey string, defaulfConf int, logger zap.Logger) int {
	if val := getEnv(envKey); val != "" {
		if envInt, err := strconv.Atoi(val); err == nil {
			return envInt
		}
		logger.Error("error parsing env val to int", zap.String("env", envKey))
	}
	return defaulfConf
}

// GetConfigBool ...
//func GetConfigBool(envKey string, defaultConf bool, logger zap.Logger) bool {
func GetConfigBool(envKey string, defaultConf bool) bool {
	if val := getEnv(envKey); val != "" {
		if envBool, err := strconv.ParseBool(val); err == nil {
			return envBool
		}
//		logger.Error("error parsing env val to bool", zap.String("env", envKey))
	}
	return defaultConf
}

// GetConfigStringList ...
//func GetConfigStringList(envKey string, defaultConf string, logger zap.Logger) []string {
func GetConfigStringList(envKey string, defaultConf string) []string {
	// Assume env var is a list of strings separated by ','
	val := defaultConf

	if getEnv(envKey) != "" {
		val = getEnv(envKey)
	}

	val = strings.Replace(val, " ", "", -1)
	return strings.Split(val, ",")
}
