package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewSnapshotDataSource creates a new snapshot data source
func NewSnapshotDataSource() datasource.DataSource {
	return &snapshotDataSource{}
}

type snapshotDataSource struct{}

func (d *snapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "jvs_snapshot"
}

func (d *snapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a JVS snapshot's information.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The snapshot ID.",
			},
			"repository": schema.StringAttribute{
				Required:    true,
				Description: "Path to the JVS repository.",
			},
			"worktree": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the worktree.",
			},
			"note": schema.StringAttribute{
				Computed:    true,
				Description: "Snapshot note.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Snapshot tags.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when snapshot was created.",
			},
		},
	}
}

func (d *snapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read snapshot information from repository
	// This would use the actual JVS library to read the descriptor

	// For now, set minimal values
	config.Worktree = types.StringValue("main")
	config.Note = types.StringValue("")
	config.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	config.CreatedAt = types.StringValue("")

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// snapshotDataSourceModel models the snapshot data source
type snapshotDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Repository types.String `tfsdk:"repository"`
	Worktree  types.String `tfsdk:"worktree"`
	Note      types.String `tfsdk:"note"`
	Tags      types.List   `tfsdk:"tags"`
	CreatedAt types.String `tfsdk:"created_at"`
}
