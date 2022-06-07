<script>
  import { onMount } from "svelte";

  import { bookmarks, queryParams, queryStr } from "../lib/stores";
  import Bookmark from "./Bookmark.svelte";

  onMount(() => bookmarks.fetch($queryStr));
  // TODO: queryParams.subscribe
  $: bookmarks.fetch($queryStr);

  // Child state
  let marked = [];

  const toggleMark = (bookmarkID) => {
    const index = marked.indexOf(bookmarkID);
    const notPresent = index === -1;

    if (notPresent) {
      marked.push(bookmarkID);
    } else {
      marked.splice(index, 1);
    }

    marked = marked;
  };
</script>

<div class="sticky">
  <h2 style="padding-left: 0.5rem;">Showing 20 of 201</h2>
  <hr />
</div>
{#each $bookmarks as bookmark}
  <Bookmark {bookmark} {toggleMark} />
{/each}
<div
  style="text-align: center;"
  on:click={() => {
    $queryParams.page++;
  }}
>
  <button>Load More</button>
</div>
