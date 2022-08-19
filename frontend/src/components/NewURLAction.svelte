<script lang="ts">
  import UrlAction from "./URLAction.svelte";
  import Fa from "svelte-fa";
  import { faCheck, faMultiply, faSpinner } from "@fortawesome/free-solid-svg-icons";
  import type { URLAction } from "./SettingsPage.svelte";
  import { api } from "@lib";
  import { toast } from "@zerodevx/svelte-toast";
  import { createEventDispatcher } from "svelte";

  let urlAction: URLAction = {
    pattern: "",
    matchDetection: "",
    shouldSaveOffline: false,
    tags: [],
    folders: [],
  };

  let isSaving = false;
  let isNewBeingAdded = false;

  const dispatch = createEventDispatcher<{
    addURLAction: { urlAction: URLAction };
  }>();
</script>

{#if isNewBeingAdded}
  <UrlAction {urlAction} key={Math.random()} isNew />
  <div class="row" style="justify-content: flex-end;">
    <button
      style="margin-right: 1em;"
      on:click={() => {
        isNewBeingAdded = false;
      }}
    >
      <Fa icon={faMultiply} /> Discard
    </button>
    <button
      on:click={async () => {
        try {
          isSaving = true;
          const res = await api("/config/url-action", "POST", urlAction);
          isNewBeingAdded = false;
          dispatch("addURLAction", { urlAction });
        } catch (error) {
          toast.push("An error occured while editing the URL action");
          console.error(error, urlAction);
        } finally {
          isSaving = false;
        }
      }}
    >
      {#if isSaving}
        <Fa icon={faSpinner} spin />
      {:else}
        <Fa icon={faCheck} /> Done
      {/if}
    </button>
  </div>
{:else}
  <div class="row" style="justify-content: flex-end; border-top: 1px solid #bbb;">
    <button
      style="margin: 0; margin-top: 1em"
      on:click={() => {
        isNewBeingAdded = true;
      }}>+ Add a URL Action</button
    >
  </div>
{/if}
