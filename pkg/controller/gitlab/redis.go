/*
Copyright 2019 The GitLab-Controller Authors.

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

package gitlab

import (
	"context"

	xpcachev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/cache/v1alpha1"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/gitlab-controller/pkg/apis/controller/v1alpha1"
)

const (
	redisClaimKind = "redis"
)

// redisReconciler
type redisReconciler struct {
	*baseResourceReconciler
	resourceClassFinder resourceClassFinder
}

func (r *redisReconciler) reconcile(ctx context.Context) error {
	ref, err := r.resourceClassFinder.find(ctx, r.GetProviderRef(), xpcachev1alpha1.RedisClusterKindAPIVersion)
	if err != nil {
		return errors.Wrapf(err, errorFmtFailedToFindResourceClass, r.getClaimKind(), r.GetProviderRef())
	}

	red := &xpcachev1alpha1.RedisCluster{
		ObjectMeta: r.newObjectMeta(),
		Spec: xpcachev1alpha1.RedisClusterSpec{
			ClassRef:      ref,
			EngineVersion: redisEngineVersion,
		},
	}
	key := r.newNamespacedName()

	if err := r.client.Get(ctx, key, red); err != nil {
		if kerrors.IsNotFound(err) {
			return errors.Wrapf(r.client.Create(ctx, red), errorFmtFailedToCreate, r.getClaimKind(), key)
		}
		return errors.Wrapf(err, errorFmtFailedToRetrieveInstance, r.getClaimKind(), key)
	}

	r.status = &red.Status
	return nil
}

func (r *redisReconciler) getClaimKind() string {
	return redisClaimKind
}

var _ resourceReconciler = &redisReconciler{}

func newRedisReconciler(gitlab *v1alpha1.GitLab, client client.Client) *redisReconciler {
	base := newBaseComponentReconciler(gitlab, client)
	return &redisReconciler{
		baseResourceReconciler: base,
		resourceClassFinder:    base,
	}
}