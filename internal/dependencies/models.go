package dependencies

import "reflect"

type param struct {
	I           int
	Name        string
	ReflectType reflect.Type
}

type model struct {
	I           int
	Name        string
	Type        reflect.Type
	ReflectType reflect.Type
	Default     any
}

func newModel(i int, name string, typ reflect.Type, rflTyp reflect.Type) *model {
	return &model{
		I:           i,
		Type:        typ,
		Name:        name,
		ReflectType: rflTyp,
		Default:     nil,
	}
}

type dependant struct {
	PathParams    []*model
	QueryParams   []*model
	BodyParams    []*model
	CookieParams  []*model
	HeaderParams  []*model
	FormParams    []*model
	ServiceParams []*model
	ContextParams []*model
	Request       *model
	Response      *model
	Websocket     *model
}

func newDependant() *dependant {
	return &dependant{
		PathParams:    []*model{},
		QueryParams:   []*model{},
		BodyParams:    []*model{},
		CookieParams:  []*model{},
		HeaderParams:  []*model{},
		FormParams:    []*model{},
		ServiceParams: []*model{},
		ContextParams: []*model{},
	}
}
