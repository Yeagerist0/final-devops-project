# Submission video

[`session21-demo.mp4`](session21-demo.mp4) is a silent, captioned, 80-second project walkthrough. It covers the architecture, application and REST API, the successful GitHub Actions run, Terraform/Kubernetes/Helm, monitoring/GitOps, and six troubleshooting challenges. The source slides are in `slides.html` and can be regenerated with `render.sh` (Chromium and FFmpeg required).

For a spoken presentation, use this short script while showing the repository and running app:

1. **0:00–0:10 — Overview:** “This is my Session 21 end-to-end DevOps project: a React and Go task board with PostgreSQL.”
2. **0:10–0:20 — Architecture:** “GitHub Actions builds and tests the code, runs security gates, publishes to GHCR and updates Helm. Argo CD then syncs Kubernetes.”
3. **0:20–0:30 — Application:** “The UI supports task creation, completion and deletion. The Go API has REST, health, readiness and metrics endpoints.”
4. **0:30–0:40 — CI/CD:** “All eight jobs passed in [this run](https://github.com/Yeagerist0/final-devops-project/actions/runs/38933948746), including image scanning and GitOps delivery.”
5. **0:40–0:50 — Infrastructure:** “Terraform defines the AWS VPC and EKS cluster. Kubernetes and Helm provide Deployment, Service, Secret, ConfigMap, Ingress, HPA, probes and persistent storage.”
6. **0:50–1:00 — Monitoring and GitOps:** “Prometheus collects app metrics; Grafana visualizes them. Argo CD watches the SHA-pinned Helm chart.”
7. **1:00–1:10 — Troubleshooting:** “The lab contains six intentional faults plus an investigation template and answer key.”
8. **1:10–1:20 — Close:** “The complete source, setup steps, screenshots and video are in the public repository. Cloud resources are not provisioned automatically because they cost money.”

The video shows verified local and CI evidence. It does not claim an AWS EKS cluster or live Argo CD sync was launched.
