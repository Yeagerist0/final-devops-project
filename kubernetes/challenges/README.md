# Troubleshooting challenge

Run against a disposable local cluster and apply **one fault at a time**. Start from `kubectl apply -f kubernetes/all.yaml`. For Deployment faults:

```sh
kubectl -n devops-lab patch deployment taskboard --type=strategic --patch-file kubernetes/challenges/01-image.yaml
```

Replace `01-image.yaml` with `02-configmap.yaml`, `03-probe.yaml`, or `05-secret.yaml`. For the Service fault:

```sh
kubectl -n devops-lab patch service taskboard --type=merge --patch-file kubernetes/challenges/04-service.yaml
```

For each fault, record symptoms, commands, root cause, exact fix and proof of recovery in a copy of `TEMPLATE.md`. Use `kubectl get pods -n devops-lab`, `kubectl describe pod`, `kubectl logs`, `kubectl get endpoints`, and `kubectl get events --sort-by=.lastTimestamp` as appropriate. Restore the healthy resource with `kubectl apply -f kubernetes/all.yaml` before the next fault; if a field lingers because of patch ownership, explicitly patch that field back using `ANSWERS.md`.

See `ANSWERS.md` only after investigation.
