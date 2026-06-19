export type PageDecision = { page: number; savedPage: number | null };

export function nextPageOnSearchChange(
  prevSearch: string,
  nextSearch: string,
  currentPage: number,
  savedPage: number | null,
): PageDecision {
  const was = prevSearch.trim();
  const now = nextSearch.trim();
  if (was === now) return { page: currentPage, savedPage };
  if (was === "" && now !== "") return { page: 1, savedPage: currentPage };
  if (was !== "" && now === "") return { page: savedPage ?? 1, savedPage: null };
  return { page: 1, savedPage };
}
