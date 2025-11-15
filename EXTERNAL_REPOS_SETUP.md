# External Repositories Setup

This directory contains compressed archives for the external template and plugin repositories.

## Archives Included

- **streamspace-templates.tar.gz** (50KB) - Template repository with 22 application templates
- **streamspace-plugins.tar.gz** (20KB) - Plugin repository with catalog structure

## Setup Instructions

### 1. Extract and Push streamspace-templates

```bash
# Extract the archive
tar -xzf streamspace-templates.tar.gz
cd streamspace-templates

# Verify contents
ls -la

# Set up git remote (if not already set)
git remote add origin git@github.com:JoshuaAFerguson/streamspace-templates.git

# Push to GitHub
git push -u origin claude/migrate-plugins-templates-repos-01Xybebni2LWWjvyHCZJzQsQ

# Create main branch (optional, after merging PR)
git checkout -b main
git push -u origin main
```

### 2. Extract and Push streamspace-plugins

```bash
# Extract the archive
tar -xzf streamspace-plugins.tar.gz
cd streamspace-plugins

# Verify contents
ls -la

# Add, commit, and push
git add -A
git commit -m "feat(plugins): Initial plugin repository setup

- Add plugin catalog structure
- Create README and CONTRIBUTING guides
- Set up GitHub Actions validation workflow
- Add directory structure for official and community plugins"

# Set up git remote (if not already set)
git remote add origin git@github.com:JoshuaAFerguson/streamspace-plugins.git

# Push to GitHub
git push -u origin claude/migrate-plugins-templates-repos-01Xybebni2LWWjvyHCZJzQsQ

# Create main branch (optional, after merging PR)
git checkout -b main
git push -u origin main
```

### 3. Cleanup (After Successful Push)

Once both repositories are successfully pushed to GitHub, you can remove these archives:

```bash
rm streamspace-templates.tar.gz streamspace-plugins.tar.gz
rm EXTERNAL_REPOS_SETUP.md
git add -A
git commit -m "chore: remove external repo archives after successful push"
git push
```

## Repository Contents

### streamspace-templates
- 22 application templates across 7 categories
- catalog.yaml for template discovery
- GitHub Actions validation workflow
- Comprehensive README and CONTRIBUTING guides

**Categories:**
- Browsers (4): Firefox, Chromium, Brave, LibreWolf
- Development (3): VS Code, GitHub Desktop, GitQlient
- Productivity (2): LibreOffice, Calligra
- Design (6): GIMP, Krita, Inkscape, Blender, FreeCAD, KiCad
- Media (2): Audacity, Kdenlive
- Gaming (2): DuckStation, Dolphin
- Webtop (3): Ubuntu XFCE, Ubuntu KDE, Alpine i3

### streamspace-plugins
- Plugin catalog structure
- official/ and community/ directories
- catalog.yaml with plugin metadata schema
- GitHub Actions validation workflow
- Security guidelines in CONTRIBUTING.md

## Integration

Once both repositories are pushed, StreamSpace will automatically sync them based on the configuration in `chart/values.yaml`:

```yaml
repositories:
  templates:
    enabled: true
    url: https://github.com/JoshuaAFerguson/streamspace-templates
    branch: main
    syncInterval: 1h

  plugins:
    enabled: true
    url: https://github.com/JoshuaAFerguson/streamspace-plugins
    branch: main
    syncInterval: 1h
```

## Verification

After pushing, verify the repositories are accessible:

```bash
# Check templates repo
curl https://raw.githubusercontent.com/JoshuaAFerguson/streamspace-templates/main/catalog.yaml

# Check plugins repo
curl https://raw.githubusercontent.com/JoshuaAFerguson/streamspace-plugins/main/catalog.yaml
```

## Support

- **Main Repository**: https://github.com/JoshuaAFerguson/streamspace
- **Templates Repository**: https://github.com/JoshuaAFerguson/streamspace-templates
- **Plugins Repository**: https://github.com/JoshuaAFerguson/streamspace-plugins
