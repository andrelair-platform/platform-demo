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

export const options = {
  stages: [
    { duration: '10s', target: 5 },   // ramp up to 5 VUs
    { duration: '20s', target: 5 },   // hold — generates Prometheus data
    { duration: '5s',  target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_failed:                     ['rate<0.01'],   // < 1% HTTP errors
    http_req_duration:                   ['p(95)<500'],   // p95 < 500ms
    'http_req_duration{endpoint:root}':  ['p(95)<200'],   // root response fast
    'http_req_duration{endpoint:health}':['p(95)<50'],    // health check sub-50ms
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
