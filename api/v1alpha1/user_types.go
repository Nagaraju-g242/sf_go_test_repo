package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeUserSpec struct {
	Name        string `json:"name"`
	DefaultRole string `json:"defaultRole,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type SnowflakeUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeUserSpec `json:"spec"`
	Status            Status            `json:"status,omitempty"`
}

type SnowflakeUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeUser `json:"items"`
}

func (x *SnowflakeUser) DeepCopyInto(o *SnowflakeUser) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeUser) DeepCopy() *SnowflakeUser {
	if x == nil {
		return nil
	}
	o := new(SnowflakeUser)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeUser) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeUserList) DeepCopyInto(o *SnowflakeUserList) {
	*o = *x
	o.Items = append([]SnowflakeUser(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeUserList) DeepCopy() *SnowflakeUserList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeUserList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeUserList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
