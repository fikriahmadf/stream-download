import http from 'k6/http';
import { check, sleep } from 'k6';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

// Load test configuration
export const options = {
    stages: [
        { duration: '30s', target: 10 },  // Ramp up to 10 users in 30s
        { duration: '1m', target: 10 },   // Stay at 10 users for 1 minute
        { duration: '30s', target: 20 },  // Ramp up to 20 users
        { duration: '1m', target: 20 },   // Stay at 20 users for 1 minute
        { duration: '30s', target: 0 },   // Ramp down to 0
    ],
    thresholds: {
        http_req_duration: ['p(95)<30000'], // 95% requests should be < 30s
        http_req_failed: ['rate<0.1'],      // Error rate < 10%
    },
};

// Save summary to file
export function handleSummary(data) {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    return {
        [`k6/results/summary-${timestamp}.json`]: JSON.stringify(data, null, 2),
        [`k6/results/summary-${timestamp}.txt`]: textSummary(data, { indent: ' ', enableColors: false }),
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
    };
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const payload = JSON.stringify({
    urls: [
        'http://192.168.1.6:4566/my-bucket/test/1.pdf',
        'http://192.168.1.6:4566/my-bucket/test/2.pdf',
        'http://192.168.1.6:4566/my-bucket/test/3.pdf',
        'http://192.168.1.6:4566/my-bucket/test/4.pdf',
        'http://192.168.1.6:4566/my-bucket/test/5.pdf',
        'http://192.168.1.6:4566/my-bucket/test/6.pdf',
        'http://192.168.1.6:4566/my-bucket/test/7.pdf',
        'http://192.168.1.6:4566/my-bucket/test/8.pdf',
        'http://192.168.1.6:4566/my-bucket/test/9.pdf',
        'http://192.168.1.6:4566/my-bucket/test/10.pdf',
    ],
});

const params = {
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: '180s', // 3 minutes
};

export default function () {
    const res = http.post(`${BASE_URL}/api/download`, payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response is zip': (r) => r.headers['Content-Type'] === 'application/zip',
    });

    sleep(1); // Wait 1 second between requests
}
