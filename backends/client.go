package backends

import (
	"errors"
	"github.com/kelseyhightower/confd/log"
	"strings"

	"github.com/kelseyhightower/confd/backends/consul"
	"github.com/kelseyhightower/confd/backends/dynamodb"
	"github.com/kelseyhightower/confd/backends/env"
	"github.com/kelseyhightower/confd/backends/etcdv3"
	"github.com/kelseyhightower/confd/backends/file"
	"github.com/kelseyhightower/confd/backends/nacos"
	"github.com/kelseyhightower/confd/backends/rancher"
	"github.com/kelseyhightower/confd/backends/redis"
	"github.com/kelseyhightower/confd/backends/ssm"
	"github.com/kelseyhightower/confd/backends/vault"
	"github.com/kelseyhightower/confd/backends/zookeeper"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {

	if config.Backend == "" {
		config.Backend = "etcd"
	}
	backendNodes := config.BackendNodes

	if config.Backend == "file" {
		log.Info("Backend source(s) set to " + strings.Join(config.YAMLFile, ", "))
	} else {
		log.Info("Backend source(s) set to " + strings.Join(backendNodes, ", "))
	}

	switch config.Backend {
	case "consul":
		return consul.New(config.BackendNodes, config.Scheme,
			config.ClientCert, config.ClientKey,
			config.ClientCaKeys,
			config.BasicAuth,
			config.Username,
			config.Password,
		)
	case "etcd":
		// etcd v2 has been deprecated and etcdv3 is now the client for both the etcd and etcdv3 backends.
		return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "etcdv3":
		return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "zookeeper":
		return zookeeper.NewZookeeperClient(backendNodes)
	case "rancher":
		return rancher.NewRancherClient(backendNodes)
	case "redis":
		return redis.NewRedisClient(backendNodes, config.ClientKey, config.Separator)
	case "env":
		return env.NewEnvClient()
	case "file":
		return file.NewFileClient(config.YAMLFile, config.Filter)
	case "vault":
		vaultConfig := map[string]string{
			"app-id":    config.AppID,
			"user-id":   config.UserID,
			"role-id":   config.RoleID,
			"secret-id": config.SecretID,
			"username":  config.Username,
			"password":  config.Password,
			"token":     config.AuthToken,
			"cert":      config.ClientCert,
			"key":       config.ClientKey,
			"caCert":    config.ClientCaKeys,
			"path":      config.Path,
		}
		return vault.New(backendNodes[0], config.AuthType, vaultConfig)
	case "dynamodb":
		table := config.Table
		log.Info("DynamoDB table set to " + table)
		return dynamodb.NewDynamoDBClient(table)
	case "ssm":
		return ssm.New()
	case "nacos":
		return nacos.NewNacosClient(backendNodes, config.Group, constant.ClientConfig{
			NamespaceId: config.Namespace,
			AccessKey:   config.AccessKey,
			SecretKey:   config.SecretKey,
			Endpoint:    config.Endpoint,
			Username:    config.Username,
			Password:    config.Password,
		})
	}
	return nil, errors.New("Invalid backend")
}
