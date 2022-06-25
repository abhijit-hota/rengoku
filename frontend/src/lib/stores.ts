import { writable, derived } from "svelte/store";
import { api } from ".";

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

type Bookmark = {
  id: number;
  url: string;
  meta: {
    title: string;
    description: string;
    favicon: string;
  };
  created_at: number;
  last_updated: number;
  last_saved_offline: number;
  tags: {
    id: number;
    name: string;
    created_at: number;
    last_updated: number;
  }[];
};

const createBookmarksStore = () => {
  const bookmarks = writable<Bookmark[]>([]);
  const { set, subscribe, update } = bookmarks;
  return {
    subscribe,
    fetch: async (q: string) => {
      try {
        const res = await api("/bookmarks" + q);
        set(res.data);
      } catch (error) {
        return set([]);
      }
    },
    add: (...newBookmarks: Bookmark[]) => {
      update((bookmarks) => [...bookmarks, ...newBookmarks]);
    },
    delete: (id: Bookmark["id"]) => {
      update((bookmarks) => {
        const i = bookmarks.findIndex((bookmark) => bookmark.id === id);
        bookmarks.splice(i, 1);
        return bookmarks;
      });
    },
    updateOne: (id: Bookmark["id"], updatedBookmark: Bookmark) => {
      update((bookmarks) =>
        bookmarks.map((bookmark) =>
          bookmark.id === id ? { ...bookmark, ...updatedBookmark } : bookmark
        )
      );
    },
  };
};

export const bookmarks = createBookmarksStore();
