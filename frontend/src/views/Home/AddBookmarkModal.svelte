<script lang="ts">
  import MultiSelect from "svelte-multiselect";
  import type { ObjectOption } from "svelte-multiselect";
  import { api, store, errors } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faSpinner } from "@icons";
  import Fa from "svelte-fa";

  const { tags, folders } = store;
  const { flattened } = folders;

  let status: string;
  let message: string;
  let url: string;
  let selectedTags: ObjectOption[] = [];
  let selectedFolders: ObjectOption[] = [];

  const isNewTag = (tag: ObjectOption) => tag.label == tag.value && typeof tag.value === "string";

  const addBookmark = async () => {
    status = "SUBMITTING";

    selectedTags = selectedTags.map((tag) =>
      typeof tag === "string" ? { label: tag, value: tag } : tag
    );
    const newTags = selectedTags.filter(isNewTag);

    if (newTags.length > 0) {
      try {
        const res = await api("/tags/bulk", "POST", { names: newTags.map(({ value }) => value) });

        tags.add(...res);
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
      <MultiSelect
        inputClass="input-like"
        outerDivClass="color-fix"
        allowUserOptions
        addOptionMsg="+ Create new tag"
        options={$tags.map(({ id, name }) => ({ label: name, value: id }))}
        bind:selected={selectedTags}
      />
    </div>

    <div class="col m-b-1">
      <label for="folders"><strong>Folder</strong></label>
      <!-- TODO: folders tree input -->
      <MultiSelect
        inputClass="input-like"
        outerDivClass="color-fix"
        maxSelect={1}
        options={$flattened.map(({ id, label }) => ({ label, value: id }))}
        bind:selected={selectedFolders}
      />
      <!-- <FolderSelect bind:selectedFolderId={selectedFolder} /> -->
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
    on:keydown={(e) => {
      if (e.key === " " || e.key === "Enter") {
        e.preventDefault();
        e.stopPropagation();
        addBookmark();
      }
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
