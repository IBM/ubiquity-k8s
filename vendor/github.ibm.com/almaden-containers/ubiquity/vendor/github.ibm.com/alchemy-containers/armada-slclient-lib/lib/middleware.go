package lib

import (
//	"context"
	"path/filepath"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	//"github.com/uber-go/zap"
	"time"
)


// NewEtcdClient ...
//func NewEtcdClient(conf *Config, logger zap.Logger) *EtcdAPI {
func NewEtcdClient(conf *Config) *EtcdAPI {
	authRequired := GetConfigBool("ETCD_AUTH", conf.Etcd.Auth) //, logger)
	var user, password string
	var clientTransport client.CancelableTransport
	if authRequired {
		user = GetConfigString("ETCD_USER", conf.Etcd.Username)
		password = GetConfigString("ETCD_PASSWORD", conf.Etcd.Password)
		if user == "" {
			//logger.Fatal("no auth found for etcd client")
		}
		caPath := GetConfigString("CERT_PATH", conf.Server.CertPath)
		tlsInfo := transport.TLSInfo{
			CAFile: filepath.Join(caPath, "compose-ca.pem"),
		}
		var err error
		if clientTransport, err = transport.NewTransport(tlsInfo, time.Second*5); err != nil {
			//logger.Fatal("error setting auth for etcd client", zap.Error(err))
		}
	} else {
		clientTransport = client.DefaultTransport
	}
	cfg := client.Config{
		Endpoints: GetConfigStringList("ETCD_ENDPOINTS", conf.Etcd.Endpoints), //, logger),
		Transport: clientTransport,
		Username:  user,
		Password:  password,
		// Timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second * 5,
	}
	c, err := client.New(cfg)
	if err != nil {
		//logger.Fatal("error creating etcd client", zap.Error(err))
	}
	kapi := client.NewKeysAPI(c)
	return &EtcdAPI{kapi} //, logger}
}

// EtcdAPI ...
type EtcdAPI struct {
	client.KeysAPI
//	Logger          zap.Logger
}
