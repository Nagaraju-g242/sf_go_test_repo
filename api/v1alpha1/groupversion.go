package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{
	Group:   "snowflake.example.com",
	Version: "v1alpha1",
}

var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&SnowflakeRole{},
		&SnowflakeRoleList{},
		&SnowflakeUser{},
		&SnowflakeUserList{},
		&SnowflakeDatabase{},
		&SnowflakeDatabaseList{},
		&SnowflakeSchema{},
		&SnowflakeSchemaList{},
		&SnowflakeWarehouse{},
		&SnowflakeWarehouseList{},
		&SnowflakeGrant{},
		&SnowflakeGrantList{},
	)

	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

func AddToScheme(scheme *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(scheme)
}
