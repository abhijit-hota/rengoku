<script context="module" lang="ts">
  import { writable } from "svelte/store";

  type BatchActions = "DELETE" | "TAGS" | "FOLDER" | "OFFLINE" | "REFETCHMETA";

  export const batchActionInProgress = writable<BatchActions | "">("");
</script>

<script lang="ts">
  import Fa from "svelte-fa";
  import { faSpinner, IconDefinition } from "@fortawesome/free-solid-svg-icons";
  import { toast } from "@zerodevx/svelte-toast";

  export let action: BatchActions;
  export let props: any = {};
  export let icon: IconDefinition;
  export let title: string;

  const sleep = (ms: number = 1000) => new Promise((resolve) => setTimeout(resolve, ms));
  export let handler: () => Promise<unknown> = sleep;
</script>

<!-- class="icon-button" -->
<button
  disabled={$batchActionInProgress !== ""}
  {title}
  on:click={async () => {
    try {
      $batchActionInProgress = action;
      await handler();
    } catch (error) {
      toast.push(error.message ?? "Error");
    } finally {
      $batchActionInProgress = "";
    }
  }}
  {...props}
>
  <Fa
    icon={$batchActionInProgress === action ? faSpinner : icon}
    spin={$batchActionInProgress === action}
  />
</button>
