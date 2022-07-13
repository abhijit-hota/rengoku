<script lang="ts">
  import { store, api } from "@lib";
  import Fa from "svelte-fa";
  import {
    faAdd,
    faCheck,
    faChevronRight,
    faEdit,
    faMultiply,
    faTrash,
  } from "@fortawesome/free-solid-svg-icons";
  import { isParent, Tree, isExpanded } from "./FolderWrapper.svelte";
  const { queryParams } = store;

  export let tree: Tree = [];
  export let parentId: number = undefined;

  let newFolder = "";
  let updatedFolderName: Record<number, string> = {};

  tree.forEach(({ id, children }) => {
    $isParent[id] = (children?.length ?? 0) > 0;
  });
  tree.sort(({ children }) => (children ? -1 : 1));
</script>

<ul>
  {#each tree as node}
    {@const { id, name, children } = node}
    {@const expanded = $isExpanded[id] ?? true}
    <li
      class:has-children={$isParent[id]}
      class:active-folder={$queryParams.folder === id.toString()}
    >
      {#if updatedFolderName[id] === undefined}
        <div class="title">
          {#if $isParent[id]}
            <button
              class="no-style"
              on:click={() => {
                $isExpanded[id] = !expanded;
              }}
            >
              <Fa
                icon={faChevronRight}
                size="xs"
                style="margin-right: 0.5rem; margin-left: -0.2em;"
                color="var(--border)"
                rotate={expanded ? 90 : 0}
              />
              {name}
            </button>
          {:else}
            <button
              class="no-style"
              on:click={() => {
                $queryParams.folder = id.toString();
              }}>{name}</button
            >
          {/if}
          <div class="actions">
            <button
              class="no-style"
              on:click={() => (updatedFolderName[id] = name)}
              title="Edit folder name"
            >
              <Fa icon={faEdit} />
            </button>
            <button
              class="no-style"
              on:click={async () => {
                try {
                  await api("/folders/" + id.toString(), "DELETE");
                  if (tree?.length === 1) $isParent[parentId] = false;
                  tree.splice(
                    tree.findIndex((n) => n.id === id),
                    1
                  );
                  tree = tree;
                } catch (error) {}
              }}
              title="Delete folder"
            >
              <Fa icon={faTrash} />
            </button>
            {#if !$isParent[id]}
              <button
                class="no-style"
                on:click={() => {
                  node.children = [];
                  $isParent[id] = true;
                }}
                title="Add Subfolder"
              >
                <Fa icon={faAdd} />
              </button>
            {/if}
          </div>
        </div>
      {:else}
        <form
          class="no-style"
          on:submit|preventDefault={async () => {
            try {
              const res = await api("/folders/" + id.toString(), "PATCH", {
                name: updatedFolderName[id],
              });
              node.name = res.name;
              updatedFolderName[id] = undefined;
            } catch (error) {
              console.error(error);
            }
          }}
        >
          <!-- svelte-ignore a11y-autofocus -->
          <input
            type="text"
            class="no-style"
            bind:value={updatedFolderName[id]}
            autofocus={updatedFolderName[id] !== undefined}
          />
          <button class="no-style" type="submit">
            <Fa icon={faCheck} />
          </button>
          <button
            class="no-style"
            type="reset"
            on:click={() => (updatedFolderName[id] = undefined)}
          >
            <Fa icon={faMultiply} />
          </button>
        </form>
      {/if}

      {#if $isParent[id] && expanded}
        <svelte:self tree={children} parentId={id} />
      {/if}
    </li>
  {/each}
  <li>
    <form
      on:submit|preventDefault={async () => {
        try {
          const res = await api("/folders", "POST", {
            name: newFolder,
            parent_id: parentId,
          });
          tree.push(res);
          tree = tree;
          newFolder = "";
        } catch (error) {}
      }}
    >
      <input type="text" class="no-style" placeholder="+ Add folder" bind:value={newFolder} />
    </form>
  </li>
</ul>

<style>
  ul {
    list-style: none;
    user-select: none;
    margin: 0;
    padding-left: 0;
    margin-bottom: 1.5em;
    border-left: 1px solid var(--border);
  }
  li {
    padding: 0.5rem 0 0.5rem 0.8rem !important;
    border-bottom: 1px dashed var(--border);
  }
  li .title,
  li form {
    display: flex;
  }
  li button,
  li form {
    flex: auto;
  }
  li form input {
    width: 100% !important;
  }
  form button {
    margin-left: 0.25rem;
    margin-right: 0.5rem;
  }

  .title:hover > .actions,
  .title > button:focus ~ .actions,
  .title:focus > .actions,
  .actions:focus {
    display: flex;
  }

  li:not(.has-children):hover {
    background: var(--background-alt);
  }
  li.active-folder {
    background: var(--background);
  }

  .actions {
    display: none;
  }
  .actions button {
    opacity: 0.5;
    margin-left: 0.2rem;
    margin-right: 0.4rem;
  }
  .actions button:hover {
    opacity: 1;
  }
  button {
    cursor: pointer;
  }
</style>
