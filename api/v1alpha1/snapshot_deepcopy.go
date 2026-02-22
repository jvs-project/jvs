package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopy implements the DeepCopy method for Snapshot
func (in *Snapshot) DeepCopy() *Snapshot {
	if in == nil {
		return nil
	}
	out := new(Snapshot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for Snapshot
func (in *Snapshot) DeepCopyInto(out *Snapshot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopyObject implements the DeepCopyObject method for Snapshot
func (in *Snapshot) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy implements the DeepCopy method for SnapshotList
func (in *SnapshotList) DeepCopy() *SnapshotList {
	if in == nil {
		return nil
	}
	out := new(SnapshotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for SnapshotList
func (in *SnapshotList) DeepCopyInto(out *SnapshotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Snapshot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopyObject implements the DeepCopyObject method for SnapshotList
func (in *SnapshotList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy implements the DeepCopy method for SnapshotSpec
func (in *SnapshotSpec) DeepCopy() *SnapshotSpec {
	if in == nil {
		return nil
	}
	out := new(SnapshotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for SnapshotSpec
func (in *SnapshotSpec) DeepCopyInto(out *SnapshotSpec) {
	*out = *in
	if in.PartialPaths != nil {
		in, out := &in.PartialPaths, &out.PartialPaths
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Compression != nil {
		in, out := &in.Compression, &out.Compression
		*out = new(CompressionSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy implements the DeepCopy method for SnapshotStatus
func (in *SnapshotStatus) DeepCopy() *SnapshotStatus {
	if in == nil {
		return nil
	}
	out := new(SnapshotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for SnapshotStatus
func (in *SnapshotStatus) DeepCopyInto(out *SnapshotStatus) {
	*out = *in
	if in.CreatedAt != nil {
		in, out := &in.CreatedAt, &out.CreatedAt
		*out = (*in).DeepCopy()
	}
	if in.CompletedAt != nil {
		in, out := &in.CompletedAt, &out.CompletedAt
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RestoreStatus != nil {
		in, out := &in.RestoreStatus, &out.RestoreStatus
		*out = new(RestoreStatus)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy implements the DeepCopy method for CompressionSpec
func (in *CompressionSpec) DeepCopy() *CompressionSpec {
	if in == nil {
		return nil
	}
	out := new(CompressionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for CompressionSpec
func (in *CompressionSpec) DeepCopyInto(out *CompressionSpec) {
	*out = *in
	return
}

// DeepCopy implements the DeepCopy method for RestoreStatus
func (in *RestoreStatus) DeepCopy() *RestoreStatus {
	if in == nil {
		return nil
	}
	out := new(RestoreStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for RestoreStatus
func (in *RestoreStatus) DeepCopyInto(out *RestoreStatus) {
	*out = *in
	if in.StartedAt != nil {
		in, out := &in.StartedAt, &out.StartedAt
		*out = (*in).DeepCopy()
	}
	if in.CompletedAt != nil {
		in, out := &in.CompletedAt, &out.CompletedAt
		*out = (*in).DeepCopy()
	}
	return
}
