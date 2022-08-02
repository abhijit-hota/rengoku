<script lang="ts">
  import { faSpinner } from "@fortawesome/free-solid-svg-icons";

  import Fa from "svelte-fa";
</script>

<div id="loader">
  <Fa icon={faSpinner} size="lg" spin />
</div>

<style>
  #loader {
    width: 100%;
    display: flex;
    justify-content: center;
    padding: 2em 0;
  }
</style>
