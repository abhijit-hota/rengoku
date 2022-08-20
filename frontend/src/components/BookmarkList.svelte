<script lang="ts">
  import { faFolderPlus, faGhost, faTags, faTrash } from "@fortawesome/free-solid-svg-icons";
  import { toast } from "@zerodevx/svelte-toast";

  import { Bookmark, Loader } from "@components";
  import { api, store } from "@lib";
  import { modals } from "@Modal";
  import BatchActionButton from "./BatchActionButton.svelte";
  import Fa from "svelte-fa";

  const { bookmarks, queryParams, queryStr, stats } = store;

  // UI State
  let fetchingStatus: "LOADING" | "DONE" | "ERROR" = "LOADING";

  const getBookmarks = async () => {
    try {
      fetchingStatus = "LOADING";
      const res = await api("/bookmarks" + $queryStr);
      fetchingStatus = "DONE";
      return res;
    } catch (error) {
      // TODO
      fetchingStatus = "ERROR";
    }
  };

  let currentPage = 0;

  queryStr.subscribe(() => {
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
  export let marked: number[] = [];
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
      {#if $bookmarks.length > 0}
        <hr />
        <span>Showing {$bookmarks.length} of {$stats.total} bookmarks</span>
      {/if}
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
            marked = [];
            $stats = {
              page: 0,
              total: 0,
              moreLeft: false,
            };
            toast.push(`Deleted ${res.deleted} bookmarks`);
          }}
        />
        <BatchActionButton
          action="TAGS"
          title="Add Tags"
          icon={faTags}
          handler={async () => {
            $modals["add-tags"].showModal();
          }}
        />
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

{#if fetchingStatus === "LOADING"}
  <Loader />
{:else if fetchingStatus === "ERROR"}
  Error
{:else if $bookmarks.length > 0}
  {#each $bookmarks as bookmark}
    <Bookmark {bookmark} {toggleMark} checked={marked.includes(bookmark.id)} />
  {/each}
{:else}
  <div id="not-found">
    <Fa icon={faGhost} size="3x" class="bounce" />
    <div id="shadow" />
    <span class="message"> No bookmarks found </span>
  </div>
{/if}

{#if $stats.moreLeft}
  <div style="text-align: center;" on:click={() => $queryParams.page++}>
    <button>Load More</button>
  </div>
{/if}

<style>
  #not-found {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-top: 3em;
  }
  #not-found .message {
    margin-top: 1em;
    font-weight: bold;
    opacity: 0.7;
  }
  :global(.bounce) {
    animation: mover 1s infinite alternate;
  }
  #shadow {
    width: 30px;
    height: 10px;
    border-radius: 100%;
    background-color: #dbdbdb;
    opacity: 0.3;
    animation: expand 1s infinite alternate;
  }

  @keyframes mover {
    0% {
      transform: translateY(0);
    }
    100% {
      transform: translateY(-5px);
    }
  }
  @keyframes expand {
    0% {
      transform: scale(1);
    }
    100% {
      transform: scale(0.8);
    }
  }
</style>
