<script lang="ts">
  import { faSpinner } from "@fortawesome/free-solid-svg-icons";

  import { api, store } from "@lib";
  const { bookmarks, tags, folders } = store;

  import Modal, { modals } from "@Modal";
  import { toast } from "@zerodevx/svelte-toast";
  import Fa from "svelte-fa";

  let status: "SUBMITTING" | "ERROR" | "" | "SUCCESS" = "";
  let message = "";
  const handleImport: svelteHTML.FormEventHandler<HTMLFormElement> = async (e) => {
    const start = Date.now();
    status = "SUBMITTING";
    message = "";

    const formElem = e.currentTarget;
    try {
      const res = await api("/bookmarks/import", "POST", new FormData(formElem));
      bookmarks.init();
      tags.init();
      folders.init();
      status = "SUCCESS";
      const end = Date.now();
      const duration = ((end - start) / 1000).toPrecision(2);
      toast.push(`${res.length} bookmarks imported in ${duration} seconds.`);
      $modals["import"].close();
    } catch (error) {
      console.debug(error);
      status = "ERROR";
      message = "An error occurred.";
      // TODO: more comprehensive
    }
  };
</script>

<Modal key="import">
  <h2 slot="header">Import bookmarks</h2>
  <form on:submit|preventDefault={handleImport} slot="body">
    <input type="file" name="export" required />
    <br />
    <div class="m-b-1">
      <input type="checkbox" name="fetchMeta" id="fetchMeta" />
      <label for="fetchMeta">Fetch fresh metadata (slower)</label>
    </div>

    {#if message !== ""}
      <span class="red">
        {message}
      </span>
    {/if}

    <button class="w-full" type="submit" disabled={status === "SUBMITTING"}>
      {#if status === "SUBMITTING"}
        <Fa icon={faSpinner} size="lg" spin />
      {:else}
        Import
      {/if}
    </button>
  </form>
</Modal>
