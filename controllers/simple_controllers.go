package controllers

import (
	"context"
	api "example.com/snowflake-operator/api/v1alpha1"
	sf "example.com/snowflake-operator/internal/snowflake"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o api.SnowflakeUser
	if e := r.Get(ctx, req.NamespacedName, &o); e != nil {
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}
	e := r.SF.EnsureUser(ctx, o.Spec.Name, o.Spec.DefaultRole, o.Spec.Comment)
	if e != nil {
		o.Status = api.Status{Message: e.Error(), ObservedGeneration: o.Generation}
		_ = r.Status().Update(ctx, &o)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, e
	}
	o.Status = api.Status{Ready: true, ObservedGeneration: o.Generation, Message: "Snowflake user is present"}
	return ctrl.Result{}, r.Status().Update(ctx, &o)
}
func (r *UserReconciler) SetupWithManager(m ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(m).For(&api.SnowflakeUser{}).Complete(r)
}

type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o api.SnowflakeDatabase
	if e := r.Get(ctx, req.NamespacedName, &o); e != nil {
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}
	e := r.SF.EnsureDatabase(ctx, o.Spec.Name, o.Spec.Comment)
	if e != nil {
		o.Status.Message = e.Error()
		_ = r.Status().Update(ctx, &o)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, e
	}
	o.Status = api.Status{Ready: true, ObservedGeneration: o.Generation, Message: "Snowflake database is present"}
	return ctrl.Result{}, r.Status().Update(ctx, &o)
}
func (r *DatabaseReconciler) SetupWithManager(m ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(m).For(&api.SnowflakeDatabase{}).Complete(r)
}

type SchemaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *SchemaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o api.SnowflakeSchema
	if e := r.Get(ctx, req.NamespacedName, &o); e != nil {
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}
	e := r.SF.EnsureSchema(ctx, o.Spec.Database, o.Spec.Name, o.Spec.Comment)
	if e != nil {
		o.Status.Message = e.Error()
		_ = r.Status().Update(ctx, &o)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, e
	}
	o.Status = api.Status{Ready: true, ObservedGeneration: o.Generation, Message: "Snowflake schema is present"}
	return ctrl.Result{}, r.Status().Update(ctx, &o)
}
func (r *SchemaReconciler) SetupWithManager(m ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(m).For(&api.SnowflakeSchema{}).Complete(r)
}

type WarehouseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *WarehouseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o api.SnowflakeWarehouse
	if e := r.Get(ctx, req.NamespacedName, &o); e != nil {
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}
	e := r.SF.EnsureWarehouse(ctx, o.Spec.Name, o.Spec.WarehouseSize, o.Spec.AutoSuspend, o.Spec.Comment)
	if e != nil {
		o.Status.Message = e.Error()
		_ = r.Status().Update(ctx, &o)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, e
	}
	o.Status = api.Status{Ready: true, ObservedGeneration: o.Generation, Message: "Snowflake warehouse is present"}
	return ctrl.Result{}, r.Status().Update(ctx, &o)
}
func (r *WarehouseReconciler) SetupWithManager(m ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(m).For(&api.SnowflakeWarehouse{}).Complete(r)
}

type GrantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	SF     *sf.Client
}

func (r *GrantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var o api.SnowflakeGrant
	if e := r.Get(ctx, req.NamespacedName, &o); e != nil {
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}
	e := r.SF.Grant(ctx, o.Spec.Privilege, o.Spec.On, o.Spec.Object, o.Spec.ToRole)
	if e != nil {
		o.Status.Message = e.Error()
		_ = r.Status().Update(ctx, &o)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, e
	}
	o.Status = api.Status{Ready: true, ObservedGeneration: o.Generation, Message: "Snowflake grant is present"}
	return ctrl.Result{}, r.Status().Update(ctx, &o)
}
func (r *GrantReconciler) SetupWithManager(m ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(m).For(&api.SnowflakeGrant{}).Complete(r)
}
