package main

import (
	"log"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	api "example.com/snowflake-operator/api/v1alpha1"
	ctrlrs "example.com/snowflake-operator/controllers"
	sf "example.com/snowflake-operator/internal/snowflake"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	log.Println("START: snowflake operator")

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		log.Fatalf("failed adding Kubernetes scheme: %v", err)
	}

	if err := api.AddToScheme(scheme); err != nil {
		log.Fatalf("failed adding Snowflake scheme: %v", err)
	}

	log.Println("START: creating controller manager")

	mgr, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme: scheme,
		},
	)
	if err != nil {
		log.Fatalf("failed creating manager: %v", err)
	}

	log.Println("START: manager created")

	log.Println("START: creating Snowflake client")

	db, err := sf.New()
	if err != nil {
		log.Fatalf("Snowflake client creation failed: %v", err)
	}
	defer db.Close()

	log.Println("START: Snowflake client created")

	log.Println("START: registering DatabaseReconciler")

	if err := (&ctrlrs.DatabaseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		SF:     db,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("failed registering DatabaseReconciler: %v", err)
	}

	log.Println("START: DatabaseReconciler registered")

	log.Println("START: registering SchemaReconciler")

	if err := (&ctrlrs.SchemaReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		SF:     db,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("failed registering SchemaReconciler: %v", err)
	}

	log.Println("START: SchemaReconciler registered")

	log.Println("START: registering WarehouseReconciler")

	if err := (&ctrlrs.WarehouseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		SF:     db,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("failed registering WarehouseReconciler: %v", err)
	}

	log.Println("START: WarehouseReconciler registered")

	log.Println("START: registering RoleReconciler")

	if err := (&ctrlrs.RoleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		SF:     db,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("failed registering RoleReconciler: %v", err)
	}

	log.Println("START: RoleReconciler registered")

	log.Println("START: starting controller manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("manager stopped: %v", err)
	}
}
