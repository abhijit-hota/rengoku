import { writable, derived } from "svelte/store";
import { api } from "../lib";

export const queryParams = writable({
  sortBy: "date",
  order: "asc",
  folder: "",
  tags: [],
  search: "",
  page: 0,
});

export const stats = writable({
  page: 0,
  total: 0,
  moreLeft: true,
});

export const queryStr = derived(queryParams, (store) => {
  const queryParamStrings = [];
  if (store.sortBy && store.order) {
    const param = `sort_by=${store.sortBy}&order=${store.order}`;
    queryParamStrings.push(param);
  }
  if (store.folder) {
    const param = `folder=${store.folder}`;
    queryParamStrings.push(param);
  }
  if (store.search) {
    const param = `search=${store.search}`;
    queryParamStrings.push(param);
  }
  if (store.tags.length > 0) {
    const tagStr = store.tags.map((id) => `tags[]=${id}`).join("&");
    const param = tagStr;
    queryParamStrings.push(param);
  }
  queryParamStrings.push(`page=${store.page}`);
  return "?" + queryParamStrings.join("&");
});

const createBookmarksStore = () => {
  const { set, subscribe, update } = writable([]);
  return {
    subscribe,
    fetch: async (q) => {
      try {
        const res = await api("/bookmarks" + q);
        set(res.data);
        update((bookmarks) => [
          ...bookmarks,
          ...res.data.filter(({ id: newID }) => !bookmarks.find(({ id }) => id === newID)),
        ]);
      } catch (error) {
        return set([]);
      }
    },
    add: (...newBookmarks) => {
      update((bookmarks) => [...bookmarks, ...newBookmarks]);
    },
    delete: (id) => {
      update((bookmarks) => {
        const i = bookmarks.findIndex((bookmark) => bookmark.id === id);
        bookmarks.splice(i, 1);
        return bookmarks;
      });
    },
  };
};

export const bookmarks = createBookmarksStore();
