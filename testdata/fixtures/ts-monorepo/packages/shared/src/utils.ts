export interface ApiResponse<T> {
  data: T;
  timestamp: string;
}

export function formatResponse<T>(data: T): ApiResponse<T> {
  return {
    data,
    timestamp: new Date().toISOString(),
  };
}

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "");
}
