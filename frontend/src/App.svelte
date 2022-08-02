<script lang="ts">
  import "./styles/water.min.css";
  import "./styles/style.css";

  import { Route } from "tinro";

  import { modals } from "@Modal";
  import Home from "./Home.svelte";
  import SettingsPage from "./components/SettingsPage.svelte";
</script>

<Route path="/">
  <Home />
</Route>
<Route path="/settings">
  <SettingsPage />
</Route>
