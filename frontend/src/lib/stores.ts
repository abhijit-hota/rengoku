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

export type Bookmark = {
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
    set,
    subscribe,
    update,
    add: (...newBookmarks: Bookmark[]) => {
      update((bookmarks) => [...bookmarks, ...newBookmarks]);
    },
    delete: (...ids: Bookmark["id"][]) => {
      update((bookmarks) => {
        return bookmarks.filter(({ id }) => !ids.includes(id));
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
