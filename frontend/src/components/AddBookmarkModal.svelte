<script lang="ts">
  import MultiSelect, { ObjectOption } from "svelte-multiselect";
  import { api, store, errors } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faSpinner } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  let status: string;
  let message: string;
  let url: string;
  let selectedTags: ObjectOption[] = [];
  let selectedFolders = [];

  const isNewTag = (tag: ObjectOption) => tag.label == tag.value && typeof tag.value === "string";
  const addBookmark = async () => {
    status = "SUBMITTING";

    const newTags = selectedTags.filter(isNewTag);

    if (newTags.length > 0) {
      try {
        const res = await api("/tags/bulk", "POST", { names: newTags.map(({ value }) => value) });
        selectedTags = [
          ...selectedTags.filter((tag) => !isNewTag(tag)),
          ...res.map(({ id, name }) => ({ label: name, value: id })),
        ];
      } catch (error) {
        status = "ERROR";
        if (error.cause === errors.NAME_ALREADY_PRESENT) {
          message = "A tag with the same name is already present.";
        }
        return;
      }
    }

    const dataToPost = {
      url,
      tags: selectedTags.map(({ value }) => value),
      folders: selectedFolders.map(({ value }) => value),
    };
    try {
      const newBookmark = await api("/bookmarks", "POST", dataToPost);
      status = "SUCCESS";
      store.bookmarks.add(newBookmark);
      store.stats.update((val) => ({ ...val, total: val.total + 1 }));
      $modals["add-bookmark"].close();
    } catch (error) {
      status = "ERROR";
      if (error.cause === errors.NAME_ALREADY_PRESENT) {
        message = "The same URL is already present.";
      }
    }
  };
</script>

<Modal key="add-bookmark" styles={{ dialog: "width: 600px;" }}>
  <h2 slot="header">Add new bookmark</h2>

  <div class="col" slot="body">
    <div class="col m-b-1">
      <label for="url"><strong>URL</strong><span class="red">*</span></label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        autofocus
        type="url"
        bind:value={url}
        placeholder="Add new URL"
        name="url"
        id="url"
        required
      />
    </div>

    <div class="col m-b-1">
      <label for="tags"><strong>Tags</strong></label>
      {#await api("/tags") then tags}
        <MultiSelect
          inputClass="input-like"
          outerDivClass="color-fix"
          allowUserOptions
          addOptionMsg="+ Create new tag"
          options={tags.map(({ id, name }) => ({ label: name, value: id }))}
          bind:selected={selectedTags}
        />
      {/await}
    </div>

    <div class="col m-b-1">
      <label for="folders"><strong>Folder</strong></label>
      {#await api("/folders") then folders}
        {#if folders.length > 0}
          <MultiSelect
            inputClass="input-like"
            outerDivClass="color-fix"
            options={folders.map(({ id, name }) => ({ label: name, value: id }))}
            bind:selected={selectedFolders}
          />
        {:else}
          No folders found. Please create from sidebar.
        {/if}
      {/await}
    </div>

    {#if status === "ERROR"}
      <span class="m-b-1 red">{message}</span>
    {/if}
  </div>
  <button
    slot="footer"
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
</Modal>

<style>
  :global(ul.options) {
    max-height: 150px;
  }
</style>
