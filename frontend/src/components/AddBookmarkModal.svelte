<script lang="ts">
  import MultiSelect from "svelte-multiselect";
  import { bookmarks } from "../lib/stores";
  import { api } from "../lib";
  import Modal from "./Modal.svelte";

  let status: string;
  let message: string;
  let url: string;
  let selectedTags = [];
  let selectedFolders = [];

  const isNewTag = (tag) => tag.label == tag.value && typeof tag.value === "string";
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
        message = "An error occurred";
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
      bookmarks.add(newBookmark);
      // @ts-ignore
      document.getElementById("add-bookmark-modal").close();
    } catch (error) {
      status = "ERROR";
      message = "An error occurred";
    }
  };
</script>

<Modal key="add-bookmark" styles={{ dialog: "width: 600px;" }}>
  <h2 slot="header">Add new bookmark</h2>

  <div class="col" slot="body">
    <div class="col m-b-1">
      <label for="url"><strong>URL</strong><span class="red">*</span></label>
      <input type="url" bind:value={url} placeholder="Add new URL" name="url" id="url" required />
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
        <!-- {#if folders.length > 0} -->
        <MultiSelect
          inputClass="input-like"
          outerDivClass="color-fix"
          options={folders.map(({ id, name }) => ({ label: name, value: id }))}
          bind:selected={selectedFolders}
        />
        <!-- {/if} -->
      {/await}
    </div>

    {#if status === "ERROR"}
      <span class="m-b-1 red">{message}</span>
    {/if}
  </div>
  <button
    slot="footer"
    class="w-full"
    type="submit"
    disabled={status === "SUBMITTING"}
    on:click={addBookmark}
    style="margin: 1em 0;"
  >
    {status === "SUBMITTING" ? "Loading" : "Add"}
  </button>
</Modal>
