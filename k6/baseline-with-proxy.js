import http from 'k6/http';
import { check } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 100 },   
        { duration: '30s', target: 500 },    
        { duration: '1m', target: 500 },     
        { duration: '30s', target: 0 },      
    ],
    rps: 500,  
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<100'],
    },
};

export default function () {
    const res = http.get('http://localhost:8080/hello');
    
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 100ms': (r) => r.timings.duration < 100,
    });
}


