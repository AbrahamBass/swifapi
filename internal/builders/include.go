package builders

import (
	"github.com/AbrahamBass/swifapi/internal/types"
)

type Include struct {
	app types.IApplication
}

func NewInclude(app types.IApplication) *Include {
	return &Include{
		app: app,
	}
}

func (i *Include) Add(rtrg func(types.IAPIRouter)) types.IIncludeBuilder {
	i.app.AddRouter(rtrg)
	return i
}

func (i *Include) Apply() types.IApplication {
	return i.app
}
