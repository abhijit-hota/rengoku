export { default as NewURLAction } from "./NewURLAction.svelte";
export { default as URLAction } from "./URLAction.svelte";

export type URLActionType = {
  pattern: string;
  matchDetection: string;
  shouldSaveOffline: boolean;
  tags: number[];
  folders: number[];
};
export type Config = {
  shouldDeleteOffline: boolean;
  shouldSaveOffline: boolean;
  urlActions: URLActionType[];
};
