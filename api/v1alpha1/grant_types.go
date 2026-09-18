package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeGrantSpec struct {
	Privilege string `json:"privilege"`
	On        string `json:"on"`
	Object    string `json:"object"`
	ToRole    string `json:"toRole"`
}

type SnowflakeGrant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeGrantSpec `json:"spec"`
	Status            Status             `json:"status,omitempty"`
}

type SnowflakeGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeGrant `json:"items"`
}

func (x *SnowflakeGrant) DeepCopyInto(o *SnowflakeGrant) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeGrant) DeepCopy() *SnowflakeGrant {
	if x == nil {
		return nil
	}
	o := new(SnowflakeGrant)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeGrant) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeGrantList) DeepCopyInto(o *SnowflakeGrantList) {
	*o = *x
	o.Items = append([]SnowflakeGrant(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeGrantList) DeepCopy() *SnowflakeGrantList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeGrantList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeGrantList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
