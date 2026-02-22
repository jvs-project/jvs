package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// NewWorktreeResource creates a new worktree resource
func NewWorktreeResource() resource.Resource {
	return &worktreeResource{}
}

type worktreeResource struct{}

func (r *worktreeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "jvs_worktree"
}

func (r *worktreeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A JVS worktree for isolated workspace development.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the worktree (worktree_name).",
			},
			"repository": schema.StringAttribute{
				Required:    true,
				Description: "Path to the JVS repository.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the worktree.",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path to the worktree payload directory.",
			},
			"engine": schema.StringAttribute{
				Optional:    true,
				Description: "Snapshot engine for this worktree (copy, reflink-copy, juicefs-clone).",
			},
			"head_snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "The current head snapshot ID.",
			},
			"latest_snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "The latest snapshot ID.",
			},
		},
	}
}

func (r *worktreeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan worktreeResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine worktree path
	worktreePath := plan.Path.ValueString()
	if worktreePath == "" {
		// Default to worktrees/<name>
		repoPath := plan.Repository.ValueString()
		worktreePath = filepath.Join(repoPath, "worktrees", plan.Name.ValueString())
	}

	// Create worktree directory
	if err := os.MkdirAll(worktreePath, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create worktree directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create worktree metadata file in .jvs/worktrees/
	jvsPath := filepath.Join(plan.Repository.ValueString(), ".jvs")
	worktreesMetaPath := filepath.Join(jvsPath, "worktrees")
	metaFile := filepath.Join(worktreesMetaPath, plan.Name.ValueString()+".json")

	// Ensure metadata directory exists
	if err := os.MkdirAll(worktreesMetaPath, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create worktree metadata directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Write worktree config
	engine := "copy"
	if !plan.Engine.IsNull() && !plan.Engine.IsUnknown() {
		engine = plan.Engine.ValueString()
	}

	configContent := fmt.Sprintf(`{
  "name": "%s",
  "path": "%s",
  "engine": "%s",
  "head_snapshot_id": "",
  "latest_snapshot_id": ""
}`, plan.Name.ValueString(), worktreePath, engine)

	if err := os.WriteFile(metaFile, []byte(configContent), 0644); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create worktree metadata",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Update state
	plan.Path = types.StringValue(worktreePath)
	plan.ID = types.StringValue(plan.Name.ValueString())
	plan.HeadSnapshotID = types.StringValue("")
	plan.LatestSnapshotID = types.StringValue("")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *worktreeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state worktreeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify worktree still exists
	worktreePath := state.Path.ValueString()
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Read worktree metadata to get current state
	jvsPath := filepath.Join(state.Repository.ValueString(), ".jvs")
	metaFile := filepath.Join(jvsPath, "worktrees", state.Name.ValueString()+".json")

	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *worktreeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan worktreeResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update worktree config if needed
	// For now, path changes require recreation

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *worktreeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state worktreeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete worktree directory
	worktreePath := state.Path.ValueString()
	if err := os.RemoveAll(worktreePath); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete worktree",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Delete worktree metadata
	jvsPath := filepath.Join(state.Repository.ValueString(), ".jvs")
	metaFile := filepath.Join(jvsPath, "worktrees", state.Name.ValueString()+".json")
	if err := os.Remove(metaFile); err != nil && !os.IsNotExist(err) {
		resp.Diagnostics.AddWarning(
			"Failed to remove worktree metadata",
			fmt.Sprintf("Error: %s", err),
		)
	}
}

func (r *worktreeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import from repository_path:worktree_name
	// For now, just use the worktree name
	name := req.ID

	// Try to find the repository by checking common paths
	// This would be improved with actual JVS library integration

	var state worktreeResourceModel
	state.Name = types.StringValue(name)
	state.ID = types.StringValue(name)
	state.Repository = types.StringValue(".")
	state.Path = types.StringValue(filepath.Join("worktrees", name))
	state.HeadSnapshotID = types.StringValue("")
	state.LatestSnapshotID = types.StringValue("")
	state.Engine = types.StringValue("copy")

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// worktreeResourceModel models the worktree resource data
type worktreeResourceModel struct {
	Repository          types.String `tfsdk:"repository"`
	Name                types.String `tfsdk:"name"`
	Path                types.String `tfsdk:"path"`
	Engine              types.String `tfsdk:"engine"`
	ID                  types.String `tfsdk:"id"`
	HeadSnapshotID      types.String `tfsdk:"head_snapshot_id"`
	LatestSnapshotID    types.String `tfsdk:"latest_snapshot_id"`
}
