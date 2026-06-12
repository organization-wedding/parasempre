import { describe, test, expect } from "bun:test";
import { validateMediaFile, MEDIA_LIMITS } from "./giftMessage";

function makeFile(name: string, type: string, sizeBytes: number): File {
  const bytes = new Uint8Array(sizeBytes);
  return new File([bytes], name, { type });
}

describe("validateMediaFile", () => {
  test("rejeita PDF", () => {
    const file = makeFile("doc.pdf", "application/pdf", 1024);
    const result = validateMediaFile(file);
    expect(result).not.toBeNull();
    expect(result).toMatch(/não suportado/i);
  });

  test("rejeita arquivo vazio", () => {
    const file = makeFile("empty.jpg", "image/jpeg", 0);
    const result = validateMediaFile(file);
    expect(result).toBe("Arquivo vazio.");
  });

  test("aceita imagem JPEG dentro do limite de 5 MB", () => {
    const file = makeFile("photo.jpg", "image/jpeg", 5 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("rejeita imagem JPEG acima de 5 MB", () => {
    const file = makeFile("photo.jpg", "image/jpeg", 5 * 1024 * 1024 + 1);
    const result = validateMediaFile(file);
    expect(result).not.toBeNull();
    expect(result).toMatch(/5 MB/);
    expect(result).toMatch(/imagem/i);
  });

  test("aceita áudio MP3 dentro do limite de 5 MB", () => {
    const file = makeFile("song.mp3", "audio/mpeg", 5 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("rejeita áudio MP3 acima de 5 MB", () => {
    const file = makeFile("song.mp3", "audio/mpeg", 5 * 1024 * 1024 + 1);
    const result = validateMediaFile(file);
    expect(result).not.toBeNull();
    expect(result).toMatch(/5 MB/);
    expect(result).toMatch(/áudio/i);
  });

  test("aceita vídeo MP4 com exatamente 50 MB", () => {
    const file = makeFile("clip.mp4", "video/mp4", 50 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("aceita vídeo MP4 abaixo de 50 MB", () => {
    const file = makeFile("clip.mp4", "video/mp4", 25 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("rejeita vídeo MP4 acima de 50 MB com mensagem '50 MB'", () => {
    const file = makeFile("clip.mp4", "video/mp4", 50 * 1024 * 1024 + 1);
    const result = validateMediaFile(file);
    expect(result).not.toBeNull();
    expect(result).toMatch(/50 MB/);
    expect(result).toMatch(/vídeo/i);
  });

  test("aceita vídeo WebM dentro do limite de 50 MB", () => {
    const file = makeFile("clip.webm", "video/webm", 40 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("aceita vídeo MOV (video/quicktime) dentro do limite de 50 MB", () => {
    const file = makeFile("clip.mov", "video/quicktime", 30 * 1024 * 1024);
    expect(validateMediaFile(file)).toBeNull();
  });

  test("MEDIA_LIMITS.video.bytes é 50 MB", () => {
    expect(MEDIA_LIMITS.video.bytes).toBe(50 * 1024 * 1024);
  });

  test("MEDIA_LIMITS.image.bytes permanece 5 MB", () => {
    expect(MEDIA_LIMITS.image.bytes).toBe(5 * 1024 * 1024);
  });

  test("MEDIA_LIMITS.audio.bytes permanece 5 MB", () => {
    expect(MEDIA_LIMITS.audio.bytes).toBe(5 * 1024 * 1024);
  });
});
