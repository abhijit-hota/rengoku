<script>
  import { bookmarks } from "../lib/stores";
  import { api } from "../lib";
  import ellipsisIcon from "../assets/ellipsis.png";
  import OutClick from "./OutClick.svelte";
  import { identity } from "svelte/internal";

  /**@type {Bookmark}*/
  export let bookmark;
  export let toggleMark;

  let hovered = false;
  let checked = false;

  const deleteBookmark = async () => {
    try {
      const res = await api("/bookmarks/" + bookmark.id, "DELETE");
      bookmarks.delete(bookmark.id);
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

  let menuOpen = false;
  let focusedID = 0;
  /** @type {HTMLUListElement} */
  let ref = null;
  $: {
    if (ref) {
      const btn = [...ref.children].filter(({ tagName }) => tagName === "LI")[focusedID]
        .firstElementChild;
      // @ts-ignore
      btn.focus();
    }
  }
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
      <div class="popup">
        <button
          id={"menu-button-" + bookmark.id}
          class="icon-button"
          on:click={() => {
            menuOpen = !menuOpen;
          }}
        >
          <img src={ellipsisIcon} alt="Menu" style="filter: invert(1); transform: rotate(90deg);" />
        </button>
        <OutClick
          on:outclick={() => {
            menuOpen = false;
          }}
          excludeByQuerySelector={[
            "#menu-button-" + bookmark.id,
            "save-offline-button-" + bookmark.id,
          ]}
        >
          {#if menuOpen}
            <ul
              class="menu"
              on:keydown={(e) => {
                const { key } = e;
                if (!(key === "ArrowUp" || key === "ArrowDown")) {
                  return;
                }

                e.preventDefault();
                if (key === "ArrowDown") {
                  focusedID = (focusedID + 1) % 6;
                } else if (key === "ArrowUp") {
                  if (focusedID <= 0) {
                    focusedID = 5;
                  } else {
                    focusedID--;
                  }
                }
              }}
              bind:this={ref}
            >
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
                      bookmarks.updateOne(bookmark.id, { ...bookmark, meta: res });
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
            </ul>
          {/if}
        </OutClick>
      </div>
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
    <div class="footer row">
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
    margin-top: 0.5em;
  }

  .tag {
    background-color: var(--background);
    border-radius: 10px;
    margin: 4px;
    margin-left: 0px;
    padding: 4px 8px;
    width: max-content;
  }

  .popup .menu {
    list-style-type: none;

    background-color: var(--button-base);
    border-radius: 6px;

    padding: 0;
    position: absolute;
    margin-top: 0.5em;
    z-index: 1000;
  }
  .popup .menu hr {
    margin: 0;
  }
  .popup .menu li > * {
    padding: 0.75em;
    cursor: pointer;
    width: calc(100% - 0.75em * 2);
  }
  .popup .menu li > *:hover,
  .popup .menu li > *:focus {
    background-color: var(--button-hover);
  }
  .popup .menu li:last-child > * {
    border-radius: 0 0 6px 6px;
  }
  .popup .menu li:first-child > * {
    border-radius: 6px 6px 0 0;
  }
</style>
