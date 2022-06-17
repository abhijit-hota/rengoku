const BASE_URL = "http://localhost:8080";
export default async function api(path, method = "GET", body = null) {
  console.debug();
  const res = await fetch(BASE_URL + "/api" + path, {
    method,
    body: body ? (body instanceof FormData ? body : JSON.stringify(body)) : null,
  });
  if (!res.ok) {
    throw new Error(res.statusText);
  }
  return await res.json();
}
