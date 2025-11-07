package goravelgcs

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.gcs"

var App foundation.Application

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			Binding,
		},
		Dependencies: []string{
			binding.Config,
		},
		ProvideFor: []string{
			binding.Storage,
		},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		gcs := NewGCS()
		gcs.disk = parameters["disk"].(string)
		return gcs, nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/edoaurahman/goravel-gcs", map[string]string{
		"config/gcs.go": app.ConfigPath(""),
	})
}
