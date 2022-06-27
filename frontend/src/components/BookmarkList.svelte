<script lang="ts">
  import { onMount, getContext } from "svelte";

  import Bookmark from "./Bookmark.svelte";

  import { store } from "@lib";
  const { bookmarks, queryParams, queryStr } = store;
  import { modals } from "@Modal";
  import Popup from "./Popup.svelte";

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
    <button on:click={() => $modals["add-bookmark"].showModal()}> + Add Bookmark </button>
    <div class="m-l-auto row">
      {#if marked.length > 0}
        <span style="padding-right: 1em">
          {marked.length} bookmark{marked.length > 1 ? "s" : ""} selected
        </span>
        <Popup>
          <svelte:fragment slot="list-items">
            <li role="none">
              <button class="no-style" on:click={() => $modals["add-bookmark"].showModal()}>
                + Add Bookmark
              </button>
            </li>
            <li role="none">
              <button class="no-style" on:click={() => $modals["add-folder"].showModal()}>
                + Move to Folder
              </button>
            </li>
            <li role="none">
              <button class="no-style" on:click={() => $modals["add-tag"].showModal()}>
                + Add Tags
              </button>
            </li>
          </svelte:fragment>
        </Popup>
      {/if}
    </div>
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
