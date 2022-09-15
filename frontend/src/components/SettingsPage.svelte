<script context="module" lang="ts">
  export type URLAction = {
    pattern: string;
    matchDetection: string;
    shouldSaveOffline: boolean;
    tags: number[];
    folders: number[];
  };
  export type Config = {
    shouldDeleteOffline: boolean;
    shouldSaveOffline: boolean;
    urlActions: URLAction[];
  };
</script>

<script lang="ts">
  import { onMount } from "svelte";

  import { api } from "@lib";
  import { Route, active, router } from "tinro";
  import UrlAction from "./URLAction.svelte";
  import NewUrlAction from "./NewURLAction.svelte";
  import Fa from "svelte-fa";
  import { faArrowLeft, faBackward, faHome } from "@fortawesome/free-solid-svg-icons";
  import { toast } from "@zerodevx/svelte-toast";

  let config: Config = {
    shouldDeleteOffline: false,
    shouldSaveOffline: false,
    urlActions: [],
  };
  const removeURLAction = (ev: CustomEvent<{ pattern: string }>) => {
    config = {
      ...config,
      urlActions: config.urlActions.filter(({ pattern }) => !(pattern === ev.detail.pattern)),
    };
  };
  const addURLAction = (ev: CustomEvent<{ urlAction: URLAction }>) => {
    config = {
      ...config,
      urlActions: [...config.urlActions, ev.detail.urlAction],
    };
  };

  onMount(async () => {
    config = await api("/config");
    config.urlActions = config.urlActions || [];
  });
</script>

<nav>
  <div class="row">
    <button
      on:click={() => {
        router.goto("/");
      }}
    >
      <Fa icon={faArrowLeft} /> Home</button
    >
    <h1>Rengoku Settings</h1>
  </div>
  <hr />
</nav>

<div class="setting-wrapper">
  <aside id="menu" role="menu">
    <a href="/settings/account" use:active data-active-class="active-menu">My Account</a>
    <a href="/settings/global" use:active data-active-class="active-menu">Global Settings</a>
    <a href="/settings/url-actions" use:active data-active-class="active-menu">URL Actions</a>
    <a href="/settings/ui" use:active data-active-class="active-menu">UI Settings</a>
  </aside>

  <div id="setting-content" class="col">
    <Route path="/account">Coming Soon</Route>
    <Route path="/ui">Coming Soon</Route>

    <Route path="/global">
      <div class="row m-b-2">
        <h3>Save offline version of the page to disk when adding bookmarks</h3>
        <input
          class="m-l-auto"
          type="checkbox"
          name="saveOfflineDefault"
          id="saveOfflineDefault"
          bind:checked={config.shouldSaveOffline}
          on:change={async (e) => {
            const { checked } = e.currentTarget;
            try {
              await api("/config", "PUT", { shouldSaveOffline: checked });
            } catch (error) {
              toast.push("Something went wrong.");
              console.error(error);
            }
          }}
        />
      </div>
      <div class="row m-b-2">
        <h3>Remove saved pages from disk when deleting bookmarks</h3>
        <input
          class="m-l-auto"
          type="checkbox"
          name="deleteOfflineDefault"
          id="deleteOfflineDefault"
          bind:checked={config.shouldDeleteOffline}
          on:change={async (e) => {
            const { checked } = e.currentTarget;
            try {
              await api("/config", "PUT", { shouldDeleteOffline: checked });
            } catch (error) {
              toast.push("Something went wrong.");
              console.error(error);
            }
          }}
        />
      </div>
    </Route>

    <Route path="/url-actions">
      <h3>URL Actions</h3>
      <hr />
      <div>
        {#each config.urlActions || [] as urlAction, i}
          <UrlAction {urlAction} key={i} on:removeURLAction={removeURLAction} />
        {/each}
      </div>
      <NewUrlAction on:addURLAction={addURLAction} />
    </Route>
  </div>
</div>

<style>
  nav {
    grid-area: nav;
    padding-top: 20px;
    padding-bottom: 1em;
    position: sticky;
    top: 0px;
    background-color: var(--background-body);
  }
  .setting-wrapper {
    display: flex;
  }
  #menu {
    display: flex;
    flex-direction: column;
    margin-right: 2em;
    width: 15vw;
  }
  #menu a {
    padding: 0.5em 1em;
    font-weight: bold;
    text-decoration: none;
    color: antiquewhite;
    border-radius: 6px;
  }
  #menu a:hover,
  :global(.active-menu) {
    background-color: var(--background-alt);
  }
  #setting-content {
    width: 100%;
  }
</style>
