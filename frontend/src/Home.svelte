<script lang="ts">
  import "./styles/water.min.css";
  import "./styles/style.css";

  import { SvelteToast } from "@zerodevx/svelte-toast";
  import { router } from "tinro";

  import { modals } from "@Modal";
  import { Filter, BookmarkList, AddBookmarkModal, ImportModal } from "@components";
  import FolderWrapper from "./components/FolderWrapper.svelte";
  import { auth, queryParams } from "./lib/stores";
  import { faGear, faPowerOff } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";
  import AllTagsModal from "./components/AllTagsModal.svelte";
</script>

<nav>
  <div class="row">
    <h1>Rengoku</h1>
    <button style="margin-left: auto;" on:click={() => $modals["import"].showModal()}
      >Import Bookmarks</button
    >
    <button
      on:click={() => {
        // Hate myself for not just using an anchor tag but have to do it for the styling
        router.goto("/settings");
      }}
    >
      <Fa icon={faGear} /> Settings
    </button>
    <button on:click={auth.logout}>
      <Fa icon={faPowerOff} /> Log Out
    </button>
  </div>
  <hr />
</nav>

<aside id="folders" class="sticky">
  <div class="row" style="min-height: 39px;">
    <h2>Folders</h2>
    {#if $queryParams.folder !== ""}
      <button
        class="m-l-auto"
        style="margin: 0;"
        on:click={() => {
          $queryParams.folder = "";
        }}>Show all</button
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
<AllTagsModal />
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
    z-index: 1;
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
