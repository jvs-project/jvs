package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// New creates a new JVS provider
func New() provider.Provider {
	return &jvsProvider{}
}

type jvsProvider struct{}

func (p *jvsProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jvs"
	resp.Version = "1.0.0"
}

func (p *jvsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repo_path": schema.StringAttribute{
				Description: "Default path to JVS repositories. Can be overridden per resource.",
				Optional:    true,
			},
		},
	}
}

func (p *jvsProvider) Configure(_ context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Get provider configuration
	var config providerConfig

	diags := req.Config.Get(context.Background(), &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Store configured values in provider data
	providerData := &providerData{
		RepoPath: config.RepoPath.ValueString(),
	}

	if providerData.RepoPath == "" {
		providerData.RepoPath = "." // Default to current directory
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *jvsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRepositoryResource,
		NewWorktreeResource,
		NewSnapshotResource,
	}
}

func (p *jvsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRepositoryDataSource,
		NewWorktreeDataSource,
		NewSnapshotDataSource,
		NewSnapshotsDataSource,
	}
}

// providerConfig holds the provider configuration
type providerConfig struct {
	RepoPath string `tfsdk:"repo_path"`
}

// providerData holds data shared between resources and data sources
type providerData struct {
	RepoPath string
}

// GetRepoPath returns the full path to a repository
func (d *providerData) GetRepoPath(name string) string {
	if d.RepoPath == "." || d.RepoPath == "" {
		return fmt.Sprintf("./%s", name)
	}
	return fmt.Sprintf("%s/%s", d.RepoPath, name)
}
