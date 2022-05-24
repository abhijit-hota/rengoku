<script>
	import './styles/water.min.css';
	import './styles/style.css';
	import { bookmarks, queryStr } from './lib/stores';
	import Bookmark from './components/Bookmark.svelte';
	import Filter from './components/Filter.svelte';
	import SettingsModal from './components/SettingsModal.svelte';
	import AddBookmarkModal from './components/AddBookmarkModal.svelte';
	import { onMount } from 'svelte';

	onMount(() => bookmarks.fetch($queryStr));
	$: bookmarks.fetch($queryStr);
</script>

<nav>
	<div class="row">
		<h1>Rengoku</h1>
		<button
			class="m-l-auto"
			on:click={() => {
				// @ts-ignore
				document.getElementById('add-bookmark-modal').showModal();
			}}
		>
			+ Add Bookmark
		</button>
	</div>
	<hr />
</nav>
<aside id="folders" />
<main>
	{#each $bookmarks as bookmark}
		<Bookmark {bookmark} />
	{/each}
</main>
<aside id="filters">
	<Filter />
</aside>
<SettingsModal />
<AddBookmarkModal />
