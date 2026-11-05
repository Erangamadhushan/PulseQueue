type ClassValue =
  | string
  | false
  | null
  | undefined
  | ((state: never) => string | undefined);

export function cn(...values: ClassValue[]): string {
  return values.filter((value): value is string => typeof value === "string").join(" ");
}
