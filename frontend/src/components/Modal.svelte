<script context="module" lang="ts">
  import { writable } from "svelte/store";

  export const modals = writable<Record<string, HTMLDialogElement>>({});
</script>

<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  export let key = "";

  let modal: HTMLDialogElement;
  onMount(() => ($modals[key] = modal));
  onDestroy(() => delete $modals[key]);

  const dialogClickHandler: svelteHTML.MouseEventHandler<HTMLDialogElement> = (e) => {
    if (e.currentTarget.tagName !== "DIALOG") return;

    const rect = e.currentTarget.getBoundingClientRect();

    const clickedInDialog =
      rect.top <= e.clientY &&
      e.clientY <= rect.top + rect.height &&
      rect.left <= e.clientX &&
      e.clientX <= rect.left + rect.width;

    if (!clickedInDialog) {
      modal.close();
    }
  };

  export let styles: Partial<Record<"dialog" | "header" | "body" | "footer", string>> = {};
</script>

<dialog bind:this={modal} on:click={dialogClickHandler} style={styles.dialog}>
  <header style={styles.header}>
    <div
      class="row"
      style="justify-content: space-between; padding-left: 20px; padding-right: 20px;"
    >
      <slot name="header" />
      <button class="icon-button" on:click={() => modal.close()}>✖</button>
    </div>
  </header>
  <div style={styles.body}>
    <slot name="body" />
  </div>
  <footer style={styles.footer}>
    <slot name="footer" />
  </footer>
</dialog>
