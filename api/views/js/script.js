const BASE_URL = 'http://localhost:8080';
async function api(path, method = 'GET', body = null) {
	const res = await fetch(BASE_URL + '/api' + path, {
		method,
		body: body ? JSON.stringify(body) : null,
	});
	if (!res.ok) {
		throw new Error(res.statusText);
	}
	return await res.json();
}

document.addEventListener('alpine:init', () => {
	Alpine.store('queryParams', {
		sortBy: 'date',
		order: 'asc',
		folder: '',
		tags: [],
		search: '',

		get() {
			const queryParams = [];
			if (this.sortBy && this.order) {
				const param = `sort_by=${this.sortBy}&order=${this.order}`;
				queryParams.push(param);
			}
			if (this.folder) {
				const param = `folder=${this.folder}`;
				queryParams.push(param);
			}
			if (this.search) {
				const param = `search=${this.search}`;
				queryParams.push(param);
			}
			if (this.tags.length > 0) {
				const tagStr = this.tags.map((id) => `tags[]=${id}`).join('&');
				const param = tagStr;
				queryParams.push(param);
			}
			return '?' + queryParams.join('&');
		},

		toggleTag(tag) {
			if (this.tags.includes(tag)) {
				const toDelete = this.tags.indexOf(tag);
				this.tags.splice(toDelete, 1);
			} else {
				this.tags.push(tag);
			}
		},
	});
});
