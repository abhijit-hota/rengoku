<script lang="ts">
  import MultiSelect from "svelte-multiselect";
  import type { ObjectOption } from "svelte-multiselect";

  import { api, store, errors } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faSpinner } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  const { tags } = store;

  let status: string;
  let message: string;
  let selectedTags: ObjectOption[] = [];
  export let markedBookmarks: number[] = [];

  const isNewTag = (tag: ObjectOption) => tag.label == tag.value && typeof tag.value === "string";
  const addBookmark = async () => {
    status = "SUBMITTING";

    const newTagsFromMultiSelect = selectedTags.filter(isNewTag);

    if (newTagsFromMultiSelect.length > 0) {
      try {
        const res = await api("/tags/bulk", "POST", {
          names: newTagsFromMultiSelect.map(({ value }) => value),
        });

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
    const newTagIds = selectedTags.map(({ value }) => value);
    const dataToPost = {
      link_ids: markedBookmarks,
      tag_ids: newTagIds,
    };
    try {
      await api("/bookmarks/tags", "PUT", dataToPost);
      status = "SUCCESS";
      store.bookmarks.update((val) =>
        val.map((v) => {
          if (markedBookmarks.includes(v.id)) {
            return {
              ...v,
              tags: $tags.filter(
                (tag) => newTagIds.includes(tag.id) || v.tags.find(({ id }) => id === tag.id)
              ),
            };
          }
          return v;
        })
      );
      status = "";
      message = "";
      selectedTags = [];
      $modals["add-tags"].close();
    } catch (error) {
      status = "ERROR";
    }
  };
</script>

<Modal key="add-tags" styles={{ dialog: "width: 600px; height: 400px;" }}>
  <h2 slot="header">Add tags</h2>

  <div slot="body">
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
