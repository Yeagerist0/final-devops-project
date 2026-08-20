# Session 21: Final DevOps Project & Troubleshooting

A complete task board capstone carried from source control to Kubernetes. React provides the responsive dashboard, Go provides authenticated REST APIs, and PostgreSQL stores tasks on a persistent volume. Students can use raw manifests for resource inspection or Helm for GitOps delivery.

This original implementation follows the [Session 21 capstone brief](https://github.com/Nency-Ravaliya/devops-heros/tree/main/session21-python). Its backend is Go instead of FastAPI; the required frontend, PostgreSQL, REST, container, infrastructure, monitoring and troubleshooting concepts are all demonstrated.

## Architecture

```mermaid
flowchart LR
    A[React dashboard + Go API + PostgreSQL] --> B[Git / GitHub]
    B --> C[GitHub Actions]
    C --> D[Build and tests]
    D --> E[CodeQL + Trivy + Gitleaks]
    E --> F[Docker image]
    F --> G[GitHub Container Registry]
    C --> H[Helm image tag commit]
    H --> I[Argo CD]
    G --> J[EKS / Kubernetes]
    I --> J
    K[Terraform: VPC + EKS + nodes] --> J
    J --> L[Prometheus + Grafana + logs]
```

## Technologies and layout

React, Vite, Go, pgx, PostgreSQL, Git/GitHub, GitHub Actions, CodeQL, Go vet, Gitleaks, Trivy, Docker Compose, GHCR, Terraform, AWS VPC/EKS, Kubernetes, Helm, Prometheus, Grafana and Argo CD.

| Directory | Purpose |
| --- | --- |
| `application/` | Task board, API, metrics and tests |
| `frontend/` | Responsive React dashboard built with Vite |
| `docker/` | Multi-stage non-root image |
| `kubernetes/` | Raw resources and broken-resource labs |
| `helm/taskboard/` | Templated deployment |
| `terraform/` | AWS VPC and EKS infrastructure |
| `.github/workflows/` | Quality, security, image and GitOps pipeline |
| `security/` | Gate explanation |
| `monitoring/` | Prometheus and Grafana with provisioned data source |
| `gitops/` | Argo CD Application |

## Application setup

Fastest full-stack start requires Docker Compose:

```sh
docker compose up --build -d
docker compose ps
```

Open `http://localhost:8080`; demo write token is `local-demo-token`. Compose starts PostgreSQL, waits for its health check, builds the React/Go image and retains database data in a named volume. Set `API_TOKEN` and `DB_PASSWORD` in a local `.env` file before exposing the app. To stop: `docker compose down` (omit `-v` to keep data).

For development, use Go 1.27+ and Node 24+:

```sh
cd application
go test ./...
API_TOKEN=local-demo-token DATA_FILE=./data/tasks.json go run .
# In another terminal, from repo root:
cd frontend
npm ci
npm run dev
```

The Vite dev UI is at `http://localhost:5173` and proxies API calls to Go. Direct Go startup uses file storage for a quick no-database demo; Compose/Kubernetes use PostgreSQL. API examples:

```sh
curl http://localhost:8080/api/tasks
curl -X POST http://localhost:8080/api/tasks -H 'Content-Type: application/json' -H 'Authorization: Bearer local-demo-token' -d '{"title":"Verify CI"}'
curl -X PUT http://localhost:8080/api/tasks/1 -H 'Content-Type: application/json' -H "Authorization: Bearer ${API_TOKEN}" -d '{"title":"Verify CI","done":true}'
curl -X DELETE http://localhost:8080/api/tasks/1 -H "Authorization: Bearer ${API_TOKEN}"
curl http://localhost:8080/api/tasks/stats
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

`GET /api/tasks`, `GET /api/tasks/stats`, `POST /api/tasks`, `PUT /api/tasks/{id}` and `DELETE /api/tasks/{id}` are the main REST APIs. PostgreSQL schema is created on first startup. `API_TOKEN` is required for writes; GET APIs and UI are public in this teaching example. Logs go to stdout/stderr.

## Docker setup

From the repo root, build a single production image containing the compiled React UI and Go API:

```sh
docker build -f docker/Dockerfile -t taskboard:local .
docker compose up -d
```

The final image uses a non-root distroless base. To push manually to GHCR, log in, tag, and push; CI does this automatically after all gates pass.

## Kubernetes deployment

Prerequisites: a cluster with a default StorageClass, `kubectl`, a published GHCR image, and (for Ingress) an ingress-nginx controller. The raw manifest uses the `latest` image tag. Replace the demo Secret first, then apply:

```sh
kubectl apply -f kubernetes/all.yaml
kubectl -n devops-lab create secret generic taskboard-auth --from-literal=API_TOKEN='choose-a-unique-token' --from-literal=DB_PASSWORD='choose-a-db-password' --dry-run=client -o yaml | kubectl apply -f -
kubectl -n devops-lab rollout status statefulset/postgres
kubectl -n devops-lab rollout status deployment/taskboard
kubectl -n devops-lab port-forward service/taskboard 8080:80
```

Run the port-forward in its own terminal and visit `http://localhost:8080`. The sample Ingress host is `taskboard.example.test`; point that host to your controller and configure TLS before internet exposure. To install ingress-nginx: `helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --create-namespace` ([official guide](https://kubernetes.github.io/ingress-nginx/deploy/)). The HPA requires Metrics Server ([official chart](https://github.com/kubernetes-sigs/metrics-server/blob/master/charts/metrics-server/README.md)).

PostgreSQL uses the `ReadWriteOnce` PVC. The Go API is stateless and the HPA can scale it from 2 to 4 replicas based on CPU. The database remains one replica for this classroom deployment.

## Helm deployment

Use either raw manifests **or** Helm for the task board in one namespace, because both define equivalent resources. Helm expects a preexisting Secret:

```sh
kubectl create namespace devops-lab --dry-run=client -o yaml | kubectl apply -f -
kubectl -n devops-lab create secret generic taskboard-auth --from-literal=API_TOKEN='choose-a-unique-token' --from-literal=DB_PASSWORD='choose-a-db-password'
helm lint helm/taskboard
helm upgrade --install taskboard helm/taskboard -n devops-lab --set image.tag=latest
kubectl -n devops-lab rollout status deployment/taskboard
```

For Ingress, add `--set ingress.enabled=true --set ingress.host=your.example.com`. For a non-default StorageClass, set `storage.storageClassName`. GHCR may mark a newly created package private; make it public in package settings or configure an image pull secret before deploying.

## Terraform infrastructure

`terraform/` provisions a two-AZ VPC, public/private subnets, NAT gateway, EKS managed node group, core add-ons and EBS CSI driver. Requires Terraform >=1.6, AWS CLI, AWS credentials with sufficient permissions, and a supported EKS version in the chosen region. AWS resources incur charges until destroyed. State is local by default; configure a protected remote backend before team use.

```sh
cd terraform
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply
aws eks update-kubeconfig --region ap-south-1 --name session21-devops
kubectl get nodes
```

Review the plan and region/version before `apply`. Run `terraform destroy` from the same directory when finished. The Terraform configuration does not automatically install Argo CD, ingress-nginx or Metrics Server; install those cluster components separately.

## CI/CD pipeline and DevSecOps

`.github/workflows/ci-cd.yml` runs on pull requests and `main` pushes. Go tests and vet, React build, npm audit, Helm lint/render, Terraform validation, CodeQL SAST, Trivy filesystem SCA and Gitleaks history scanning must all pass. CodeQL alerts require a code scanning ruleset to act as a PR gate; Go vet fails the build for static-analysis errors. On `main`, the pipeline builds the Docker image, blocks HIGH/CRITICAL image findings, pushes immutable SHA and `latest` tags to GHCR, then commits that SHA into Helm values. Argo CD sees the Helm commit and deploys it. This is the Kubernetes deployment step: CI changes desired state; Argo CD performs cluster reconciliation. No cluster credentials are stored in GitHub Actions.

The GitOps commit is authored as `Yeagerist0 <180412661+Yeagerist0@users.noreply.github.com>`. GitHub Actions still appears as the push actor. To enforce gates, mark `test`, `manifests`, `terraform`, `sast`, `sca` and `secrets` as required checks in branch protection. If branch protection blocks bot writes to `main`, configure a deployment branch or a reviewed pull request for Helm value changes. See [`security/README.md`](security/README.md) for details.

## Monitoring and logs

The app exposes `/metrics` in Prometheus text format. `monitoring/prometheus.yaml` deploys Prometheus and Grafana, with Prometheus preconfigured as Grafana's data source:

```sh
kubectl apply -f monitoring/prometheus.yaml
kubectl -n monitoring port-forward service/prometheus 9090:9090
# In another terminal:
kubectl -n monitoring port-forward service/grafana 3000:3000
```

Open Prometheus at `http://localhost:9090` or Grafana at `http://localhost:3000` (`admin` / `replace-for-your-lab` for the local lab only). Query `app_http_requests_total`, `app_tasks_created_total`, or `app_tasks_stored`. Prometheus/Grafana use ephemeral storage in this demo; historical data disappears when pods are recreated. Application/database logs and Kubernetes events are available with:

```sh
kubectl -n devops-lab logs deployment/taskboard --tail=100
kubectl -n devops-lab logs statefulset/postgres --tail=100
kubectl -n devops-lab get events --sort-by=.lastTimestamp
kubectl -n devops-lab describe deployment taskboard
```

## GitOps workflow

Install [Argo CD](https://argo-cd.readthedocs.io/en/latest/getting_started/), then create the app Secret and apply `gitops/application.yaml`:

```sh
kubectl create namespace argocd
kubectl apply -n argocd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl create namespace devops-lab
kubectl -n devops-lab create secret generic taskboard-auth --from-literal=API_TOKEN='choose-a-unique-token' --from-literal=DB_PASSWORD='choose-a-db-password'
kubectl apply -f gitops/application.yaml
kubectl -n argocd get application taskboard
```

If the namespace or Secret already exists, keep it. Argo CD targets `helm/taskboard` on `main`, auto-syncs and repairs drift. A successful CI run changes `image.tag` to its commit SHA. The GHCR package must be readable by the cluster. Do not also run `kubectl apply -f kubernetes/all.yaml` against the same release.

## Troubleshooting challenge

Six isolated broken-resource patches are in [`kubernetes/challenges/`](kubernetes/challenges/README.md): bad image, absent ConfigMap, bad readiness path, Service label mismatch, absent Secret and incorrect database DNS. Apply each in a disposable cluster, investigate pods/events/logs/endpoints, fix it, and verify `/healthz`, the API and metrics. Copy `TEMPLATE.md` for each incident. The instructor answer key is separate in `ANSWERS.md`.

## Screenshots

![React task board running locally](screenshots/taskboard-local.png)

The [successful eight-job CI run](https://github.com/Yeagerist0/final-devops-project/actions/runs/38933948746) and [public GHCR package](https://github.com/users/Yeagerist0/packages/container/package/final-devops-project) are live evidence. Capture Argo CD sync and Prometheus query screenshots after deploying to your own cluster; no paid AWS cluster was started for this repository-only submission.

## Submission video

Watch the [80-second captioned project walkthrough](video/session21-demo.mp4). Its [slide source and spoken presentation script](video/README.md) cover the application, architecture, successful CI run, infrastructure, security, monitoring, GitOps and troubleshooting. The video is a repository/CI demonstration; it does not imply paid AWS resources were created.

## Lessons learned

- A passing build is only one gate; test, source, dependency, secret and image checks each catch different failures.
- Git holds desired deployment state, so a release can be audited and reconciled by Argo CD.
- Kubernetes events and resource descriptions often explain startup failures before application logs exist.
- PostgreSQL separates persistent data from stateless API replicas, allowing the HPA to scale the API safely.
