<script lang="ts">
  import "./styles/water.min.css";
  import "./styles/style.css";

  import { SvelteToast } from "@zerodevx/svelte-toast";

  import { modals } from "@Modal";
  import { Filter, SettingsModal, BookmarkList, AddBookmarkModal, ImportModal } from "@components";
  import FolderWrapper from "./components/FolderWrapper.svelte";
  import { queryParams } from "./lib/stores";
</script>

<nav>
  <div class="row">
    <h1>Rengoku</h1>
    <button class="m-l-auto" on:click={() => $modals["settings"].showModal()}>Open Settings</button>
    <button on:click={() => $modals["import"].showModal()}>Import Bookmarks</button>
  </div>
  <hr />
</nav>
<aside id="folders" class="sticky">
  <div class="row m-b-1">
    <h2>Folders</h2>
    {#if $queryParams.folder !== ""}
      <button
        class="m-l-auto"
        style="margin-right: 0;"
        on:click={() => {
          $queryParams.folder = "";
        }}>Show root</button
      >
    {/if}
  </div>
  <div style="overflow: auto; max-height: 70vh;">
    <hr style="margin-bottom: 0 !important;" />
    <FolderWrapper />
  </div>
</aside>
<main>
  <BookmarkList />
</main>
<aside id="filters" class="sticky">
  <Filter />
</aside>
<SettingsModal />
<AddBookmarkModal />
<ImportModal />
<SvelteToast options={{}} />

<style>
  nav {
    grid-area: nav;
    padding-top: 20px;
    padding-bottom: 1em;
    position: sticky;
    top: 0px;
    background-color: var(--background-body);
  }

  #filters {
    grid-area: filters;
    width: 15vw;
    min-width: 300px;
  }

  main {
    padding-right: 1em;
    width: 35vw;
  }

  #folders {
    grid-area: folders;
    width: 15vw;
  }
</style>
