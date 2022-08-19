<script lang="ts">
  import { api, store } from "@lib";
  const { queryParams, tags } = store;

  import MultiSelect from "svelte-multiselect";
  let selectedTags = [];
  $: $queryParams.tags = selectedTags.map(({ value }) => value);
</script>

<input
  type="text"
  id="search"
  name="search"
  placeholder="Search"
  class="m-b-2 w-full"
  style="box-sizing: border-box;"
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

<div class="row m-b-1">
  <strong>FILTER BY TAGS</strong>
  <button
    class="m-l-auto"
    style="margin-right: 0;"
    disabled={$queryParams.tags.length === 0}
    on:click={() => {
      selectedTags = [];
      $queryParams.tags = [];
    }}
  >
    Clear
  </button>
</div>
<MultiSelect
  inputClass="input-like"
  placeholder="Search tags"
  removeAllTitle="Clear all tags"
  outerDivClass="color-fix filter-tags"
  options={$tags.map(({ id, name }) => ({ label: name, value: id }))}
  bind:selected={selectedTags}
/>

<style>
  input[type="radio"],
  label {
    cursor: pointer;
  }
</style>
