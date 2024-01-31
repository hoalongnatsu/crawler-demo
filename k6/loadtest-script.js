import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 30, // vus (virtual users)
  duration: '30s',
};

export default function () {
  http.get('http://localhost:3000/posts');
  sleep(0.5);
}