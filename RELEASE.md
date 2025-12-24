# Release Guide (Gitflow)

Short: Create a release branch `release/x.y.z` from `develop` and push it — CI will create a tag and a GitHub Release. Afterwards merge the branch back into `master` and `develop`.

1) Create the release branch

```bash
git checkout develop
git pull origin develop
git checkout -b release/1.2.0
# Update version/changelog, then commit
git commit -am "Bump version to 1.2.0"
git push origin release/1.2.0
```

2) What happens after the push
- Pushing `release/1.2.0` triggers the GitHub Action `.github/workflows/release.yml`.
- The Action runs tests, builds the binary, creates a tag `v1.2.0` and publishes a GitHub Release.
- Alternatively you can trigger the workflow manually with `workflow_dispatch` and provide a `version` input.

3) After a successful release: merge

```bash
git checkout master
git pull origin master
git merge --no-ff release/1.2.0
git push origin master

# If the tag does not exist locally (Action already pushes the tag):
git tag -a v1.2.0 -m "Release v1.2.0" || true
git push origin --tags

git checkout develop
git pull origin develop
git merge --no-ff release/1.2.0
git push origin develop

git branch -d release/1.2.0
git push origin --delete release/1.2.0
```

4) Notes
- Ensure under `Settings → Actions → General → Workflow permissions` that Actions has `Read and write` permission; otherwise the workflow cannot push tags or create releases.
- Automatic merging or PR automation after a release is possible, but can be fragile in case of merge conflicts — consider handling merges manually or via PRs.
- Use `release/` branches without a leading `v`; the workflow will add the leading `v` to the tag.

If you want, I can add automatic merge/PR steps after the release.
