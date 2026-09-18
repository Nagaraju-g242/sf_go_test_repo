package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeDatabaseSpec struct {
	Name    string `json:"name"`
	Comment string `json:"comment,omitempty"`
}

type SnowflakeDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeDatabaseSpec `json:"spec"`
	Status            Status                `json:"status,omitempty"`
}

type SnowflakeDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeDatabase `json:"items"`
}

func (x *SnowflakeDatabase) DeepCopyInto(o *SnowflakeDatabase) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeDatabase) DeepCopy() *SnowflakeDatabase {
	if x == nil {
		return nil
	}
	o := new(SnowflakeDatabase)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeDatabase) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeDatabaseList) DeepCopyInto(o *SnowflakeDatabaseList) {
	*o = *x
	o.Items = append([]SnowflakeDatabase(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeDatabaseList) DeepCopy() *SnowflakeDatabaseList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeDatabaseList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeDatabaseList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
