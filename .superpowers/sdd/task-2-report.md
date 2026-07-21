# Task 2 Report: K8s Inspector (Real + Stub) + Unit Tests

**Status:** DONE  
**Branch:** `feat/k8s-inspection`  
**Date:** 2026-07-20

## Summary

Implemented read-only K8s inspector under `server/service/inspection/k8s/` with stub fallback, real client-go integration, and unit tests (TDD).

## Files Created

| File | Purpose |
|------|---------|
| `inspector.go` | DTOs (`NodeInfo`, `PodInfo`, `NamespaceInfo`), `Inspector` interface, `NewInspector` factory |
| `stub.go` | `stubInspector` — 3 nodes, 7 pods (multiple anomalies), 2 namespaces |
| `client.go` | `clientInspector` — client-go read-only: nodes, pods, namespaces, optional metrics |
| `inspector_test.go` | 8 unit tests covering stub behavior, forceStub, env var, bad kubeconfig fallback |

## Files Modified

| File | Change |
|------|--------|
| `server/go.mod` | Added `k8s.io/client-go@v0.32.4`, `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/metrics`; pinned transitive deps for Go 1.24 |
| `server/go.sum` | Updated checksums |

## Interface

```go
type Inspector interface {
    TestConnection(ctx context.Context) error
    GetVersion(ctx context.Context) (string, error)
    ListNodes(ctx context.Context) ([]NodeInfo, error)
    ListPods(ctx context.Context, namespace string) ([]PodInfo, error)
    ListNamespaces(ctx context.Context) ([]NamespaceInfo, error)
}

func NewInspector(kubeconfigPath string, forceStub bool) Inspector
```

## Behavior

- `forceStub==true` OR `K8S_FORCE_STUB=1` → always `stubInspector`
- Invalid/missing kubeconfig → fallback to `stubInspector` (connected=true for demo)
- Real client: metrics-server optional; CPU/Mem usage 0 when unavailable (no error)
- `ListNamespaces`: client lists all NS + pod counts; stub derives from stub pods
- Field naming aligned with brief: `CPUUsage`/`MemUsage`, `RestartCount`, `CrashLoop`, `Status` (Ready/NotReady)

## Dependencies

```
k8s.io/client-go v0.32.4
k8s.io/api v0.32.4
k8s.io/apimachinery v0.32.4
k8s.io/metrics v0.32.4
```

Pinned for Go 1.24 compatibility (latest v0.36.x requires Go 1.26):
- `k8s.io/kube-openapi@v0.0.0-20241105132330-32ad38e42d3f`
- `sigs.k8s.io/yaml@v1.4.0`
- `github.com/google/gnostic-models@v0.6.8`

## TDD Steps

1. Wrote `inspector_test.go` first → compile FAIL (undefined `NewInspector`)
2. Implemented stub + client + factory
3. Resolved dependency conflicts (structured-merge-diff v4/v6, yaml fork)
4. All tests PASS; `go build` PASS

## Verification

```bash
cd server
go test ./service/inspection/k8s/ -count=1   # ok, 8 tests, ~0.2s
go build -o NUL .                             # success
```

## Commit

```
feat(inspection): add k8s inspector with stub and real client
```

## Concerns / Notes

1. **Dependency pinning:** `@latest` client-go (v0.36) bumps Go to 1.26; stayed on v0.32.4 with manual transitive pins.
2. **Real cluster tests:** No integration test against live K8s (stub-only CI); real client path validated by compile + port from source project.
3. **Stub connected always true on fallback:** Bad kubeconfig silently serves demo data — intentional per brief; callers should validate cluster config separately in Task 3+.

## Next Task

Task 3: Cluster CRUD service using `NewInspector` for connection test on create/update.
