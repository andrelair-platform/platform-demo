/**
 * k6 smoke load test for platform-demo.
 *
 * Usage (local):
 *   k6 run --env BASE_URL=http://localhost:9898 k6/smoke.js
 *
 * Usage (CI, via kubectl port-forward):
 *   kubectl port-forward svc/platform-demo 9898:9898 -n platform-demo-dev &
 *   k6 run k6/smoke.js
 *
 * Thresholds are tuned to catch hard regressions, not to gate on
 * sub-millisecond improvements. p95 < 500ms is generous for an in-cluster
 * HTTP handler — real regressions (GC pause, blocking call, OOM) show up well
 * above that threshold.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const rootLatency = new Trend('root_latency', true);

// When CANARY_DURATION is set (by the CI canary-load job), run a sustained
// constant load for that duration so k6 traffic is flowing during the entire
// Argo Rollouts AnalysisRun window. Without this, the analysis fires after
// the Rollout is already Healthy and the gates have no real traffic data.
const CANARY_DURATION = __ENV.CANARY_DURATION;

export const options = CANARY_DURATION
  ? {
      stages: [
        { duration: '10s',           target: 3 },
        { duration: CANARY_DURATION, target: 3 },
        { duration: '10s',           target: 0 },
      ],
      thresholds: {
        http_req_failed:   ['rate<0.01'],
        http_req_duration: ['p(95)<500'],
        errors:            ['rate<0.01'],
      },
    }
  : {
      stages: [
        { duration: '10s', target: 5 },
        { duration: '20s', target: 5 },
        { duration: '5s',  target: 0 },
      ],
      thresholds: {
        http_req_failed:                     ['rate<0.01'],
        http_req_duration:                   ['p(95)<500'],
        'http_req_duration{endpoint:root}':  ['p(95)<200'],
        'http_req_duration{endpoint:health}':['p(95)<50'],
        errors:                              ['rate<0.01'],
      },
    };

const BASE_URL = __ENV.BASE_URL || 'http://localhost:9898';

export default function () {
  // Root endpoint — JSON payload
  {
    const res = http.get(`${BASE_URL}/`, { tags: { endpoint: 'root' } });
    rootLatency.add(res.timings.duration);
    const ok = check(res, {
      'root: status 200':              (r) => r.status === 200,
      'root: content-type json':       (r) => r.headers['Content-Type']?.startsWith('application/json'),
      'root: body has app field':      (r) => JSON.parse(r.body)?.app === 'platform-demo',
    });
    errorRate.add(!ok);
  }

  // Health check — must be fast
  {
    const res = http.get(`${BASE_URL}/healthz`, { tags: { endpoint: 'health' } });
    check(res, {
      'healthz: status 200': (r) => r.status === 200,
      'healthz: body ok':    (r) => r.body.trim() === 'ok',
    });
    errorRate.add(res.status !== 200);
  }

  // Readiness probe
  {
    const res = http.get(`${BASE_URL}/readyz`, { tags: { endpoint: 'ready' } });
    check(res, { 'readyz: status 200': (r) => r.status === 200 });
    errorRate.add(res.status !== 200);
  }

  sleep(1);
}
