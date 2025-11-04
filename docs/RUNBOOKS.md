# Operational Runbooks for COINjecture v4.0

**Last Updated:** 2025-11-04
**Owner:** DevOps / SRE Team
**On-Call:** [PagerDuty rotation]

---

## Table of Contents

1. [Deployment Procedures](#deployment-procedures)
2. [Rollback Procedures](#rollback-procedures)
3. [Incident Response](#incident-response)
4. [Monitoring & Alerts](#monitoring--alerts)
5. [Soak Testing](#soak-testing)

---

## Deployment Procedures

### Pre-Deployment Checklist

- [ ] All CI/CD quality gates passed (ci.yml)
- [ ] Determinism tests passed across all platforms (determinism.yml)
- [ ] Parity tests show 100% match rate (parity.yml)
- [ ] Security scans clean (security.yml)
- [ ] SBOMs generated and signed
- [ ] Release notes written
- [ ] Rollback plan documented
- [ ] On-call engineer notified

### Gradual Rollout Strategy

**Phase 1: Canary (10% traffic, 2 hours)**

```bash
# 1. Deploy to canary node
ssh canary.coinjecture.com
cd /opt/coinjecture
git fetch origin
git checkout v4.0.0

# 2. Build and install
./scripts/build_and_install.sh

# 3. Set feature flag
export CODEC_MODE=shadow

# 4. Restart daemon
sudo systemctl restart coinjectured

# 5. Monitor metrics
watch -n 5 'curl -s http://localhost:9090/metrics | grep parity'

# 6. Check logs for parity drifts
tail -f /var/log/coinjecture/daemon.log | grep "PARITY DRIFT"
```

**Health Checks:**
- [ ] API responds: `curl http://canary:12346/health`
- [ ] Metrics exporting: `curl http://canary:9090/metrics`
- [ ] P2P peers connected: `coinjectured status | grep peers`
- [ ] No parity drifts in logs
- [ ] Error rate < 0.1%
- [ ] p95 latency < 2x baseline

**If canary fails:** Execute [Rollback Procedure](#rollback-procedures)

---

**Phase 2: Gradual (25% traffic, 6 hours)**

```bash
# Deploy to 25% of nodes
for node in node-{01..05}; do
  ssh $node.coinjecture.com "./scripts/deploy_v4.sh"
done

# Monitor aggregate metrics
./scripts/monitor_deployment.sh --target-version 4.0.0 --coverage 25%
```

**Health Checks:**
- [ ] Parity 100% across all deployed nodes
- [ ] CID pin quorum success > 95%
- [ ] No state violations
- [ ] User-reported issues: 0

---

**Phase 3: Majority (75% traffic, 24 hours)**

```bash
# Deploy to 75% of nodes
./scripts/rollout.sh --version 4.0.0 --coverage 75% --wait 300

# Enable stricter feature flag
for node in node-{01..15}; do
  ssh $node.coinjecture.com "export CODEC_MODE=refactored_primary; sudo systemctl restart coinjectured"
done
```

**Health Checks:**
- [ ] 7-day soak test passed
- [ ] Zero legacy fallbacks
- [ ] No drift alerts
- [ ] Performance metrics within SLO

---

**Phase 4: Full (100% traffic)**

```bash
# Deploy to all nodes
./scripts/rollout.sh --version 4.0.0 --coverage 100%

# Switch to refactored-only mode (legacy removed)
export CODEC_MODE=refactored_only

# Archive legacy code
git tag legacy-python-final
git branch -D legacy-compat
```

**Post-Deployment:**
- [ ] Update documentation
- [ ] Archive old binaries (S3 cold storage)
- [ ] Send release announcement
- [ ] Post-mortem (lessons learned)

---

## Rollback Procedures

### Emergency Rollback (< 5 minutes)

**Trigger Conditions:**
- Parity drift > 1%
- State violation detected
- API error rate > 5%
- P2P network partition

**Steps:**

```bash
# 1. STOP: Halt deployment immediately
./scripts/halt_rollout.sh

# 2. Revert to previous version
for node in $(./scripts/list_v4_nodes.sh); do
  ssh $node "cd /opt/coinjecture && git checkout v3.17.0 && sudo systemctl restart coinjectured" &
done

# 3. Verify rollback
./scripts/verify_version.sh --expected v3.17.0

# 4. Monitor recovery
watch -n 5 './scripts/health_check.sh'

# 5. Incident notification
./scripts/notify_incident.sh --severity critical --message "v4.0.0 rollback executed"
```

**Expected Recovery Time:** < 5 minutes
**Data Loss:** None (blockchain state preserved)

---

### Planned Rollback (testing/staging)

```bash
# 1. Set feature flag to legacy mode
export CODEC_MODE=legacy_only

# 2. Restart daemons (no code change needed)
ansible-playbook playbooks/restart_daemons.yml

# 3. Verify metrics
curl http://prometheus:9090/api/v1/query?query=coinjecture_codec_mode

# 4. Monitor for 1 hour before removing v4 binaries
```

---

## Incident Response

### Severity Levels

| Level | Response Time | Example |
|-------|---------------|---------|
| **P0 (Critical)** | < 15 min | State violation, consensus fork, data loss |
| **P1 (High)** | < 1 hour | Parity drift, CID quorum failure, API down |
| **P2 (Medium)** | < 4 hours | Elevated error rate, performance degradation |
| **P3 (Low)** | < 24 hours | Non-critical bugs, UI issues |

### P0: State Violation

**Symptoms:**
- `InflationDetected` error in logs
- Balance invariant failure
- State root mismatch

**Response:**

```bash
# 1. IMMEDIATE: Halt block production
./scripts/emergency_halt.sh

# 2. Identify bad block
grep "StateRootMismatch" /var/log/coinjecture/*.log

# 3. Reorg to last good block
coinjectured reorg --to-block <last_good_hash>

# 4. Investigate root cause
./scripts/forensics.sh --issue state-violation --block <bad_block_hash>

# 5. Coordinate network-wide reorg (if needed)
# Requires consensus from node operators
```

**Escalation:** Notify all node operators immediately

---

### P1: Parity Drift

**Symptoms:**
- `PARITY DRIFT` in logs
- Legacy and refactored hashes differ
- Prometheus: `coinjecture_parity_drifts_total` increasing

**Response:**

```bash
# 1. Switch to legacy mode (safe fallback)
export CODEC_MODE=legacy_only
sudo systemctl restart coinjectured

# 2. Collect parity logs
journalctl -u coinjectured --since "1 hour ago" | grep "PARITY" > parity_incident.log

# 3. Generate diff report
./scripts/parity_diff_analysis.sh --log parity_incident.log

# 4. File GitHub issue with label: parity-drift
gh issue create --title "Parity drift detected" --label parity-drift --body "$(cat parity_incident.log)"

# 5. Schedule hot fix PR
```

**Resolution:** Fix drift in refactored code, validate with 1000-block parity test

---

### P1: CID Quorum Failure

**Symptoms:**
- `PinQuorumFailed` errors
- Blocks rejected due to missing CID
- IPFS nodes unreachable

**Response:**

```bash
# 1. Check IPFS node health
for node in ipfs-{1..3}.coinjecture.com; do
  curl -s http://$node:5001/api/v0/id | jq '.ID'
done

# 2. Restart failed IPFS nodes
ansible-playbook playbooks/restart_ipfs.yml

# 3. Re-pin missing CIDs from backup
./scripts/recover_missing_cids.sh --from-backup

# 4. Adjust quorum temporarily (if 1 node down)
# In config.yaml: pin_quorum: "1/2" (instead of "2/3")
# Revert once node recovered

# 5. Run CID audit
./scripts/audit_cids.sh --repair
```

---

## Monitoring & Alerts

### Key Metrics (Prometheus)

**Parity Metrics:**
```promql
# Parity match rate (should be 100%)
rate(coinjecture_parity_matches_total[5m]) /
(rate(coinjecture_parity_matches_total[5m]) + rate(coinjecture_parity_drifts_total[5m]))

# Parity drift count (should be 0)
increase(coinjecture_parity_drifts_total[1h])
```

**Performance Metrics:**
```promql
# Verification duration p95 (should be < tier budget)
histogram_quantile(0.95, rate(coinjecture_verification_duration_ms_bucket[5m]))

# API latency p95 (should be < 100ms)
histogram_quantile(0.95, rate(http_request_duration_ms_bucket[5m]))
```

**IPFS Metrics:**
```promql
# Pin quorum success rate (should be > 95%)
rate(coinjecture_pin_quorum_success_total[5m]) /
rate(coinjecture_pin_quorum_success_total[5m] + coinjecture_pin_quorum_failures_total[5m])
```

### Alert Rules

**alert-rules.yml:**
```yaml
groups:
  - name: coinjecture_critical
    interval: 30s
    rules:
      - alert: ParityDriftDetected
        expr: increase(coinjecture_parity_drifts_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Parity drift detected between legacy and refactored"
          description: "{{ $value }} parity drifts in last 5 minutes"

      - alert: StateViolation
        expr: increase(coinjecture_state_violations_total[1m]) > 0
        for: 0s
        labels:
          severity: critical
        annotations:
          summary: "State violation detected (CRITICAL)"
          description: "Immediate investigation required"

      - alert: PinQuorumFailure
        expr: rate(coinjecture_pin_quorum_failures_total[5m]) > 0.05
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "IPFS pin quorum failures elevated"
          description: "{{ $value }}% of pins failing quorum"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "API error rate elevated"
          description: "{{ $value }}% 5xx errors"
```

---

## Soak Testing

### Pre-Production Soak Test

**Duration:** 7 days
**Environment:** Staging (3-node cluster)
**Load:** 10 blocks/hour, 100 tx/block

```bash
# 1. Deploy to staging
ansible-playbook -i inventory/staging playbooks/deploy_v4.yml

# 2. Configure shadow mode
export CODEC_MODE=shadow

# 3. Start load generator
./scripts/load_generator.sh --rate 10 --duration 7d

# 4. Monitor continuously
./scripts/soak_monitor.sh --alert-on-drift

# 5. Daily health checks
for day in {1..7}; do
  ./scripts/daily_health_check.sh --day $day
  sleep 86400
done

# 6. Generate soak report
./scripts/soak_report.sh --output SOAK_REPORT.md
```

**Success Criteria:**
- [ ] Zero parity drifts (100% match rate)
- [ ] No state violations
- [ ] API p95 latency < 100ms
- [ ] Pin quorum success > 99%
- [ ] No memory leaks (RSS stable)
- [ ] No panics or crashes

**If soak test fails:** Do NOT proceed to production. Root cause analysis required.

---

## Appendix

### Quick Reference

**Logs:**
- Main daemon: `/var/log/coinjecture/daemon.log`
- Parity: `/var/log/coinjecture/parity.log`
- Metrics: `http://localhost:9090/metrics`

**Commands:**
```bash
# Check version
coinjectured --version

# Check status
coinjectured status

# View parity stats
coinjectured parity-report

# Emergency stop
sudo systemctl stop coinjectured

# View real-time logs
journalctl -u coinjectured -f
```

**Contacts:**
- On-Call: [PagerDuty]
- Engineering: adz@alphx.io
- GitHub Issues: https://github.com/Quigles1337/COINjecture1337-REFACTOR/issues

---

**Document Version:** 1.0
**Review Schedule:** Quarterly
**Next Review:** 2025-02-04
