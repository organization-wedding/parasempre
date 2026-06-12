import { z } from "zod";

export const MEDIA_KIND_VALUES = ["image", "audio", "video"] as const;

export type MediaKind = (typeof MEDIA_KIND_VALUES)[number];

export const MEDIA_LIMITS: Record<MediaKind, { bytes: number; mimes: readonly string[] }> = {
  image: {
    bytes: 5 * 1024 * 1024,
    mimes: ["image/jpeg", "image/png", "image/webp"],
  },
  audio: {
    bytes: 5 * 1024 * 1024,
    mimes: ["audio/mpeg", "audio/mp4", "audio/x-m4a", "audio/ogg"],
  },
  video: {
    bytes: 50 * 1024 * 1024,
    mimes: ["video/mp4", "video/webm", "video/quicktime"],
  },
};

export const MAX_AUTHOR_NAME_LEN = 120;
export const MAX_CONTENT_LEN = 500;

export const publicMessageSchema = z.object({
  id: z.number().int(),
  gift_id: z.number().int(),
  author_name: z.string(),
  content: z.string(),
  media_url: z.string().nullable(),
  media_kind: z.enum(MEDIA_KIND_VALUES).nullable(),
  created_at: z.string(),
});

export const adminMessageSchema = publicMessageSchema.extend({
  user_id: z.number().int(),
  gift_transaction_id: z.number().int(),
});

export const pagedPublicMessagesSchema = z.object({
  data: z.array(publicMessageSchema),
  page: z.number().int(),
  limit: z.number().int(),
  total: z.number().int(),
});

export const pagedAdminMessagesSchema = z.object({
  data: z.array(adminMessageSchema),
  page: z.number().int(),
  limit: z.number().int(),
  total: z.number().int(),
});

export type PublicMessage = z.infer<typeof publicMessageSchema>;
export type AdminMessage = z.infer<typeof adminMessageSchema>;
export type PagedPublicMessages = z.infer<typeof pagedPublicMessagesSchema>;
export type PagedAdminMessages = z.infer<typeof pagedAdminMessagesSchema>;

export function detectMediaKind(mime: string): MediaKind | null {
  for (const kind of MEDIA_KIND_VALUES) {
    if (MEDIA_LIMITS[kind].mimes.includes(mime)) return kind;
  }
  return null;
}

export function validateMediaFile(file: File): string | null {
  const kind = detectMediaKind(file.type);
  if (!kind) {
    return "Formato de arquivo não suportado. Use imagem (JPG/PNG/WEBP), áudio (MP3/M4A/OGG) ou vídeo (MP4/WEBM/MOV).";
  }
  const limit = MEDIA_LIMITS[kind];
  if (file.size <= 0) {
    return "Arquivo vazio.";
  }
  if (file.size > limit.bytes) {
    const mb = Math.round(limit.bytes / (1024 * 1024));
    return `Arquivo excede o limite de ${mb} MB para ${kind === "image" ? "imagem" : kind === "audio" ? "áudio" : "vídeo"}.`;
  }
  return null;
}
