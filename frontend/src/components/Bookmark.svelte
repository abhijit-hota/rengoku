<script lang="ts">
  import { api, store } from "@lib";
  import Popup from "./Popup.svelte";

  export let bookmark: store.Bookmark;
  export let toggleMark: Function;

  let hovered = false;
  let checked = false;

  const deleteBookmark = async () => {
    try {
      const res = await api("/bookmarks/" + bookmark.id, "DELETE");
      store.bookmarks.delete(bookmark.id);
    } catch (error) {
      console.debug(error);
    }
  };

  const activate = () => {
    hovered = true;
  };
  const deActivate = () => {
    if (!checked) {
      hovered = false;
    }
  };
</script>

<div
  class={`bookmark ${checked || hovered ? "active" : ""}`}
  on:mouseover={activate}
  on:focus={activate}
  on:mouseout={deActivate}
  on:blur={deActivate}
>
  <div class="row header-row">
    {#if bookmark.meta.favicon}
      <img class="favicon" src={bookmark.meta.favicon} alt="🔥" />
    {/if}
    <h3 style="text-overflow: ellipsis;">{bookmark.meta.title}</h3>
    <div class="m-l-auto row">
      <Popup id={bookmark.id.toString()} excludeClicks={["#save-offline-button-" + bookmark.id]}>
        <svelte:fragment slot="list-items">
          <li role="none">
            <button class="no-style">Edit tags</button>
          </li>
          <li role="none">
            <button class="no-style">Edit folders</button>
          </li>
          <hr />
          <li role="none">
            <button
              id={"save-offline-button-" + bookmark.id}
              class="no-style"
              on:click={() => {
                api("/bookmarks/" + bookmark.id + "/save", "PUT");
              }}>Save Offline</button
            >
          </li>
          {#if bookmark.last_saved_offline}
            <li role="none">
              <a
                class="no-style"
                target="_blank"
                href={"http://localhost:8080/saved/" + bookmark.id}>Open saved copy</a
              >
            </li>
          {/if}
          <li role="none">
            <button
              class="no-style"
              on:click={async () => {
                try {
                  const /** @type {Bookmark["meta"]}*/ res = await api(
                      "/bookmarks/" + bookmark.id + "/meta",
                      "PUT"
                    );
                  store.bookmarks.updateOne(bookmark.id, { ...bookmark, meta: res });
                } catch (error) {
                  console.error(error);
                }
              }}>Refetch Metadata</button
            >
          </li>
          <hr />
          <li role="none">
            <button class="no-style red" on:click={deleteBookmark}>Delete</button>
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

  .bookmark.active {
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
