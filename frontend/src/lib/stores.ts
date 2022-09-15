import { writable, derived } from "svelte/store";
import type { Writable } from "svelte/store";

import api from "./api";

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
  moreLeft: false,
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
  folders: number[];
};

export type Tag = {
  id: number;
  name: string;
  created_at: number;
  last_updated: number;
};

export type Tree = { id?: number; name?: string; children?: Tree }[];

const createBookmarksStore = () => {
  const bookmarks = writable<Bookmark[]>([]);
  const { set, subscribe, update } = bookmarks;
  return {
    set,
    subscribe,
    update,
    init: async () => {
      try {
        const res = await api("/bookmarks");
        set(res.data);
        stats.set({
          total: res.total,
          moreLeft: res.total > 20,
          page: res.page,
        });
      } catch (error) {
        // TODO
        set([]);
      }
    },
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

const createTagsStore = () => {
  const tags = writable<Tag[]>([]);
  const { set, subscribe, update } = tags;
  return {
    subscribe,
    init: async () => {
      try {
        const res = await api("/tags");
        set(res);
      } catch (error) {
        // TODO
        set([]);
      }
    },
    add: (...newTags: Tag[]) => {
      update((tags) => [...tags, ...newTags]);
    },
    delete: (...ids: Tag["id"][]) => {
      update((tags) => {
        return tags.filter(({ id }) => !ids.includes(id));
      });
    },
    updateOne: (id: Tag["id"], updatedTag: Tag) => {
      update((tags) => tags.map((tag) => (tag.id === id ? { ...tag, ...updatedTag } : tag)));
    },
  };
};

const createFolderStore = () => {
  const folders = writable<Tree>([]);
  const flattened = derived<Writable<Tree>, { id: number; label: string }[]>(
    folders,
    (tree, set) => {
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
      set(res);
    }
  );
  const { set, subscribe, update } = folders;
  return {
    subscribe,
    init: async () => {
      try {
        const res = await api("/folders/tree");
        set(res);
      } catch (error) {
        // TODO
        set([]);
      }
    },
    flattened,
    // add: (...newTags: Tag[]) => {
    //   update((tags) => [...tags, ...newTags]);
    // },
    // delete: (...ids: Tag["id"][]) => {
    //   update((tags) => {
    //     return tags.filter(({ id }) => !ids.includes(id));
    //   });
    // },
    // updateOne: (id: Tag["id"], updatedTag: Tag) => {
    //   update((tags) => tags.map((tag) => (tag.id === id ? { ...tag, ...updatedTag } : tag)));
    // },
  };
};

const createAuthStore = () => {
  const { set, subscribe } = writable({
    loggedIn: !!window.localStorage.getItem("AUTH_TOKEN"),
    token: "",
  });
  return {
    subscribe,
    login: (token: string) => {
      window.localStorage.setItem("AUTH_TOKEN", token);
      set({ loggedIn: true, token });
    },
    logout: () => {
      window.localStorage.removeItem("AUTH_TOKEN");
      set({ loggedIn: false, token: "" });
    },
  };
};

export const bookmarks = createBookmarksStore();
export const tags = createTagsStore();
export const auth = createAuthStore();
export const folders = createFolderStore();
