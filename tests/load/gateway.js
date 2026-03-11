import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 20,
  duration: "20s",
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<200"],
  },
};

const baseUrl = __ENV.GATEWAY_BASE_URL || "http://localhost:8080";
const token = __ENV.GATEWAY_JWT || "";

export default function () {
  const headers = token ? { Authorization: `Bearer ${token}` } : {};
  const res = http.get(`${baseUrl}/v1/protected/ping`, { headers });

  // Load test is valid in either authorized (200) or unauthorized (401) mode.
  check(res, {
    "status is 200 or 401": (r) => r.status === 200 || r.status === 401,
  });

  sleep(0.1);
}
