export function toTitleCase(value: string) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/\b\w/g, (char) => char.toUpperCase());
}
