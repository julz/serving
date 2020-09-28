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

package domainmapping

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	network "knative.dev/networking/pkg"
	"knative.dev/serving/pkg/client/injection/informers/serving/v1alpha1/domainmapping"
	kindreconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1alpha1/domainmapping"
	"knative.dev/serving/pkg/reconciler/domainmapping/config"
)

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)
	domainmappingInformer := domainmapping.Get(ctx)

	r := &Reconciler{}
	impl := kindreconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
		configsToResync := []interface{}{
			&network.Config{},
		}
		resync := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
			impl.GlobalResync(domainmappingInformer.Informer())
		})
		configStore := config.NewStore(logging.WithLogger(ctx, logger.Named("config-store")), resync)
		configStore.WatchConfigs(cmw)
		return controller.Options{ConfigStore: configStore}
	})

	logger.Info("Setting up event handlers")
	domainmappingInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
