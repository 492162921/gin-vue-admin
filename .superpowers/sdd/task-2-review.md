# Task 2 Review: K8s Inspector (Real + Stub) + Unit Tests

**Reviewer:** task-scoped gate  
**Base:** `160187254e91ddcfeea2a40e41bccee58c08508a`  
**Head:** `3b6bcdec552541930b401317acc600b82fed0ed3`  
**Date:** 2026-07-20

---

## 1. Spec Compliance: ✅

| Requirement | Verdict |
|:---|:---|
| Files: `inspector.go`, `client.go`, `stub.go`, `inspector_test.go` | Met |
| `go.mod` / `go.sum`: `k8s.io/client-go` + metrics, Go 1.24 compatible (v0.32.4) | Met |
| DTOs: `NodeInfo`, `PodInfo`, `NamespaceInfo` with brief fields | Met |
| `Inspector` interface (5 methods) + `NewInspector(kubeconfigPath, forceStub)` | Met |
| `forceStub==true` or invalid kubeconfig → `stubInspector` | Met |
| `K8S_FORCE_STUB=1` OR’d into forceStub | Met |
| Stub: ≥3 nodes, multiple anomalous pods, namespaces derived from pods | Met |
| Client: kubeconfig file → client-go; metrics optional, CPU/Mem 0 on failure | Met |
| `ListNamespaces` added (client lists NS + pod counts) | Met |
| Brief TDD tests: `TestStubListPodsContainsAnomalies`, `TestForceStubIgnoresBadPath` | Met |
| Commit message matches brief Step 5 | Met |

**Missing / extra vs brief:** None blocking. Six additional unit tests (nodes, namespaces, version, env, bad-path fallback) are acceptable extras.

---

## 2. Global Constraints

| Constraint | Verdict |
|:---|:---|
| Read-only K8s | Met — client uses only `Discovery().ServerVersion()`, `List` on Nodes/Pods/Namespaces/NodeMetricses; no Create/Update/Delete/Patch/Apply |
| No secrets in logs | Met — package has no logging; errors wrap stat/parse failures without dumping kubeconfig contents |
| Insp domain scoped to `service/inspection/k8s` | Met — all logic in four files under that path; only `go.mod`/`go.sum` touched elsewhere |

---

## 3. Task Quality: **Approved**

### Critical
None.

### Important
None.

### Minor

1. **`TestEnvForceStubUnset` env hygiene** — Uses `os.Unsetenv` instead of `t.Setenv("", "")` / restore; can leak or flake if `K8S_FORCE_STUB` is set in the outer environment. Prefer `t.Setenv` for isolation.

2. **`ctx` ignored on discovery calls** — `clientInspector.TestConnection` / `GetVersion` accept `context.Context` but call `Discovery().ServerVersion()` without cancellation wiring. List methods correctly pass `ctx`. Low impact for Task 2; consider `RESTClient` context in a follow-up.

3. **`go mod tidy` churn** — Diff removes several unused indirect deps (e.g. alibaba/tencent SMS, `gorilla/websocket`) and bumps `golang.org/x/*`. Report claims `go build` PASS; no repo references found. Worth a quick full `server` build in CI, not a Task 2 blocker.

4. **Silent stub fallback** — Invalid kubeconfig returns connected stub (demo data). Intentional per brief and report; Task 3+ should surface misconfiguration to operators.

5. **No live-cluster integration test** — Real client path validated by compile + port from source project only. Acceptable for stub-only CI scope in brief.

---

## Summary

Task 2 delivers the specified K8s inspector: interface/DTOs, stub with demo anomalies, read-only client-go client with optional metrics, factory with env/path fallback, and passing unit tests. Global read-only / no-log / path-scope constraints hold. **Gate: pass.**
