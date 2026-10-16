# Security gates

`ci-cd.yml` blocks image publication unless all of these pass:

| Gate | Tool | Failure condition |
| --- | --- | --- |
| SAST | Gosec and GitHub CodeQL for Go | Gosec high-severity/high-confidence finding; CodeQL results appear in Security tab |
| SCA | Trivy filesystem scan | HIGH or CRITICAL dependency vulnerability |
| Secret scan | Gitleaks | A credential is detected in Git history |
| Image scan | Trivy image scan | HIGH or CRITICAL vulnerability in the final image |

CodeQL findings require a GitHub code scanning ruleset to block a pull request by alert severity; the analysis action alone uploads findings but does not fail on every alert. The sample Kubernetes Secret is classroom-only and must be replaced before exposure. Never commit a real token or `terraform.tfvars`. For production, integrate External Secrets or AWS Secrets Manager, use OIDC for cloud credentials, pin Actions by commit SHA, and require the security jobs in branch protection.

`.gitleaksignore` names one fingerprint from the first commit: the README's `local-demo-token` tutorial value. It is not a credential. The fingerprint is tied to that commit and line, so new findings in later commits are still checked.
