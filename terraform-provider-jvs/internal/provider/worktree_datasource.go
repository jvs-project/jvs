package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewWorktreeDataSource creates a new worktree data source
func NewWorktreeDataSource() datasource.DataSource {
	return &worktreeDataSource{}
}

type worktreeDataSource struct{}

func (d *worktreeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "jvs_worktree"
}

func (d *worktreeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a JVS worktree's information.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (worktree name).",
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
				Computed:    true,
				Description: "Path to the worktree payload directory.",
			},
			"engine": schema.StringAttribute{
				Computed:    true,
				Description: "Snapshot engine for this worktree.",
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

func (d *worktreeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config worktreeDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read worktree information from repository
	// This would use the actual JVS library to read the worktree config

	// For now, set default values
	config.ID = types.StringValue(config.Name.ValueString())
	config.Path = types.StringValue("worktrees/" + config.Name.ValueString())
	config.Engine = types.StringValue("copy")
	config.HeadSnapshotID = types.StringValue("")
	config.LatestSnapshotID = types.StringValue("")

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// worktreeDataSourceModel models the worktree data source
type worktreeDataSourceModel struct {
	Repository       types.String `tfsdk:"repository"`
	Name             types.String `tfsdk:"name"`
	Path             types.String `tfsdk:"path"`
	Engine           types.String `tfsdk:"engine"`
	ID               types.String `tfsdk:"id"`
	HeadSnapshotID   types.String `tfsdk:"head_snapshot_id"`
	LatestSnapshotID types.String `tfsdk:"latest_snapshot_id"`
}
