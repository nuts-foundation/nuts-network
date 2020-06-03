package client

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/api"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
	"time"
)

// NewRegistryClient creates a new Local- or RemoteClient for the nuts registry
func NewNetworkClient() pkg.NetworkClient {
	instance := pkg.NetworkInstance()

	if core.NutsConfig().GetEngineMode(instance.Config.Mode) == core.ServerEngineMode {
		if err := instance.Configure(); err != nil {
			logrus.Panic(err)
		}

		return instance
	} else {
		return api.HttpClient{
			ServerAddress: instance.Config.Address,
			Timeout:       30 * time.Second,
		}
	}
}
