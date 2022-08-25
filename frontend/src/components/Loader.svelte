<script lang="ts">
  import { faSpinner, type IconDefinition } from "@fortawesome/free-solid-svg-icons";

  import Fa from "svelte-fa";
  export let size: "xs" | "sm" | "lg" | `${number}x` = "lg";
</script>

<div id="loader">
  <Fa icon={faSpinner} {size} spin />
</div>

<style>
  #loader {
    width: 100%;
    display: flex;
    justify-content: center;
    padding: 2em 0;
  }
</style>
