<script>
import { bookmarks } from '../lib/stores';

	import { api } from '../lib/api';
	export let bookmark;
	let hovered = false;

	const deleteBookmark = async () => {
		try {
			const res = await api('/bookmarks/' + bookmark.id, 'DELETE');
			bookmarks.delete(bookmark.id)
		} catch (error) {
			console.debug(error);
		}
	};
	const activate = () => {
		hovered = true;
	};
	const deActivate = () => {
		hovered = false;
	};
</script>

<div class="bookmark" on:mouseover={activate} on:focus={activate} on:mouseout={deActivate} on:blur={deActivate}>
	<div class="row">
		{#if bookmark.meta.favicon}
			<img class="favicon" src={bookmark.meta.favicon} alt="Favicon" />
		{/if}
		<h3>{bookmark.meta.title}</h3>
	</div>
	<small><a href={bookmark.url}>{bookmark.url}</a> </small>
	<p>{bookmark.meta.description}</p>
	<div class="m-l-auto">
		<button style="background-color: var(--text-muted); color: var(--background)" on:click={deleteBookmark}>
			Delete
		</button>
	</div>
	<div class="footer row">
		<div class="tags">
			{#each bookmark.tags as tag}<div class="tag">{tag.name}</div>{/each}
		</div>
	</div>
</div>
