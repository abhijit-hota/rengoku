<script lang="ts">
  import { api } from "@lib";
  import Modal from "@Modal";

  const handleImport: svelteHTML.FormEventHandler<HTMLFormElement> = (e) => {
    const formElem = e.currentTarget;
    try {
      api("/bookmarks/import", "POST", new FormData(formElem));
    } catch (error) {
      console.debug(error);
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
    <button class="w-full" type="submit">Import</button>
  </form>
</Modal>
