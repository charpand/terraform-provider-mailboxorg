# TODO: security hardening, manual steps

Temporary checklist for the repository settings that go with the
security hardening in the pull requests below this one. None of this can
be done from a commit. Delete this file once every box is ticked.

## 1. Access

- [ ] Give the `@charpand/maintainers` team (dsdeboer, dsdeboerbot) at
      least **Write** access to this repository (Settings > Collaborators
      and teams). Without it GitHub ignores the team in `CODEOWNERS`
      and no code owner review is requested.
- [ ] Open `.github/CODEOWNERS` on GitHub and check that it shows no
      errors.

## 2. `release` environment and signing secrets

The release job runs in the `release` environment
(`.github/workflows/ci.yml`). Until the environment is configured the
job falls back to the repository secrets, so releases keep working.

- [ ] Create the `release` environment (Settings > Environments), if a
      tag build has not created it already.
- [ ] Deployment branches and tags: **Selected branches and tags**, with
      only the tag pattern `v*`.
- [ ] Required reviewers: `@charpand/maintainers` (optionally with
      **Prevent self-review**).
- [ ] Add `GPG_PRIVATE_KEY` and `PASSPHRASE` as **environment** secrets.
- [ ] Delete `GPG_PRIVATE_KEY` and `PASSPHRASE` from the **repository**
      secrets (Settings > Secrets and variables > Actions).
- [ ] Check that no other repository secret is left that only the
      release needs.

## 3. Rulesets

### Branch ruleset for `main`

- [ ] Restrict deletions and block force pushes.
- [ ] Require a pull request before merging, with 1 approval,
      **Require review from Code Owners**, dismiss stale approvals on
      new pushes, and require conversation resolution.
- [ ] Require status checks to pass (branches up to date), using the
      job names:
  - `ci.yml`: `lint`, `test`, `docs`
  - `security.yml`: `sbom`, `trivy`, `grype`, `govulncheck`,
    `analyze (go)`, `analyze (actions)`, `gitleaks`, `zizmor`,
    `dependency-review`
  - Snyk app: `security/snyk`, `code/snyk`
- [ ] Optionally require code scanning results (CodeQL, high or
      higher) and a linear history.
- [ ] Bypass list: none, or only the maintainers team for pull
      requests.

### Tag ruleset for `v*`

- [ ] Restrict creations, updates and deletions of `v*` tags, with only
      the maintainers team allowed to bypass, so only they can cut a
      release.

## 4. Code security settings (Settings > Code security)

- [ ] Dependency graph: enabled (dependency review needs it).
- [ ] Dependabot alerts: enabled (Renovate keeps opening the update
      PRs, so Dependabot security updates can stay off).
- [ ] Code scanning: keep CodeQL **default setup off**. The `analyze`
      jobs are an advanced setup, and their upload fails while default
      setup is on.
- [ ] Secret scanning and push protection: enabled.
- [ ] Private vulnerability reporting: enabled (`SECURITY.md` links to
      it).

## 5. Actions settings (Settings > Actions > General)

- [ ] Workflow permissions: **Read repository contents** by default,
      and do not allow GitHub Actions to create or approve pull
      requests.
- [ ] Fork pull request workflows: require approval for all outside
      collaborators.
- [ ] Optionally: allow only selected actions and require actions to be
      pinned to a full-length commit SHA (every action in the workflows
      already is).

## 6. Follow-up checks

- [ ] After the stack is merged: the `Security` and `Scorecard` runs on
      `main` pass, the CodeQL and Scorecard results appear under
      Security > Code scanning, and the README badge shows a score.
- [ ] Work through the Scorecard findings that remain.
- [ ] After a few weeks of runs: review the egress that harden-runner
      recorded and switch `egress-policy: audit` to `block` with an
      allowlist, job by job.
- [ ] After the next release: the release job waited for the
      environment approval, and
      `gh attestation verify <archive> --repo charpand/terraform-provider-mailboxorg`
      succeeds for a downloaded archive.
- [ ] Delete this file.
