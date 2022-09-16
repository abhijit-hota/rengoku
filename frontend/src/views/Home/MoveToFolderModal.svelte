<script lang="ts">
  import MultiSelect from "svelte-multiselect";
  import type { ObjectOption } from "svelte-multiselect";

  import { api, store } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faSpinner } from "@icons";
  import Fa from "svelte-fa";

  const { folders } = store;
  const { flattened } = folders;

  let status: string;
  let message: string;
  let selectedFolders: ObjectOption[] = [];
  export let markedBookmarks: number[] = [];

  const addBookmark = async () => {
    status = "SUBMITTING";

    const dataToPost = {
      link_ids: markedBookmarks,
      folder_ids: selectedFolders.map(({ value }) => value),
    };
    try {
      await api("/bookmarks/folders", "PUT", dataToPost);
      status = "SUCCESS";
      store.bookmarks.update((val) =>
        val.map((v) => {
          if (markedBookmarks.includes(v.id)) {
            return {
              ...v,
              folders: dataToPost.folder_ids as number[],
            };
          }
          return v;
        })
      );
      status = "";
      message = "";
      selectedFolders = [];
      $modals["add-folders"].close();
    } catch (error) {
      status = "ERROR";
    }
  };
</script>

<Modal key="add-folders" styles={{ dialog: "width: 600px; height: 400px;" }}>
  <h2 slot="header">Assign to folders</h2>

  <div slot="body">
    <div class="col m-b-1">
      <label for="tags"><strong>Folders</strong></label>
      <MultiSelect
        inputClass="input-like"
        outerDivClass="color-fix"
        maxSelect={1}
        options={$flattened.map(({ id, label }) => ({ label, value: id }))}
        bind:selected={selectedFolders}
      />
    </div>
  </div>
  <div slot="footer" style="margin-top: auto;">
    {#if status === "ERROR"}
      <span class="m-b-1 red">{message}</span>
    {/if}
    <button
      class="w-full"
      disabled={status === "SUBMITTING"}
      on:click={addBookmark}
      on:keydown|preventDefault|stopPropagation={(e) => {
        if (e.key === " " || e.key === "Enter") addBookmark();
      }}
      style="margin: 1em 0;"
    >
      {#if status === "SUBMITTING"}
        <Fa icon={faSpinner} size="lg" spin />
      {:else}
        Add
      {/if}
    </button>
  </div>
</Modal>

<style>
  :global(ul.options) {
    max-height: 150px;
  }
</style>
