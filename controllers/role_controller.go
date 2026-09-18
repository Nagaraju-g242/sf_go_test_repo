package controllers

import (
	"context"
	"time"

	api "example.com/snowflake-operator/api/v1alpha1"
	sf "example.com/snowflake-operator/internal/snowflake"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *RoleReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	start := time.Now()

	ReconcileTotal.WithLabelValues("role").Inc()

	defer func() {
		Duration.WithLabelValues("role").
			Observe(time.Since(start).Seconds())
	}()

	var obj api.SnowflakeRole

	if err := r.Get(
		ctx,
		req.NamespacedName,
		&obj,
	); err != nil {
		return ctrl.Result{},
			client.IgnoreNotFound(err)
	}

	if obj.Spec.Name == "" {
		return ctrl.Result{}, nil
	}

	err := r.SF.EnsureRole(
		ctx,
		obj.Spec.Name,
		obj.Spec.Comment,
	)

	if err != nil {
		ReconcileErrors.
			WithLabelValues("role").
			Inc()

		obj.Status = api.Status{
			Ready:              false,
			ObservedGeneration: obj.Generation,
			Message:            err.Error(),
		}

		_ = r.Status().Update(ctx, &obj)

		return ctrl.Result{
			RequeueAfter: 30 * time.Second,
		}, err
	}

	obj.Status = api.Status{
		Ready:              true,
		ObservedGeneration: obj.Generation,
		Message:            "Snowflake role is present",
	}

	if err := r.Status().Update(ctx, &obj); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 60 * time.Second,
	}, nil
}

func (r *RoleReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.SnowflakeRole{}).
		Complete(r)
}