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
const expectedResponse = http.expectedStatuses(200, 401, 429);

export default function () {
  const headers = token ? { Authorization: `Bearer ${token}` } : {};
  const res = http.get(`${baseUrl}/v1/protected/ping`, {
    headers,
    responseCallback: expectedResponse,
  });

  // 200 (authorized), 401 (no token), and 429 (rate-limited) are valid outcomes.
  check(res, {
    "status is 200, 401, or 429": (r) =>
      r.status === 200 || r.status === 401 || r.status === 429,
  });

  sleep(0.1);
}
