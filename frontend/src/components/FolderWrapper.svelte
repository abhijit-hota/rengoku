<script context="module" lang="ts">
  import { writable } from "svelte/store";
  export type Tree = { id?: number; name?: string; children?: Tree }[];

  export const isExpanded = writable<Record<number, boolean>>({});
  export const isParent = writable<Record<number, boolean>>({});
</script>

<script lang="ts">
  import { api } from "@lib";
  import { Folder } from "@components";
  import Loader from "./Loader.svelte";
</script>

{#await api("/folders/tree")}
  <Loader />
{:then tree}
  <Folder {tree} />
{:catch}
  <span class="red"> Error occurred while loading folders.</span>
{/await}
