<script lang="ts">
  import { onMount, getContext } from "svelte";

  import Bookmark from "./Bookmark.svelte";

  import { bookmarks, queryParams, queryStr } from "../lib/stores";
  import { modals } from "../components/Modal.svelte";

  onMount(() => bookmarks.fetch($queryStr));
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
  <div class="row">
    <h2 style="padding-left: 0.5rem;">Showing 20 of 201</h2>
    <button class="m-l-auto" on:click={() => $modals["add-bookmark"].showModal()}>
      + Add Bookmark
    </button>
  </div>
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
