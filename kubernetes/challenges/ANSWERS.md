# Instructor answer key

| Fault | Symptom | Root cause | Repair |
| --- | --- | --- | --- |
| 01 | `ImagePullBackOff` | Image tag does not exist | Set image to a published SHA or `latest` |
| 02 | `CreateContainerConfigError` | Referenced ConfigMap is missing | Restore `taskboard-config` reference |
| 03 | Pod running, `0/1 Ready`; Service has no ready endpoints | Readiness URL returns 404 | Restore `/readyz` |
| 04 | Service endpoints empty despite ready pod | Service selector misses pod label | Restore `app: taskboard` |
| 05 | `CreateContainerConfigError` | Referenced Secret is missing | Restore `taskboard-auth` reference |
| 06 | App `CrashLoopBackOff` and DB connection log error | `DB_HOST` points to a nonexistent Service | Restore `postgres` in ConfigMap and restart Deployment |

Useful proof: `kubectl -n devops-lab get pods,svc,endpoints`, `kubectl -n devops-lab describe pod POD`, `kubectl -n devops-lab logs deploy/taskboard`, and `curl http://localhost:8080/healthz` after port-forwarding. For 01, 02, and 05, container logs may be absent because the container never started; inspect pod events. For 03, `kubectl describe pod` shows readiness probe 404. For 04, compare Service selector with pod labels.
