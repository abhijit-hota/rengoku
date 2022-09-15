<script lang="ts">
  import MultiSelect from "svelte-multiselect";
  import type { ObjectOption } from "svelte-multiselect";
  import { api, store, errors } from "@lib";
  import Modal, { modals } from "@Modal";
  import { faSpinner } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import type { Bookmark } from "../lib/stores";

  const { tags, bookmarks, folders } = store;
  const { flattened } = folders;
  let status: string;
  let message: string;
  let selectedTags: ObjectOption[] = [];
  let selectedFolders: ObjectOption[] = [];

  export let activeBookmarkId: number;
  let bookmark: Bookmark;

  let updateState: {
    title?: string;
    description?: string;
    tag_ids?: number[];
    folder_ids?: number[];
  } = {
    title: "",
    description: "",
    tag_ids: [],
    folder_ids: [],
  };

  const setBookmarkAndTags = () => {
    bookmark = $bookmarks.find(({ id }) => id === activeBookmarkId);
    selectedTags = bookmark?.tags.map(({ id, name }) => ({ label: name, value: id }));
    selectedFolders = bookmark?.folders.map((id) => ({
      label: $flattened.find(({ id: _id }) => id === _id)?.label || "",
      value: id,
    }));

    updateState = {
      title: bookmark.meta.title,
      description: bookmark.meta.description,
      tag_ids: bookmark.tags.map(({ id }) => id),
      folder_ids: bookmark.folders,
    };
  };

  $: if (activeBookmarkId) setBookmarkAndTags();

  const isNewTag = (tag: ObjectOption) => tag.label == tag.value && typeof tag.value === "string";
  const updateBookmark = async () => {
    status = "SUBMITTING";

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
        status = "";
        return;
      }
    }

    const dataToPost = {
      title: updateState.title,
      description: updateState.description,
      tag_ids: selectedTags.map(({ value }) => value),
      folder_ids: selectedFolders.map(({ value }) => value),
    };
    try {
      const updatedBookmarkRes = await api("/bookmarks/" + bookmark.id, "PATCH", dataToPost);
      status = "SUCCESS";
      store.bookmarks.updateOne(bookmark.id, {
        ...bookmark,
        meta: {
          ...bookmark.meta,
          title: updatedBookmarkRes.title,
          description: updatedBookmarkRes.description,
        },
        last_updated: updatedBookmarkRes.last_updated,
        tags: $tags.filter(({ id }) => updatedBookmarkRes.tag_ids.includes(id)),
      });
      $modals["edit-bookmark"].close();
    } catch (error) {
      // TODO
      status = "ERROR";
      if (error.cause === errors.NAME_ALREADY_PRESENT) {
        message = "The same URL is already present.";
      }
    }
  };
</script>

<Modal key="edit-bookmark" styles={{ dialog: "width: 600px;" }}>
  <h2 slot="header">Edit bookmark</h2>

  <div class="col" slot="body">
    <div class="col m-b-1">
      <label for="title"><strong>Title</strong></label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        autofocus
        type="text"
        bind:value={updateState.title}
        placeholder="Title"
        name="title"
        id="title"
        required
      />
    </div>

    <div class="col m-b-1">
      <label for="description"><strong>Description</strong></label>
      <textarea
        bind:value={updateState.description}
        placeholder="Description"
        name="description"
        id="description"
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
      <MultiSelect
        inputClass="input-like"
        outerDivClass="color-fix"
        maxSelect={1}
        options={$flattened.map(({ id, label }) => ({ label, value: id }))}
        bind:selected={selectedFolders}
      />
    </div>

    {#if status === "ERROR"}
      <span class="m-b-1 red">{message}</span>
    {/if}
  </div>
  <button
    slot="footer"
    class="w-full"
    disabled={status === "SUBMITTING"}
    on:click={updateBookmark}
    on:keydown={(e) => {
      if (e.key === " " || e.key === "Enter") {
        e.preventDefault();
        e.stopPropagation();
        updateBookmark();
      }
    }}
    style="margin: 1em 0;"
  >
    {#if status === "SUBMITTING"}
      <Fa icon={faSpinner} size="lg" spin />
    {:else}
      Save
    {/if}
  </button>
</Modal>

<style>
  :global(ul.options) {
    max-height: 150px;
  }
</style>
