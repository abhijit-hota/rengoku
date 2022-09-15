<script context="module" lang="ts">
  import { writable } from "svelte/store";

  export const isExpanded = writable<Record<number, boolean>>({});
  export const isParent = writable<Record<number, boolean>>({});
</script>

<script lang="ts">
  import { store } from "@lib";
  import { Folder } from "@components";

  const { folders } = store;
</script>

<Folder tree={$folders} />
