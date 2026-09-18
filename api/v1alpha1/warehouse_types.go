package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SnowflakeWarehouseSpec struct {
	Name          string `json:"name"`
	WarehouseSize string `json:"warehouseSize,omitempty"`
	AutoSuspend   int    `json:"autoSuspend,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

type SnowflakeWarehouse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnowflakeWarehouseSpec `json:"spec"`
	Status            Status                 `json:"status,omitempty"`
}

type SnowflakeWarehouseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowflakeWarehouse `json:"items"`
}

func (x *SnowflakeWarehouse) DeepCopyInto(o *SnowflakeWarehouse) {
	*o = *x
	o.ObjectMeta = *x.ObjectMeta.DeepCopy()
}
func (x *SnowflakeWarehouse) DeepCopy() *SnowflakeWarehouse {
	if x == nil {
		return nil
	}
	o := new(SnowflakeWarehouse)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeWarehouse) DeepCopyObject() runtime.Object { return x.DeepCopy() }

func (x *SnowflakeWarehouseList) DeepCopyInto(o *SnowflakeWarehouseList) {
	*o = *x
	o.Items = append([]SnowflakeWarehouse(nil), x.Items...)
	for i := range o.Items {
		o.Items[i].DeepCopyInto(&o.Items[i])
	}
}
func (x *SnowflakeWarehouseList) DeepCopy() *SnowflakeWarehouseList {
	if x == nil {
		return nil
	}
	o := new(SnowflakeWarehouseList)
	x.DeepCopyInto(o)
	return o
}
func (x *SnowflakeWarehouseList) DeepCopyObject() runtime.Object { return x.DeepCopy() }
