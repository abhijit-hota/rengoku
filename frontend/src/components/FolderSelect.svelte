<script context="module" lang="ts">
  type Tree = store.Tree;

  const flattenTree = (tree: Tree): { id: number; label: string }[] => {
    const res: { id: number; label: string }[] = [];

    const rec = (tree: Tree, parentLabel: string = "") => {
      tree.forEach(({ children, id, name }) => {
        const currentLabel = parentLabel + name + "/";

        res.push({
          label: currentLabel,
          id,
        });
        if ((children?.length ?? 0) > 0) {
          rec(children, currentLabel);
        }
      });
    };

    rec(tree);
    return res;
  };
</script>

<script lang="ts">
  import { store } from "@lib";
  const { folders } = store;

  import MultiSelect from "svelte-multiselect";

  let selected = [];
  export let selectedFolderId: Number;
  $: selectedFolderId = selected[0]?.value ?? null;
  $: flattened = flattenTree($folders);
</script>

<MultiSelect
  inputClass="input-like"
  outerDivClass="color-fix"
  maxSelect={1}
  options={flattened.map(({ id, label }) => ({ label, value: id }))}
  bind:selected
/>
