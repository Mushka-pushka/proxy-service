import http from 'k6/http';
import { check } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 80 },
        { duration: '2m', target: 100 },
        { duration: '30s', target: 0 },
    ],
    rps: 500, 
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<100'],
    },
    cloud: {
        projectID: 7832713,
        name: 'Baseline - without proxy (500 RPS)',
    },
};

export default function () {
    const res = http.get('http://localhost:8081/hello');
    
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 100ms': (r) => r.timings.duration < 100,
    });
    
}