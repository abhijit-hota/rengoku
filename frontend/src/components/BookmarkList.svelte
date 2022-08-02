<script lang="ts">
  import { faFolderPlus, faTags, faTrash } from "@fortawesome/free-solid-svg-icons";
  import { toast } from "@zerodevx/svelte-toast";

  import { Bookmark } from "@components";
  import { api, store } from "@lib";
  import { modals } from "@Modal";
  import BatchActionButton from "./BatchActionButton.svelte";

  const { bookmarks, queryParams, queryStr, stats } = store;

  const getBookmarks = async () => {
    try {
      const res = await api("/bookmarks" + $queryStr);
      return res;
    } catch (error) {
      // TODO
      console.debug(error);
    }
  };

  let currentPage = 0;

  queryStr.subscribe((q) => {
    if ($queryParams.page !== currentPage && $queryParams.page !== 0) {
      currentPage = $queryParams.page;
      getBookmarks().then((res) => {
        bookmarks.add(...res.data);
        stats.set({
          total: res.total,
          moreLeft: $bookmarks.length < res.total,
          page: res.page,
        });
      });
    } else {
      $queryParams.page = 0;
      getBookmarks().then((res) => {
        bookmarks.set(res.data);
        stats.set({
          total: res.total,
          moreLeft: res.total > 20,
          page: res.page,
        });
      });
    }
  });

  // Child state
  let marked: number[] = [];
  const toggleMark = (bookmarkID: number) => {
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
    <div>
      <button class="w-full" on:click={() => $modals["add-bookmark"].showModal()}>
        + Add Bookmark
      </button>
      <hr />
      <span>Showing {$bookmarks.length} of a total of {$stats.total} bookmarks</span>
    </div>
    <div class="m-l-auto">
      {#if marked.length > 0}
        <BatchActionButton
          action="DELETE"
          title="Delete"
          icon={faTrash}
          handler={async () => {
            const res = await api("/bookmarks", "DELETE", { ids: marked });
            bookmarks.delete(...marked);
            toast.push(`Deleted ${res.deleted} bookmarks`);
          }}
        />
        <BatchActionButton action="TAGS" title="Add Tags" icon={faTags} />
        <BatchActionButton action="FOLDER" title="Move to Folder" icon={faFolderPlus} />
        <hr />
        <span>
          {marked.length} bookmark{marked.length > 1 ? "s" : ""} selected
        </span>
      {/if}
    </div>
  </div>
  <hr />
</div>
{#each $bookmarks as bookmark}
  <Bookmark {bookmark} {toggleMark} />
{/each}

{#if $stats.moreLeft}
  <div
    style="text-align: center;"
    on:click={() => {
      $queryParams.page++;
    }}
  >
    <button>Load More</button>
  </div>
{/if}
