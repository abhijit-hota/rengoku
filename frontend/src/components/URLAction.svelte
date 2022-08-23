<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import {
    faCheck,
    faMultiply,
    faPencil,
    faSpinner,
    faTrash,
  } from "@fortawesome/free-solid-svg-icons";

  import { api, store } from "@lib";
  import { toast } from "@zerodevx/svelte-toast";
  import Fa from "svelte-fa";

  import MultiSelect from "svelte-multiselect";
  import type { ObjectOption } from "svelte-multiselect";

  import type { URLAction } from "./SettingsPage.svelte";

  export let urlAction: URLAction;
  export let key: string | number;
  export let isNew = false;

  const dispatch = createEventDispatcher<{
    removeURLAction: { pattern: string };
  }>();

  const matchDetections = {
    starts_with: "starts with",
    regex: "matches the Regex",
    origin: "has the origin",
    domain: "belongs to the domain",
  };

  let { tags } = store;
  let selectedTags: ObjectOption[] = (urlAction.tags || []).map((value) => ({
    value,
    label: $tags.find(({ id }) => value === id).name,
  }));

  let isEditing = false;
  let isSaving = false;
  let isDeleting = false;
</script>

<div class="url-action">
  <div class="content">
    <div>
      If the URL
      {#if isNew}
        <select
          class="temp"
          id={"matchDetection" + key}
          name={"matchDetection" + key}
          bind:value={urlAction.matchDetection}
          readonly={!isNew}
          disabled={!isNew}
        >
          {#each Object.entries(matchDetections) as [k, v]}
            <option value={k}>{v}</option>
          {/each}
        </select>
      {:else}
        <span class="special">
          {matchDetections[urlAction.matchDetection]}
        </span>
      {/if}

      {#if isNew}
        <input
          class="temp"
          type="text"
          id={"pattern" + key}
          bind:value={urlAction.pattern}
          readonly={!isNew}
          disabled={!isNew}
        />
      {:else}
        <span class="special">
          {urlAction.pattern}
        </span>
      {/if}
    </div>

    <!-- TODO -->
    <!-- <div class="row">
    <div class="label">then add the link to the folders</div>
    <select
      class="temp"
      id={"matchDetection" + key}
      name={"matchDetection" + key}
      bind:value={urlAction.matchDetection}
    >
      <option value="" disabled>TODO</option>
    </select>
  </div> -->

    <div class="row" style="margin: 10px 0;">
      <div class="label">add the following tags to it&nbsp;</div>
      <MultiSelect
        disabled={!isEditing && !isNew}
        inputClass="input-like"
        placeholder="Search tags"
        removeAllTitle="Clear all tags"
        outerDivClass="color-fix filter-tags temp"
        options={$tags.map(({ id, name }) => ({ label: name, value: id }))}
        bind:selected={selectedTags}
      />
    </div>

    <div class="row">
      <div class="label">and save it offline</div>
      <input
        disabled={!isEditing && !isNew}
        class="temp"
        type="checkbox"
        id={"saveOffline" + key}
        name={"saveOffline" + key}
        bind:checked={urlAction.shouldSaveOffline}
      />
    </div>
  </div>
  {#if !isNew}
    <div class="actions">
      <button
        on:click={() => {
          isEditing = !isEditing;
        }}
      >
        {#if isEditing}
          <Fa icon={faMultiply} />
        {:else}
          <Fa icon={faPencil} />
        {/if}
      </button>
      {#if isEditing}
        <button
          on:click={async () => {
            try {
              isSaving = true;
              const res = await api("/config/url-action", "PUT", {
                ...urlAction,
                tags: selectedTags.map(({ value }) => value),
              });
              isEditing = false;
            } catch (error) {
              toast.push("An error occured while editing the URL action");
              console.error(error);
            } finally {
              isSaving = false;
            }
          }}
        >
          {#if isSaving}
            <Fa icon={faSpinner} spin />
          {:else}
            <Fa icon={faCheck} />
          {/if}</button
        >
      {:else}
        <button
          on:click={async () => {
            try {
              isDeleting = true;
              await api("/config/url-action", "DELETE", urlAction);
              dispatch("removeURLAction", { pattern: urlAction.pattern });
            } catch (error) {
              toast.push("An error occured while deleting the URL action");
              console.error(error);
            } finally {
              isDeleting = false;
            }
          }}
        >
          {#if isDeleting}
            <Fa icon={faSpinner} spin />
          {:else}
            <Fa icon={faTrash} />
          {/if}</button
        >
      {/if}
    </div>
  {/if}
</div>

<style>
  .url-action {
    padding: 1em 0;
    display: flex;
    justify-content: space-between;
  }
  .url-action .actions {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }
  .url-action:not(:first-child) {
    border-top: 1px solid #bbb;
  }
  .temp {
    display: inline-flex !important;
  }
  .label {
    min-width: 250px;
  }
  .special {
    font-weight: 500;
    background-color: var(--background);
    padding: 3px 6px;
    border-radius: 4px;
  }
</style>
