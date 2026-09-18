package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeSchemaSpec struct {
	Database string `json:"database"`
	Name     string `json:"name"`
	Comment  string `json:"comment,omitempty"`
}

type SnowflakeSchema struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeSchemaSpec `json:"spec"`
	Status            Status              `json:"status,omitempty"`
}

type SnowflakeSchemaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeSchema `json:"items"`
}

func (x *SnowflakeSchema) DeepCopyInto(o *SnowflakeSchema) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeSchema) DeepCopy() *SnowflakeSchema {
	if x == nil {
		return nil
	}
	o := new(SnowflakeSchema)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeSchema) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeSchemaList) DeepCopyInto(o *SnowflakeSchemaList) {
	*o = *x
	o.Items = append([]SnowflakeSchema(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeSchemaList) DeepCopy() *SnowflakeSchemaList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeSchemaList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeSchemaList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
