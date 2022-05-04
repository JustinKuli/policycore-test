/*
Copyright 2022.

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

package metrics

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	policycore "github.com/JustinKuli/policycore-test/api/v1"
)

const ControllerName string = "compliance_metrics"

var log = ctrl.Log.WithName(ControllerName)

// SetupWithManager sets up the controller with the Manager, and registers the
// metrics gauges in prometheus
func SetupWithManager(mgr ctrl.Manager, policyKinds ...schema.GroupVersionKind) error {
	if len(policyKinds) == 0 {
		// TODO: Better error
		return fmt.Errorf("must provide at least one policy type to report metrics on")
	}

	complianceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ocm_policy_compliance",
			Help: "The compliance status of the named policy. 0 == Compliant. 1 == NonCompliant. " +
				"2 == UnknownCompliancy.",
		},
		[]string{
			"policy_kind",      // The kind of the policy
			"policy_name",      // The name of the policy
			"policy_namespace", // The namespace where the policy is defined
		},
	)

	if err := metrics.Registry.Register(complianceGauge); err != nil {
		return err
	}

	// Each policy type gets its own controller. This lets us use a generic reconcile function, and
	// inject the GVK for each type. Note that the controllers will all share the cache, so this
	// probably isn't too bad for performance.
	for _, policyGVK := range policyKinds {
		policyObj := unstructured.Unstructured{}
		policyObj.SetGroupVersionKind(policyGVK)

		err := ctrl.NewControllerManagedBy(mgr).
			Named(ControllerName + "_" + policyGVK.Kind).
			For(&policyObj).
			Complete(reconcileMetric(mgr.GetClient(), policyGVK, complianceGauge))
		if err != nil {
			return err
		}
	}

	return nil
}

func reconcileMetric(r client.Reader, gvk schema.GroupVersionKind, gauge *prometheus.GaugeVec) reconcile.Func {
	return reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		log := log.V(1).WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name,
			"Request.Kind", gvk.Kind)
		log.Info("Reconciling metric for the policy")

		promLabels := prometheus.Labels{
			"policy_kind":      gvk.Kind,
			"policy_name":      req.Name,
			"policy_namespace": req.Namespace,
		}

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		if err := r.Get(ctx, req.NamespacedName, u); err != nil {
			if errors.IsNotFound(err) {
				// Try to delete the gauge, but don't get hung up on errors.
				// Just log whether the gauge was deleted or not.
				statusGaugeDeleted := gauge.Delete(promLabels)
				log.Info("Policy not found. It must have been deleted.", "status-gauge-deleted", statusGaugeDeleted)
				return reconcile.Result{}, nil
			}

			log.Error(err, "Failed to get policy")
			return reconcile.Result{}, err
		}

		statusMetric, err := gauge.GetMetricWith(promLabels)
		if err != nil {
			log.Error(err, "Failed to get status metric from GaugeVec")

			return reconcile.Result{}, err
		}

		// TODO: Can we use policyCore.ObjectWithCompliance instead of Unstructured?
		compliance, ok := u.Object["status"].(map[string]interface{})["compliant"].(string)
		if !ok {
			log.V(1).Info("Couldn't get compliance, using policycore.UnknownCompliancy")
			compliance = string(policycore.UnknownCompliancy)
		}

		log.Info("Setting metric based on compliance", "compliance", compliance)

		switch compliance {
		case string(policycore.Compliant):
			statusMetric.Set(0)
		case string(policycore.NonCompliant):
			statusMetric.Set(1)
		default:
			statusMetric.Set(2)
		}

		return reconcile.Result{}, nil
	})
}
