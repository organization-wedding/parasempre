export type RsvpStage =
  | "cover"
  | "opening"
  | "pages-flipping"
  | "family"
  | "signing"
  | "closing-pages"
  | "closing-cover"
  | "done";

export const STAGE_LABELS: Record<RsvpStage, string> = {
  cover: "Capa do livro",
  opening: "Livro abrindo",
  "pages-flipping": "Folheando páginas",
  family: "Página da família",
  signing: "Assinando",
  "closing-pages": "Folheando de volta",
  "closing-cover": "Fechando o livro",
  done: "Resposta registrada",
};

export const STAGE_TIMING = {
  opening: 1400,
  pagesFlipping: 1800,
  signing: 2500,
  closingPages: 1200,
  closingCover: 1400,
} as const;
