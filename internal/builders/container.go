package builders

import (
	"github.com/AbrahamBass/swiftapi/internal/types"

	"go.uber.org/zap"
)

type Container struct {
	app types.IApplication
	dig types.IDigContainer
}

func NewDi(app types.IApplication, dig types.IDigContainer) *Container {
	return &Container{
		app: app,
		dig: dig,
	}
}

func (c *Container) Provide(constructor interface{}) types.IContainerBuilder {
	err := c.dig.Provide(constructor)
	if err != nil {
		c.app.GetLogger().Fatal("🚨 dependency registration",
			zap.String("msg", err.Error()),
		)
	}
	return c
}

func (d *Container) Apply() types.IApplication {
	return d.app
}
