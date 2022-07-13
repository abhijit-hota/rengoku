<script context="module" lang="ts">
  import { writable } from "svelte/store";
  export type Tree = { id?: number; name?: string; children?: Tree }[];

  export const isExpanded = writable<Record<number, boolean>>({});
  export const isParent = writable<Record<number, boolean>>({});
</script>

<script lang="ts">
  import { api } from "@lib";
  import { Folder } from "@components";
</script>

{#await api("/folders/tree")}
  Loading...
{:then tree}
  <Folder {tree} />
{:catch}
  Error occurred while loading folders.
{/await}
