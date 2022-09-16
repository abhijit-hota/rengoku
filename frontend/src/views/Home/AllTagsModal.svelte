<script lang="ts">
  import { api, store } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faCheck, faPencil, faSpinner, faTrash } from "@icons";
  import Fa from "svelte-fa";

  const { tags } = store;

  let status: string;
  let message: string;
  let editing: store.Tag;
  let editingVal = "";
</script>

<Modal key="all-tags" styles={{ dialog: "width: 600px; height: 400px;" }}>
  <h2 slot="header">All tags</h2>

  <div slot="body">
    <div class="tags-container">
      {#each $tags as tag (tag.id)}
        <div class="tag">
          <div>
            {#if editing?.id == tag.id}
              <!-- svelte-ignore a11y-autofocus -->
              <input
                type="text"
                bind:value={editingVal}
                class="no-style"
                autofocus
                style="width: 100px;"
              />
            {:else}
              {tag.name}
            {/if}
          </div>
          {#if editing?.id == tag.id}
            <button
              class="no-style"
              on:click={async () => {
                try {
                  await api("/tags/" + tag.id, "PATCH", { name: editingVal });
                  tags.updateOne(tag.id, { ...tag, name: editingVal });
                  editing = undefined;
                  editingVal = "";
                } catch (error) {
                  console.error(error);
                }
              }}
              title="Edit folder name"
              style="margin-left: 0.75em; cursor: pointer;"
            >
              <Fa icon={faCheck} />
            </button>
          {:else}
            <button
              class="no-style"
              on:click={() => {
                editing = tag;
                editingVal = tag.name;
              }}
              title="Edit folder name"
              style="margin-left: 0.75em; cursor: pointer;"
            >
              <Fa icon={faPencil} />
            </button>
          {/if}

          <button
            class="no-style"
            on:click={async () => {
              try {
                await api("/tags/" + tag.id, "DELETE");
                tags.delete(tag.id);
              } catch (error) {
                console.error(error);
              }
            }}
            title="Delete folder"
            style="margin-left: 0.75em; cursor: pointer;"
          >
            <Fa icon={faTrash} />
          </button>
        </div>
      {/each}
    </div>
  </div>
  <div slot="footer" style="margin-top: auto;">
    {#if status === "ERROR"}
      <span class="m-b-1 red">{message}</span>
    {/if}
    <button
      class="w-full"
      disabled={status === "SUBMITTING"}
      on:click={() => $modals["all-tags"].close()}
      style="margin: 1em 0;"
    >
      {#if status === "SUBMITTING"}
        <Fa icon={faSpinner} size="lg" spin />
      {:else}
        Close
      {/if}
    </button>
  </div>
</Modal>

<style>
  .tags-container {
    display: flex;
    flex-wrap: wrap;
  }
  .tag {
    display: flex;
    padding: 0.5em;
    border-radius: 8px;
    background-color: var(--background);
    margin: 0 0.5em 0.5em 0;
  }
</style>
