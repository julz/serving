/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"context"

	network "knative.dev/networking/pkg"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
)

type cfgKey struct{}

// +k8s:deepcopy-gen=false
type Config struct {
	Network *network.Config
}

func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

func (s *Store) Load() *Config {
	return &Config{
		Network: s.UntypedLoad(network.ConfigName).(*network.Config).DeepCopy(),
	}
}

func NewStore(ctx context.Context, onAfterStore ...func(name string, value interface{})) *Store {
	return &Store{
		UntypedStore: configmap.NewUntypedStore(
			"domainmapping",
			logging.FromContext(ctx),
			configmap.Constructors{
				network.ConfigName: network.NewConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}
}
