<script>
  import { api } from "../lib";

  import { queryParams } from "../lib/stores";
  import { onMount } from "svelte";
  let tags = [];
  onMount(async () => {
    try {
      tags = await api("/tags");
    } catch (error) {
      console.debug(error);
    }
  });
</script>

<input
  type="text"
  id="search"
  name="search"
  placeholder="Search"
  class="m-b-2"
  bind:value={$queryParams.search}
/>
<div class="row m-b-2">
  <div class="col">
    <strong>SORT BY</strong>
    <div>
      <input type="radio" name="sortBy" id="title" value="title" bind:group={$queryParams.sortBy} />
      <label for="title">Title</label>
    </div>
    <div>
      <input type="radio" name="sortBy" id="date" value="date" bind:group={$queryParams.sortBy} />
      <label for="date">Date</label>
    </div>
  </div>
  <div class="col m-l-auto">
    <strong>ORDER</strong>
    <div>
      <input type="radio" name="order" id="asc" value="asc" bind:group={$queryParams.order} />
      <label for="asc">Ascending</label>
    </div>
    <div>
      <input type="radio" name="order" id="desc" value="desc" bind:group={$queryParams.order} />
      <label for="desc">Descending</label>
    </div>
  </div>
</div>

<div class="row">
  <strong>FILTER BY TAGS</strong>
  <button
    class="m-l-auto"
    disabled={$queryParams.tags.length === 0}
    on:click={() => {
      $queryParams.tags = [];
    }}
  >
    Clear
  </button>
</div>
<div class="tags m-b-2">
  {#each tags as tag}
    <div
      class="tag"
      role="checkbox"
      style={`opacity: ${$queryParams.tags.includes(tag.id) ? 1 : 0.5}`}
      on:click={() => {
        const queryTags = $queryParams.tags;
        if (queryTags.includes(tag.id)) {
          const toDelete = queryTags.indexOf(tag);
          queryTags.splice(toDelete, 1);
        } else {
          queryTags.push(tag.id);
        }
        $queryParams.tags = queryTags;
      }}
    >
      {tag.name}
    </div>
  {/each}
</div>

<div>
  <button
    class="w-full"
    on:click={() => {
      // @ts-ignore
      document.getElementById("settings").showModal();
    }}>Open Settings</button
  >
</div>

<style>
  input[type="radio"], label {
    cursor: pointer;
  }
  .tags .tag {
    cursor: pointer;
    opacity: 0.5;
    transition: opacity 100ms ease-out;
  }

  .tags .tag:hover {
    opacity: 1 !important;
  }
</style>
