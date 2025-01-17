/*
Copyright 2019 The Knative Authors

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

package sequence

import (
	"context"

	"github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	"github.com/knative/eventing/pkg/duck"
	"github.com/knative/eventing/pkg/reconciler"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	"github.com/knative/eventing/pkg/client/injection/informers/eventing/v1alpha1/subscription"
	"github.com/knative/eventing/pkg/client/injection/informers/messaging/v1alpha1/sequence"
)

const (
	// ReconcilerName is the name of the reconciler
	ReconcilerName = "Sequences"
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "sequence-controller"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	sequenceInformer := sequence.Get(ctx)
	subscriptionInformer := subscription.Get(ctx)
	addressableInformer := duck.NewAddressableInformer(ctx)

	r := &Reconciler{
		Base:               reconciler.NewBase(ctx, controllerAgentName, cmw),
		sequenceLister:     sequenceInformer.Lister(),
		subscriptionLister: subscriptionInformer.Lister(),
	}
	impl := controller.NewImpl(r, r.Logger, ReconcilerName)

	r.Logger.Info("Setting up event handlers")

	r.addressableTracker = addressableInformer.NewTracker(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	sequenceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Register handler for Subscriptions that are owned by Sequence, so that
	// we get notified if they change.
	subscriptionInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Sequence")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
