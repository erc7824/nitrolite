# Release Process & Branching Model

This document describes the **branching model** and **release process** for this repository.  
We follow the **[Git Flow](https://git-flow.sh/workflows/gitflow/)** workflow to maintain stable releases while enabling continuous development.

_This document ensures consistent and reliable release management across all contributors._

## Branching Model Overview

Our repository uses the following primary branches:

| Branch | Purpose |
|--------|----------|
| `main` | Always reflects the **production-ready** state. Every tagged commit on `main` represents a released version. |
| `develop` | Contains the **latest completed features** and integration code for the upcoming release. |
| `feature/*` | Used for **individual feature development**. Branched from `develop`. |
| `release/*` | Used to **prepare a new release** (version bump, changelog, final testing). Branched from `develop`. |
| `hotfix/*` | Used for **urgent fixes** to the production code. Branched from `main`. |


## Workflow Summary

### Feature Development

1. Create a new branch from `develop`:
   ```bash
   git checkout develop
   git pull
   git checkout -b feature/<feature-name>
   ```
2. Implement the feature and commit changes.
3. When done, open a **Pull Request** (PR) into `develop`.
4. After review and testing, merge and delete the feature branch.


### Starting a Release

When the `develop` branch is stable and ready for release:

```bash
git checkout develop
git pull
git checkout -b release/<version>
```

Perform:
- Version bump in `package.json` / `pyproject.toml` / etc.
- Update `CHANGELOG.md`
- Final testing and QA.

When ready:
```bash
git checkout main
git merge --no-ff release/<version>
git tag -a v<version> -m "Release <version>"
git checkout develop
git merge --no-ff release/<version>
git branch -d release/<version>
git push origin main develop --tags
```

### Hotfix Process

For critical fixes in production:

```bash
git checkout main
git pull
git checkout -b hotfix/<version>
```

Fix the issue, bump the patch version, and then:
```bash
git commit -am "Hotfix <version>: <description>"
git checkout main
git merge --no-ff hotfix/<version>
git tag -a v<version> -m "Hotfix <version>"
git checkout develop
git merge --no-ff hotfix/<version>
git branch -d hotfix/<version>
git push origin main develop --tags
```

## Creating a GitHub Release

After pushing the new tag and commits to the remote repository:

1. Go to the **GitHub Releases** page.
2. Click **“Draft a new release”**.
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
- [ ] All features merged into `develop`
- [ ] Tests passing
- [ ] Documentation updated
- [ ] Version bump committed
- [ ] CHANGELOG updated
- [ ] Release PR reviewed and approved

## References
- Git Flow: [https://git-flow.sh/workflows/gitflow/](https://git-flow.sh/workflows/gitflow/)
- Semantic Versioning: [https://semver.org/](https://semver.org/)
