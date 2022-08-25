const BASE_URL = "http://localhost:8080";

export default async function api(path: string, method: string = "GET", body = null) {
  const options: RequestInit = {
    method,
    body: body ? (body instanceof FormData ? body : JSON.stringify(body)) : null,
  };

  if (!(body instanceof FormData)) {
    options.headers = {
      ...options.headers,
      "Content-Type": "application/json",
    };
  }

  if (path !== "/auth/login") {
    options.headers = {
      ...options.headers,
      Authorization: "Bearer " + localStorage.getItem("AUTH_TOKEN"),
    };
  }

  const res = await fetch(BASE_URL + "/api" + path, options);
  if (!res.ok) {
    const error = new Error();
    error.cause = (await res.json()).cause;
    throw error;
  }
  return await res.json();
}
