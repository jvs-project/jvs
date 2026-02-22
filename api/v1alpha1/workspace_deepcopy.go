package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopy implements the DeepCopy method for Workspace
func (in *Workspace) DeepCopy() *Workspace {
	if in == nil {
		return nil
	}
	out := new(Workspace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for Workspace
func (in *Workspace) DeepCopyInto(out *Workspace) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopyObject implements the DeepCopyObject method for Workspace
func (in *Workspace) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy implements the DeepCopy method for WorkspaceList
func (in *WorkspaceList) DeepCopy() *WorkspaceList {
	if in == nil {
		return nil
	}
	out := new(WorkspaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for WorkspaceList
func (in *WorkspaceList) DeepCopyInto(out *WorkspaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Workspace, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopyObject implements the DeepCopyObject method for WorkspaceList
func (in *WorkspaceList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopy implements the DeepCopy method for WorkspaceSpec
func (in *WorkspaceSpec) DeepCopy() *WorkspaceSpec {
	if in == nil {
		return nil
	}
	out := new(WorkspaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for WorkspaceSpec
func (in *WorkspaceSpec) DeepCopyInto(out *WorkspaceSpec) {
	*out = *in
	if in.JuiceFSConfig != nil {
		in, out := &in.JuiceFSConfig, &out.JuiceFSConfig
		*out = new(JuiceFSConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.RetentionPolicy != nil {
		in, out := &in.RetentionPolicy, &out.RetentionPolicy
		*out = new(RetentionPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.AutoSnapshot != nil {
		in, out := &in.AutoSnapshot, &out.AutoSnapshot
		*out = new(AutoSnapshotConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.MountOptions != nil {
		in, out := &in.MountOptions, &out.MountOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy implements the DeepCopy method for WorkspaceStatus
func (in *WorkspaceStatus) DeepCopy() *WorkspaceStatus {
	if in == nil {
		return nil
	}
	out := new(WorkspaceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for WorkspaceStatus
func (in *WorkspaceStatus) DeepCopyInto(out *WorkspaceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastSnapshotTime != nil {
		in, out := &in.LastSnapshotTime, &out.LastSnapshotTime
		*out = (*in).DeepCopy()
	}
	if in.NextSnapshotTime != nil {
		in, out := &in.NextSnapshotTime, &out.NextSnapshotTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy implements the DeepCopy method for JuiceFSConfig
func (in *JuiceFSConfig) DeepCopy() *JuiceFSConfig {
	if in == nil {
		return nil
	}
	out := new(JuiceFSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for JuiceFSConfig
func (in *JuiceFSConfig) DeepCopyInto(out *JuiceFSConfig) {
	*out = *in
	if in.SecretsRef != nil {
		in, out := &in.SecretsRef, &out.SecretsRef
		*out = new(JuiceFSSecretsRef)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy implements the DeepCopy method for JuiceFSSecretsRef
func (in *JuiceFSSecretsRef) DeepCopy() *JuiceFSSecretsRef {
	if in == nil {
		return nil
	}
	out := new(JuiceFSSecretsRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for JuiceFSSecretsRef
func (in *JuiceFSSecretsRef) DeepCopyInto(out *JuiceFSSecretsRef) {
	*out = *in
	if in.Keys != nil {
		in, out := &in.Keys, &out.Keys
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy implements the DeepCopy method for RetentionPolicy
func (in *RetentionPolicy) DeepCopy() *RetentionPolicy {
	if in == nil {
		return nil
	}
	out := new(RetentionPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for RetentionPolicy
func (in *RetentionPolicy) DeepCopyInto(out *RetentionPolicy) {
	*out = *in
	if in.KeepTags != nil {
		in, out := &in.KeepTags, &out.KeepTags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy implements the DeepCopy method for AutoSnapshotConfig
func (in *AutoSnapshotConfig) DeepCopy() *AutoSnapshotConfig {
	if in == nil {
		return nil
	}
	out := new(AutoSnapshotConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto implements the DeepCopyInto method for AutoSnapshotConfig
func (in *AutoSnapshotConfig) DeepCopyInto(out *AutoSnapshotConfig) {
	*out = *in
	return
}
