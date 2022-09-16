<script lang="ts">
  import { faEllipsis } from "@icons";
  import Fa from "svelte-fa";
  import OutClick from "./OutClick.svelte";

  let menuOpen = false;
  let focusedID = 0;
  let ref: HTMLUListElement = null;
  let listItems: HTMLLIElement[] = [];

  const setup = () => {
    if (ref) {
      listItems = [...ref.getElementsByTagName("li")];
    }
  };

  $: setup(), ref;

  $: if (listItems.length > 0) {
    const btn = listItems[focusedID].firstElementChild as HTMLButtonElement;
    btn.focus();
  }

  // Temp Props
  export let excludeClicks: string[] = [];
  export let id: string = Math.random().toString();
</script>

<div class="popup">
  <button id={"menu-button-" + id} class="icon-button" on:click={() => (menuOpen = !menuOpen)}>
    <Fa icon={faEllipsis} />
  </button>
  <OutClick
    on:outclick={() => {
      menuOpen = false;
    }}
    excludeByQuerySelector={["#menu-button-" + id, ...excludeClicks]}
  >
    {#if menuOpen}
      <ul
        class="menu"
        on:keydown={(e) => {
          const { key } = e;
          if (!(key === "ArrowUp" || key === "ArrowDown")) {
            return;
          }
          e.preventDefault();
          if (key === "ArrowDown") {
            focusedID = (focusedID + 1) % listItems.length;
          } else if (key === "ArrowUp") {
            if (focusedID <= 0) {
              focusedID = listItems.length - 1;
            } else {
              focusedID--;
            }
          }
        }}
        bind:this={ref}
      >
        <slot name="list-items" />
      </ul>
    {/if}
  </OutClick>
</div>

<style>
  .popup .menu {
    list-style-type: none;

    background-color: var(--button-base);
    border-radius: 6px;

    padding: 0;
    position: absolute;
    margin-top: 0.5em;
    z-index: 9999;
  }
  .popup .menu :global(*) {
    z-index: 9999;
    position: relative;
  }
  .popup .menu :global(hr) {
    margin: 0;
  }
  .popup .menu :global(li > *) {
    padding: 0.75em;
    cursor: pointer;
    width: calc(100% - 0.75em * 2);
    position: relative;
  }
  .popup .menu :global(li > *:hover),
  .popup .menu :global(li > *:focus) {
    background-color: var(--button-hover);
  }
  .popup .menu :global(li:last-child > *) {
    border-radius: 0 0 6px 6px;
  }
  .popup .menu :global(li:first-child > *) {
    border-radius: 6px 6px 0 0;
  }
</style>
