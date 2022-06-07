<script context="module">
  // retain module scoped expansion state for each tree node
  const _expansionState = {
    /* treeNodeId: expanded <boolean> */
  };
</script>

<script>
  import { queryParams } from "../lib/stores";

  //	import { slide } from 'svelte/transition'
  export let tree;
  const { name, id, children } = tree;

  let expanded = _expansionState[id] || true;
  const toggleExpansion = () => {
    expanded = _expansionState[id] = !expanded;
  };
  $: arrowDown = expanded;
</script>

<ul>
  <!-- transition:slide -->
  <li>
    <span>
      <input
        type="checkbox"
        name={id}
        {id}
        bind:value={$queryParams.folder}
        on:change={(e) => {
          // @ts-ignore
          if (e.target.checked) {
            $queryParams.folder = id;
          } else {
            $queryParams.folder = null;
          }
        }}
      />
      <label for={id} style="cursor: pointer;">{name}</label>
      {#if children}
        <span on:click={toggleExpansion} class="arrow" class:arrowDown>&#x25b6</span>
        {#if expanded}
          {#each children as child}
            <svelte:self tree={child} />
          {/each}
        {/if}
      {/if}
    </span>
  </li>
</ul>

<style>
  ul {
    margin: 0;
    list-style: none;
    padding-left: 1rem;
    user-select: none;

    padding-top: 0.5em;
    padding-bottom: 0.5em;
    cursor: pointer;
  }
  label {
    font-size: larger;
  }
  .arrow {
    cursor: pointer;
    display: inline-block;
    /* transition: transform 200ms; */
  }
  .arrowDown {
    transform: rotate(90deg);
  }
</style>
