# JVS Ansible Role

An Ansible role for installing and configuring [JVS (Juicy Versioned Workspaces)](https://github.com/jvs-project/jvs).

## Requirements

- Ansible 2.9+
- Go 1.21+ (when building from source)

## Role Variables

### Installation

| Variable | Default | Description |
|----------|---------|-------------|
| `jvs_version` | `v7.0` | JVS version to install |
| `jvs_install_method` | `release` | Installation method: `release` or `source` |
| `jvs_bin_dir` | `/usr/local/bin` | Directory to install JVS binary |
| `jvs_user` | `root` | JVS binary owner |
| `jvs_group` | `root` | JVS binary group |

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `jvs_default_engine` | `auto` | Default snapshot engine |
| `jvs_default_tags` | `[]` | Default tags for snapshots |
| `jvs_output_format` | `text` | Output format: `text` or `json` |
| `jvs_progress_enabled` | `true` | Enable progress bars |

### Repositories

| Variable | Default | Description |
|----------|---------|-------------|
| `jvs_repositories` | `[]` | List of repositories to create |
| `jvs_repositories_base_path` | `/opt` | Base path for repositories |

### Garbage Collection

| Variable | Default | Description |
|----------|---------|-------------|
| `jvs_gc_enabled` | `false` | Enable garbage collection |
| `jvs_gc_schedule` | `""` | GC schedule (cron format) |
| `jvs_gc_systemd_timer` | `false` | Create systemd timers for GC |

### Webhooks

| Variable | Default | Description |
|----------|---------|-------------|
| `jvs_webhooks_enabled` | `false` | Enable webhook notifications |
| `jvs_webhooks` | `[]` | List of webhook endpoints |

## Dependencies

None.

## Example Playbook

### Basic Installation

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
```

### With Configuration

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
      vars:
        jvs_version: "v7.2"
        jvs_default_engine: "juicefs-clone"
        jvs_default_tags:
          - "production"
          - "managed"
```

### Creating Repositories

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
      vars:
        jvs_repositories:
          - name: workspace
            path: /data/workspace
            engine: juicefs-clone
            default_tags:
              - production
          - name: experiments
            path: /data/experiments
            default_tags:
              - experimental
```

### With Webhooks

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
      vars:
        jvs_webhooks_enabled: true
        jvs_webhooks:
          - url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
            events:
              - snapshot.created
              - restore.complete
          - url: https://example.com/webhook
            secret: your-hmac-secret
            events:
              - "*"
```

### With Garbage Collection

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
      vars:
        jvs_gc_enabled: true
        jvs_gc_schedule: "0 2 * * *"  # Daily at 2 AM
        jvs_gc_systemd_timer: true
        jvs_repositories:
          - name: workspace
            path: /data/workspace
```

### Building from Source

```yaml
---
- hosts: servers
  become: true
  roles:
    - role: jvs-project.jvs
      vars:
        jvs_install_method: "source"
        jvs_version: "main"  # or specific branch/tag
```

## Tasks

This role includes the following task tags:

- `install` - Install JVS binary
- `config` - Configure JVS
- `repositories` - Create repositories
- `gc` - Set up garbage collection
- `completion` - Install shell completions
- `verify` - Verify installation

Run specific tasks with tags:

```bash
ansible-playbook playbook.yml --tags "install,config"
```

Skip specific tasks:

```bash
ansible-playbook playbook.yml --skip-tags "gc"
```

## Handlers

- `reload systemd` - Reload systemd daemon
- `jvs installed` - Notification that JVS was installed

## License

MIT

## Author Information

JVS Project - https://github.com/jvs-project/jvs
