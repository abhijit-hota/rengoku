<script lang="ts">
  import "./styles/water.min.css";
  import "./styles/style.css";

  import { Route, router } from "tinro";

  import Home from "./views/Home/Home.svelte";
  import SettingsPage from "./views/Settings/SettingsPage.svelte";
  import { store } from "@lib";
  import Loader from "./components/Loader.svelte";
  import LoginPage from "./views/Login/LoginPage.svelte";

  const { tags, bookmarks, folders, auth } = store;

  let loading = true;

  const handleLogin = async () => {
    if ($auth.loggedIn) {
      Promise.all([tags.init(), bookmarks.init(), folders.init()]).then(() => {
        loading = false;
        router.goto("/");
      });
    } else {
      router.goto("/login");
    }
  };

  $: handleLogin(), $auth.loggedIn;
</script>

{#if !$auth.loggedIn}
  <Route path="/login">
    <LoginPage />
  </Route>
{:else if loading}
  <div id="full-body-loader-wrapper">
    <Loader size="4x" />
  </div>
{:else}
  <Route path="/">
    <Home />
  </Route>
  <Route path="/settings" redirect="/settings/account" />
  <Route path="/settings/*">
    <SettingsPage />
  </Route>
{/if}

<style>
  #full-body-loader-wrapper {
    position: absolute;
    top: 50%;
    right: 50%;
    transform: translate(50%, -50%);
  }
</style>
