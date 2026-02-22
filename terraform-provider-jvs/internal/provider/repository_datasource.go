package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewRepositoryDataSource creates a new repository data source
func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

type repositoryDataSource struct{}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "jvs_repository"
}

func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a JVS repository's information.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the repository.",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the JVS repository.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the repository.",
			},
			"format_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Format version of the repository.",
			},
		},
	}
}

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read repository information
	jvsPath := filepath.Join(config.Path.ValueString(), ".jvs")
	repoIDFile := filepath.Join(jvsPath, "repo_id")
	formatFile := filepath.Join(jvsPath, "format_version")

	// Check if repository exists
	if _, err := os.Stat(jvsPath); os.IsNotExist(err) {
		resp.Diagnostics.AddError(
			"Repository not found",
			fmt.Sprintf("No JVS repository found at: %s", config.Path.ValueString()),
		)
		return
	}

	// Read repo_id
	repoIDBytes, err := os.ReadFile(repoIDFile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read repo_id",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Read format version
	formatBytes, err := os.ReadFile(formatFile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read format_version",
			fmt.Sprintf("Error: %s", err),
		)
		return
	}

	// Parse format version
	var formatVersion int64
	if _, err := fmt.Sscanf(string(formatBytes), "%d", &formatVersion); err != nil {
		formatVersion = 1
	}

	// Set state
	config.Name = types.StringValue(filepath.Base(config.Path.ValueString()))
	config.ID = types.StringValue(string(repoIDBytes))
	config.FormatVersion = types.Int64Value(formatVersion)

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// repositoryDataSourceModel models the repository data source
type repositoryDataSourceModel struct {
	Path          types.String `tfsdk:"path"`
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
	FormatVersion types.Int64  `tfsdk:"format_version"`
}
