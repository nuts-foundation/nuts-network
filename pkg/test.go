package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/sirupsen/logrus"
	"path"
)

func NewTestNetworkInstance(testDirectory string) *Network {
	config := TestNetworkConfig(testDirectory)
	newInstance := NewNetworkInstance(config, pkg.NewTestCryptoInstance(testDirectory))
	if err := newInstance.Configure(); err != nil {
		logrus.Fatal(err)
	}
	instance = newInstance
	return newInstance
}

func TestNetworkConfig(testDirectory string) NetworkConfig {
	config := DefaultNetworkConfig()
	config.StorageConnectionString = "file:" + path.Join(testDirectory, "network.db")
	config.PublicAddr = "test:5555"
	return config
}