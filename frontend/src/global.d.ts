type Bookmark = {
  id: number;
  url: string;
  meta: {
    title: string;
    description: string;
    favicon: string;
  };
  created: number;
  last_updated: number;
  last_saved_offline: number;
  tags: {
    id: number;
    name: string;
    created: number;
    last_updated: number;
  }[];
};
