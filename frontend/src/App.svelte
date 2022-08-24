<script lang="ts">
  import "./styles/water.min.css";
  import "./styles/style.css";

  import { Route } from "tinro";

  import Home from "./Home.svelte";
  import SettingsPage from "./components/SettingsPage.svelte";
  import { store } from "@lib";

  const { tags, bookmarks } = store;
</script>

{#await Promise.all([tags.init(), bookmarks.init()]) then}
  <Route path="/">
    <Home />
  </Route>
  <Route path="/settings/*">
    <SettingsPage />
  </Route>
{/await}
