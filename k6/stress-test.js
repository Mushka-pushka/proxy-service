import http from 'k6/http';
import { check } from 'k6';  

export const options = {
    stages: [
        { duration: '2m', target: 20 },     
        { duration: '2m', target: 40 },     
        { duration: '2m', target: 60 },     
        { duration: '2m', target: 80 },     
        { duration: '2m', target: 100 },    
        { duration: '1m', target: 0 },      
    ],
    rps: 5000,
    thresholds: {
        http_req_failed: ['rate<0.10'],
    },
    cloud: {
        projectID: 7832713,
        name: 'Stress test - 100 VU max',
    },
};

export default function () {
    const res = http.get('http://localhost:8080/hello');
    
    check(res, {
        'status is 200': (r) => r.status === 200,
    });
    
}