# Release Process & Branching Model

This document describes the **branching model** and **release process** for this repository.  
We follow an extended version of **[Git Flow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)** workflow to suit our collaboration and review process in GitHub.

_This document ensures consistent and reliable release management across all contributors._

## Branching Model Overview

Our repository uses the following primary branches:

| Branch | Purpose |
|--------|----------|
| `stable` | Reflects the state in **production**. Every tagged commit on `stable` represents a released version. |
| `main` | The **main development branch** containing the latest completed features and integration code. |
| `feat/*` | Used for **feature development**. Can be based on either `main` or `release/*` branches. |
| `fix/*` | Used for **development fixes**. Can be based on either `main` or `release/*` branches. |
| `release/*` | Used to **prepare a new stable release**. Based on `main`. |
| `hotfix/*` | Used for **urgent fixes to production**. Based on `stable`. |

## Workflow Summary

### Important Note
- The only option for merging GitHub PRs should be **"Merge Pull Request"**

### Feature Development

1. Create a new branch from either `main` (if no release branch exists) or the appropriate `release/*` branch:
   ```bash
   git checkout main  # or release/vX.Y.Z
   git pull
   git checkout -b feat/<feature-name>
   ```
2. Implement the feature and commit changes.
3. Open a **Pull Request** (PR) on GitHub into the base branch.
4. During development, the branch can have all commits to facilitate the review process.
5. **Before merging**: All commits should be either squashed or interactively rebased to avoid littering the commit history with misleading commits.
6. After review and approval, merge using "Merge Pull Request" and delete the feature branch.

### Fix Development

The process is identical to feature development, but uses the `fix/` prefix for branch names:
```bash
git checkout main  # or release/vX.Y.Z
git pull
git checkout -b fix/<fix-name>
```

### Release Preparation

1. Create a release branch from `main`:
   ```bash
   git checkout main
   git pull
   git checkout -b release/<version>
   ```
   Where `<version>` follows semver format with a minor or major bump.

2. Open a **Pull Request** on GitHub from the release branch into `stable`.

3. The release branch can have a dedicated deployment (UAT) once configuration is provided:
   - Initially through manual preparation
   - Eventually automated via CI

4. Tag commits with release candidates:
   - Format: `<version>-rc.<num>` (starting from rc.0)
   - Create initial rc.0 tag on branch creation
   - Increment the rc number on each merge into the release branch

5. For each release candidate:
   - Publish packages (golang, npm, etc.)
   - Redeploy the release branch
   - Ideally automate these actions via CI

6. When the release is ready:
   - Merge the PR into `stable`
   - Tag HEAD of `stable` with the stable release version
   - Publish packages
   - Update deployment
   - Merge `stable` back into `main` (via PR or locally)

### Hotfix Process

Similar to release branches with key differences:

1. Create from `stable` branch:
   ```bash
   git checkout stable
   git pull
   git checkout -b hotfix/<version>
   ```
   Where `<version>` follows semver format with a patch bump.

2. Open a **Pull Request** on GitHub from the hotfix branch into `stable`.

3. Important differences from release branches:
   - No dedicated deployment
   - No pre-release versions (no rc tags)

4. When the hotfix is ready:
   - Merge the PR into `stable`
   - Tag HEAD of `stable` with the stable version
   - Publish packages
   - Update deployment
   - Merge `stable` back into `main` (via PR or locally)

## Creating a GitHub Release

After pushing the new tag and commits to the remote repository:

1. Go to the **GitHub Releases** page.
2. Click **"Draft a new release"**.
3. Select the new tag (`vX.Y.Z`), set the release title, and paste the changelog section.
4. Publish the release.

## Versioning

We follow **[Semantic Versioning](https://semver.org/)**:
```
MAJOR.MINOR.PATCH
```

- **MAJOR**: incompatible API changes  
- **MINOR**: backward-compatible features  
- **PATCH**: backward-compatible bug fixes  

Example: `v1.3.2`

## Example Release Checklist

Before finalizing a release:
- [ ] All features merged into release branch
- [ ] Tests passing
- [ ] Documentation updated
- [ ] Version bump committed
- [ ] CHANGELOG updated
- [ ] Release PR reviewed and approved
- [ ] All rc versions tested

## References
- Git Flow: [https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)
- Semantic Versioning: [https://semver.org/](https://semver.org/)
