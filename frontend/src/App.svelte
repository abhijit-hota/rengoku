<script>
  import "./styles/water.min.css";
  import "./styles/style.css";
  import { api } from "./lib";

  import { Filter, SettingsModal, AddBookmarkModal, Folder, BookmarkList } from "./components";
</script>

<nav>
  <div class="row">
    <h1>Rengoku</h1>
    <button
      class="m-l-auto"
      on:click={() => {
        // @ts-ignore
        document.getElementById("add-bookmark-modal").showModal();
      }}
    >
      + Add Bookmark
    </button>
  </div>
  <hr />
</nav>
<aside id="folders" class="sticky">
  <h2>Folders</h2>
  <hr />
  {#await api("/folders/tree") then folders}
    {#each folders as tree}
      <Folder {tree} />
    {/each}
  {/await}
</aside>
<main>
  <BookmarkList />
</main>
<aside id="filters" class="sticky">
  <Filter />
</aside>
<AddBookmarkModal />
<SettingsModal />

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
