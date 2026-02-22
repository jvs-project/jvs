package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewSnapshotResource creates a new snapshot resource
func NewSnapshotResource() resource.Resource {
	return &snapshotResource{}
}

type snapshotResource struct{}

func (r *snapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "jvs_snapshot"
}

func (r *snapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A JVS snapshot capturing workspace state at a point in time.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique snapshot ID.",
			},
			"repository": schema.StringAttribute{
				Required:    true,
				Description: "Path to the JVS repository.",
			},
			"worktree": schema.StringAttribute{
				Required:    true,
				Description: "Name of the worktree to snapshot.",
			},
			"note": schema.StringAttribute{
				Optional:    true,
				Description: "Descriptive note for the snapshot.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags to attach to the snapshot.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when snapshot was created.",
			},
			"snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "The generated snapshot ID.",
			},
		},
	}
}

func (r *snapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate snapshot ID (simplified - would use actual JVS library)
	snapshotID := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), "abcd1234")
	repoPath := plan.Repository.ValueString()

	// Create snapshot directory
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", snapshotID)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create snapshot directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create descriptor
	descriptorDir := filepath.Join(repoPath, ".jvs", "descriptors")
	if err := os.MkdirAll(descriptorDir, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create descriptors directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	descriptorContent := fmt.Sprintf(`{
  "snapshot_id": "%s",
  "worktree_name": "%s",
  "created_at": "%s",
  "note": "%s",
  "tags": %s,
  "engine": "copy"
}`, snapshotID, plan.Worktree.ValueString(), time.Now().UTC().Format(time.RFC3339),
		plan.Note.ValueString(), plan.Tags.String())

	descriptorFile := filepath.Join(descriptorDir, snapshotID+".json")
	if err := os.WriteFile(descriptorFile, []byte(descriptorContent), 0644); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create descriptor",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Update state
	plan.ID = types.StringValue(snapshotID)
	plan.SnapshotID = types.StringValue(snapshotID)
	plan.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *snapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify snapshot still exists
	repoPath := state.Repository.ValueString()
	descriptorFile := filepath.Join(repoPath, ".jvs", "descriptors", state.ID.ValueString()+".json")

	if _, err := os.Stat(descriptorFile); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *snapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Snapshots are immutable - no updates allowed
	resp.Diagnostics.AddError(
		"Cannot update snapshot",
		"Snapshots are immutable. Create a new snapshot instead.",
	)
}

func (r *snapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete snapshot files
	repoPath := state.Repository.ValueString()
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", state.ID.ValueString())
	descriptorFile := filepath.Join(repoPath, ".jvs", "descriptors", state.ID.ValueString()+".json")

	// Remove snapshot directory
	if err := os.RemoveAll(snapshotDir); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddWarning(
			"Failed to remove snapshot directory",
			fmt.Sprintf("Error: %s", err),
		)
	}

	// Remove descriptor
	if err := os.Remove(descriptorFile); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddWarning(
			"Failed to remove descriptor",
			fmt.Sprintf("Error: %s", err),
		)
	}
}

func (r *snapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	snapshotID := req.ID

	var state snapshotResourceModel
	state.ID = types.StringValue(snapshotID)
	state.SnapshotID = types.StringValue(snapshotID)
	state.Repository = types.StringValue(".")
	state.Worktree = types.StringValue("main")
	state.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// snapshotResourceModel models the snapshot resource data
type snapshotResourceModel struct {
	Repository   types.String `tfsdk:"repository"`
	Worktree     types.String `tfsdk:"worktree"`
	Note         types.String `tfsdk:"note"`
	Tags         types.List   `tfsdk:"tags"`
	ID           types.String `tfsdk:"id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	SnapshotID   types.String `tfsdk:"snapshot_id"`
}
