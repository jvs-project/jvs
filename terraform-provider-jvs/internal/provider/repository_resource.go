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
)

// NewRepositoryResource creates a new repository resource
func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct{}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "jvs_repository"
}

func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A JVS repository for versioned workspace management.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the repository (repo_id).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the repository.",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Path where the repository will be created. Defaults to provider repo_path + name.",
			},
			"engine": schema.StringAttribute{
				Optional:    true,
				Description: "Default snapshot engine (copy, reflink-copy, juicefs-clone).",
			},
			"format_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Format version of the repository.",
			},
		},
	}
}

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get provider data
	providerData := resp.ResourceData.(*providerData)

	// Determine full path
	fullPath := plan.Path.ValueString()
	if fullPath == "" {
		fullPath = providerData.GetRepoPath(plan.Name.ValueString())
	}

	// Create repository directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create repository directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Initialize JVS repository
	// Note: This would call the actual JVS library
	// For now, we'll simulate by creating the structure
	jvsPath := filepath.Join(fullPath, ".jvs")
	if err := os.MkdirAll(jvsPath, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create .jvs directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create repo_id file
	repoID := "repo-" + plan.Name.ValueString()
	if err := os.WriteFile(filepath.Join(jvsPath, "repo_id"), []byte(repoID), 0644); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create repo_id",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create format_version file
	if err := os.WriteFile(filepath.Join(jvsPath, "format_version"), []byte("1"), 0644); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create format_version",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create worktrees directory
	if err := os.MkdirAll(filepath.Join(jvsPath, "worktrees"), 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create worktrees directory",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Create main worktree
	mainPath := filepath.Join(fullPath, "main")
	if err := os.MkdirAll(mainPath, 0755); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create main worktree",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Update state with computed values
	plan.Path = types.StringValue(fullPath)
	plan.ID = types.StringValue(repoID)
	plan.FormatVersion = types.Int64Value(1)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify repository still exists
	fullPath := state.Path.ValueString()
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update engine setting if changed
	// Note: This would update the actual JVS config

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete repository directory
	fullPath := state.Path.ValueString()
	if err := os.RemoveAll(fullPath); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete repository",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import from path
	path := req.ID

	// Read existing repository to populate state
	jvsPath := filepath.Join(path, ".jvs")
	repoIDBytes, err := os.ReadFile(filepath.Join(jvsPath, "repo_id"))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read repository",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	var state repositoryResourceModel
	state.Path = types.StringValue(path)
	state.ID = types.StringValue(string(repoIDBytes))
	state.Name = types.StringValue(filepath.Base(path))
	state.FormatVersion = types.Int64Value(1)
	state.Engine = types.StringValue("copy") // Default, would read from config

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// repositoryResourceModel models the repository resource data
type repositoryResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Path          types.String `tfsdk:"path"`
	Engine        types.String `tfsdk:"engine"`
	ID            types.String `tfsdk:"id"`
	FormatVersion types.Int64  `tfsdk:"format_version"`
}
