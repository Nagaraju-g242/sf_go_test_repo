package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeRoleSpec struct {
	Name          string `json:"name"`
	Comment       string `json:"comment,omitempty"`
	AutoReconcile bool   `json:"autoReconcile,omitempty"`
}

type SnowflakeRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeRoleSpec `json:"spec"`
	Status            Status            `json:"status,omitempty"`
}

type SnowflakeRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeRole `json:"items"`
}

func (x *SnowflakeRole) DeepCopyInto(o *SnowflakeRole) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeRole) DeepCopy() *SnowflakeRole {
	if x == nil {
		return nil
	}
	o := new(SnowflakeRole)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeRole) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeRoleList) DeepCopyInto(o *SnowflakeRoleList) {
	*o = *x
	o.Items = append([]SnowflakeRole(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeRoleList) DeepCopy() *SnowflakeRoleList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeRoleList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeRoleList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
