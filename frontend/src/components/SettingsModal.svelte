<script lang="ts">
  import { onMount } from "svelte";
  
  import { api } from "@lib";
  import Modal from "@Modal";

  let config = {
    shouldSaveOffline: null,
    urlActions: [],
  };
  onMount(async () => {
    config = await api("/config");
  });
</script>

<Modal key="settings">
  <h2 slot="header">Settings</h2>
  <div slot="body" class="col" style="width: 600px; max-height: 700px;">
    <div class="row m-b-2">
      <h3>Save pages offline</h3>
      <input
        class="m-l-auto"
        type="checkbox"
        name="saveOfflineDefault"
        id="saveOfflineDefault"
        bind:checked={config.shouldSaveOffline}
      />
    </div>

    <h3>URL Actions</h3>
    <hr />
    {#each config.urlActions as urlAction, i}
      <div class="url-action m-b-2">
        <div class="row w-100">
          <div class="col" style="flex-grow: 3">
            <label for={"pattern" + i}>Pattern</label>
            <input
              type="text"
              id={"pattern" + i}
              bind:value={urlAction.pattern}
              readonly
              disabled
            />
          </div>

          <div class="col">
            <label for={"matchDetection" + i}>Match detection</label>
            <select
              id={"matchDetection" + i}
              name={"matchDetection" + i}
              bind:value={urlAction.matchDetection}
            >
              <option value="starts_with">Starts with (default)</option>
              <option value="regex">Regex</option>
              <option value="origin">Origin</option>
              <option value="domain">Domain</option>
            </select>
          </div>
        </div>
        <div class="row">
          <div class="row">
            <input
              type="checkbox"
              id={"saveOffline" + i}
              name={"saveOffline" + i}
              bind:value={urlAction.shouldSaveOffline}
            />
            <label for="'saveOffline' + i">Save offline</label>
          </div>
        </div>
        <hr />
      </div>
    {/each}
  </div>
</Modal>

<style>
  .url-action {
    background-color: var(--background-alt);
    border-radius: 6px;
  }
</style>
