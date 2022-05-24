const BASE_URL = 'http://localhost:8080';
export async function api(path, method = 'GET', body = null) {
	const res = await fetch(BASE_URL + '/api' + path, {
		method,
		body: body ? JSON.stringify(body) : null,
	});
	if (!res.ok) {
		throw new Error(res.statusText);
	}
	return await res.json();
}
