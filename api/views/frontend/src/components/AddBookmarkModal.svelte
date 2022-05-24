<script>
	import { bookmarks } from '../lib/stores';

	import { api } from '../lib/api';

	let status, message;
</script>

<dialog id="add-bookmark-modal" style="width: 300px">
	<header><h2>Add new bookmark</h2></header>
	<form
		class="m-l-auto col"
		on:submit|preventDefault={async (e) => {
			status = 'SUBMITTING';
			const formData = new FormData(e.currentTarget);
			const dataToPost = Object.fromEntries(formData.entries());
			try {
				const tags = [].filter((option) => option.selected).map((option) => +option.value);
				const newBookmark = await api('/bookmarks', 'POST', dataToPost);
				status = 'SUCCESS';
				bookmarks.add(newBookmark);
				// @ts-ignore
				document.getElementById('add-bookmark-modal').close();
			} catch (error) {
				status = 'ERROR';
				message = 'An error occurred';
			}
		}}
	>
		<div class="col m-b-1">
			<label for="url"><strong>URL</strong><span class="red">*</span></label>
			<input type="url" placeholder="Add new URL" name="url" id="url" required />
		</div>

		<div class="col m-b-1">
			<label for="tags"><strong>Tags</strong></label>
			<!-- <TagInput/> -->
		</div>
		<span class="m-b-1 red" x-show="status === 'ERROR'" x-text="message" />
		<button type="submit" disabled={status === 'SUBMITTING'}>{status === 'SUBMITTING' ? 'Loading' : 'Add'}</button>
	</form>
</dialog>
