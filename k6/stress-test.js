import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 100 },    
        { duration: '1m', target: 500 },    
        { duration: '1m', target: 1000 },   
        { duration: '1m', target: 2000 },   
        { duration: '1m', target: 3000 },   
        { duration: '1m', target: 4000 },   
        { duration: '1m', target: 5000 },   
        { duration: '1m', target: 0 },      
    ],
    thresholds: {
        http_req_failed: ['rate<0.10'],    
    },
};

export default function () {
    const res = http.get('http://localhost:8080/hello');
    
    check(res, {
        'status is 200': (r) => r.status === 200,
    });
    
    sleep(0.1);
}