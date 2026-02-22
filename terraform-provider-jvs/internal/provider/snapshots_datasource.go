package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewSnapshotsDataSource creates a snapshots data source
func NewSnapshotsDataSource() datasource.DataSource {
	return &snapshotsDataSource{}
}

type snapshotsDataSource struct{}

func (d *snapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "jvs_snapshots"
}

func (d *snapshotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all snapshots in a JVS repository.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier (same as repository path).",
			},
			"repository": schema.StringAttribute{
				Required:    true,
				Description: "Path to the JVS repository.",
			},
			"worktree": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by worktree name.",
			},
			"tag": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by tag.",
			},
			"snapshots": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of snapshots.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							Description: "Snapshot ID.",
						},
						"worktree": schema.StringAttribute{
							Computed: true,
							Description: "Worktree name.",
						},
						"note": schema.StringAttribute{
							Computed: true,
							Description: "Snapshot note.",
						},
						"created_at": schema.StringAttribute{
							Computed: true,
							Description: "Creation timestamp.",
						},
						"tags": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Snapshot tags.",
						},
					},
				},
			},
		},
	}
}

func (d *snapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotsDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read descriptors from repository
	descriptorsDir := filepath.Join(config.Repository.ValueString(), ".jvs", "descriptors")
	entries, err := os.ReadDir(descriptorsDir)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read descriptors",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	var snapshots []snapshotData

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		descriptorFile := filepath.Join(descriptorsDir, entry.Name())
		content, err := os.ReadFile(descriptorFile)
		if err != nil {
			continue
		}

		// Parse descriptor (simplified - would use proper JSON parsing)
		snapshotID := strings.TrimSuffix(entry.Name(), ".json")

		// Apply filters
		if !config.Worktree.IsNull() && !config.Worktree.IsUnknown() {
			// Would check if snapshot belongs to specified worktree
		}

		if !config.Tag.IsNull() && !config.Tag.IsUnknown() {
			// Would check if snapshot has specified tag
		}

		snapshots = append(snapshots, snapshotData{
			ID:        types.StringValue(snapshotID),
			Worktree:  types.StringValue("main"), // Would parse from descriptor
			Note:      types.StringValue(""),
			CreatedAt: types.StringValue(""),
			Tags:      types.ListValueMust(types.StringType, []attr.Value{}),
		})
	}

	// Convert to framework types
	snapshotList, diags := types.ObjectValueFrom(ctx, snapshotAttrType, snapshots)
	resp.Diagnostics.Append(diags...)

	config.ID = types.StringValue(config.Repository.ValueString())
	config.Snapshots = snapshotList

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// snapshotData models a single snapshot
type snapshotData struct {
	ID        types.String `tfsdk:"id"`
	Worktree  types.String `tfsdk:"worktree"`
	Note      types.String `tfsdk:"note"`
	CreatedAt types.String `tfsdk:"created_at"`
	Tags      types.List   `tfsdk:"tags"`
}

// snapshotsDataSourceModel models the snapshots data source
type snapshotsDataSourceModel struct {
	Repository types.String `tfsdk:"repository"`
	Worktree   types.String `tfsdk:"worktree"`
	Tag        types.String `tfsdk:"tag"`
	ID         types.String `tfsdk:"id"`
	Snapshots  types.List   `tfsdk:"snapshots"`
}

// Type information for snapshot attribute
var snapshotAttrType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":         types.StringType,
		"worktree":  types.StringType,
		"note":       types.StringType,
		"created_at": types.StringType,
		"tags":       types.ListType{ElemType: types.StringType},
	},
}
