<script lang="ts" context="module">
  // TODO: replace time-ago with custom script. It adds 40kB worth of JS!!!
  import TimeAgo from "javascript-time-ago";
  import en from "javascript-time-ago/locale/en";

  TimeAgo.addDefaultLocale(en);
  const timeAgo = new TimeAgo("en-US");
</script>

<script lang="ts">
  import { faClockRotateLeft, faSpinner } from "@fortawesome/free-solid-svg-icons";
  import { toast } from "@zerodevx/svelte-toast";
  import Fa from "svelte-fa";

  import { api, store } from "@lib";
  const { stats } = store;

  import Popup from "./Popup.svelte";
  import { modals } from "@Modal";

  export let bookmark: store.Bookmark;
  export let toggleMark: Function;
  export let activeBookmark: number;

  export let checked = false;

  let isDeleting = false,
    isRefetching = false,
    isSaving = false;

  const deleteBookmark = async () => {
    try {
      isDeleting = true;
      await api("/bookmarks/" + bookmark.id, "DELETE");
      store.bookmarks.delete(bookmark.id);
      $stats.total--;
    } catch (error) {
      console.debug(error);
      // TODO: Toast
    } finally {
      isDeleting = false;
    }
  };
</script>

<div id={bookmark.id.toString()} class="bookmark" class:checked>
  <div class="row header-row">
    {#if bookmark.meta.favicon}
      <img class="favicon" src={bookmark.meta.favicon} alt="🔥" />
    {/if}
    <h3 style="text-overflow: ellipsis;">{bookmark.meta.title}</h3>
    <div class="m-l-auto row">
      <Popup id={bookmark.id.toString()} excludeClicks={["#save-offline-button-" + bookmark.id]}>
        <svelte:fragment slot="list-items">
          <li role="none">
            <button
              class="no-style"
              on:click={() => {
                activeBookmark = bookmark.id;
                $modals["edit-bookmark"].showModal();
              }}>Edit</button
            >
          </li>
          <hr />
          <li role="none">
            <button
              id={"save-offline-button-" + bookmark.id}
              class="no-style"
              disabled={isSaving}
              style:cursor={isSaving ? "not-allowed" : "pointer"}
              on:click={async () => {
                try {
                  isSaving = true;
                  const { saved, lastSavedOffline } = await api(
                    "/bookmarks/" + bookmark.id + "/save",
                    "PUT"
                  );
                  if (saved) {
                    bookmark.last_saved_offline = lastSavedOffline;
                    bookmark = bookmark;
                  }
                } catch (error) {
                  console.debug(error);
                  toast.push("An error occurred while saving the page.");
                } finally {
                  isSaving = false;
                }
              }}
            >
              {#if isSaving}
                <Fa icon={faSpinner} spin color="#888" />
              {/if}
              <span style:color={isSaving ? "#888" : "inherit"}> Save Offline </span>
            </button>
          </li>
          {#if bookmark.last_saved_offline}
            <li role="none">
              <button
                class="no-style"
                on:click={(e) => {
                  e.preventDefault();
                  window.open(
                    "http://localhost:8080/saved/" + bookmark.id + "_" + bookmark.last_saved_offline
                  );
                }}
              >
                Open saved copy
                <br />
                <small style="background: unset !important; color: #bbb;">
                  <Fa icon={faClockRotateLeft} /> Last saved
                  {timeAgo.format(new Date(bookmark.last_saved_offline * 1000))}
                </small>
              </button>
            </li>
          {/if}
          <li role="none">
            <button
              class="no-style"
              disabled={isRefetching}
              style:cursor={isRefetching ? "not-allowed" : "pointer"}
              on:click={async () => {
                try {
                  isRefetching = true;
                  const /** @type {Bookmark["meta"]}*/ res = await api(
                      "/bookmarks/" + bookmark.id + "/meta",
                      "PUT"
                    );
                  store.bookmarks.updateOne(bookmark.id, { ...bookmark, meta: res });
                } catch (error) {
                  console.error(error);
                } finally {
                  isRefetching = false;
                }
              }}
            >
              {#if isRefetching}
                <Fa icon={faSpinner} spin color="#888" />
              {/if}
              <span style:color={isRefetching ? "#888" : "inherit"}> Refetch Metadata </span>
            </button>
          </li>
          <hr />
          <li role="none">
            <button
              class="no-style"
              on:click={deleteBookmark}
              disabled={isDeleting}
              style:cursor={isDeleting ? "not-allowed" : "pointer"}
            >
              {#if isDeleting}
                <Fa icon={faSpinner} spin color="#888" />
              {/if}
              <span style:color={isDeleting ? "#888" : "rgb(231, 126, 126)"}> Delete </span>
            </button>
          </li>
        </svelte:fragment>
      </Popup>
      <input
        type="checkbox"
        name={bookmark.id.toString()}
        id={bookmark.id.toString()}
        class="bookmark-checkbox"
        bind:checked
        on:change={() => toggleMark(bookmark.id)}
      />
    </div>
  </div>
  <small><a href={bookmark.url}>{bookmark.url}</a> </small>
  {#if bookmark.meta.description}
    <p>{bookmark.meta.description}</p>
  {/if}
  {#if bookmark.tags.length > 0}
    <div id="footer" class="row">
      <div class="tags">
        {#each bookmark.tags as tag}<div class="tag">{tag.name}</div>{/each}
      </div>
    </div>
  {/if}
</div>

<style>
  :root {
    --size: 1.8em;
  }
  .bookmark {
    display: flex;
    flex-direction: column;
    padding: 0.5em;

    margin-bottom: 1em;
    border-radius: 6px;
  }

  .bookmark.checked,
  .bookmark:hover {
    background-color: var(--background-alt);
  }

  .bookmark h3,
  .bookmark small {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .favicon {
    aspect-ratio: 1/1;
    width: 1.17em;
    margin-right: 0.5em;
  }
  .bookmark-checkbox {
    margin: 0 0 0 0.5em !important;
    width: calc(var(--size) - 2px);
    height: calc(var(--size) - 2px);
    border-radius: 6px !important;
  }

  .tags {
    display: flex;
    flex-wrap: nowrap;
    overflow-x: scroll;
  }

  .tag {
    background-color: var(--background);
    border-radius: 10px;
    margin: 4px;
    margin-left: 0px;
    padding: 4px 8px;
    width: max-content;
  }
</style>
