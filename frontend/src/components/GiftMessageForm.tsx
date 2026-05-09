import { useEffect, useRef, useState } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Send from "lucide-react/dist/esm/icons/send";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import Paperclip from "lucide-react/dist/esm/icons/paperclip";
import X from "lucide-react/dist/esm/icons/x";
import { useCreateGiftMessageMutation } from "../lib/message-queries";
import {
  MAX_AUTHOR_NAME_LEN,
  MAX_CONTENT_LEN,
  detectMediaKind,
  validateMediaFile,
  type PublicMessage,
} from "../schemas/giftMessage";

const formSchema = z.object({
  author_name: z
    .string()
    .trim()
    .min(1, "Como você gostaria de assinar?")
    .max(MAX_AUTHOR_NAME_LEN, `Máximo ${MAX_AUTHOR_NAME_LEN} caracteres.`),
  content: z
    .string()
    .trim()
    .min(1, "Escreva uma mensagem para os noivos.")
    .max(MAX_CONTENT_LEN, `Máximo ${MAX_CONTENT_LEN} caracteres.`),
});

type FormValues = z.infer<typeof formSchema>;

interface Props {
  transactionId: number;
  defaultAuthorName?: string;
  onCreated?: (message: PublicMessage) => void;
  onCancel?: () => void;
}

export function GiftMessageForm({
  transactionId,
  defaultAuthorName,
  onCreated,
  onCancel,
}: Props) {
  const mutation = useCreateGiftMessageMutation(transactionId);
  const [file, setFile] = useState<File | null>(null);
  const [filePreviewURL, setFilePreviewURL] = useState<string | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      author_name: defaultAuthorName ?? "",
      content: "",
    },
  });

  useEffect(() => {
    if (!file) {
      setFilePreviewURL(null);
      return;
    }
    const url = URL.createObjectURL(file);
    setFilePreviewURL(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    setFileError(null);
    const next = e.target.files?.[0];
    if (!next) {
      setFile(null);
      return;
    }
    const validationError = validateMediaFile(next);
    if (validationError) {
      setFileError(validationError);
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = "";
      return;
    }
    setFile(next);
  }

  function clearFile() {
    setFile(null);
    setFileError(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  async function onSubmit(values: FormValues) {
    if (fileError) return;
    const formData = new FormData();
    formData.set("author_name", values.author_name.trim());
    formData.set("content", values.content.trim());
    if (file) formData.set("media", file);

    const created = await mutation.mutateAsync(formData);
    onCreated?.(created);
  }

  const contentValue = watch("content") ?? "";
  const mediaKind = file ? detectMediaKind(file.type) : null;

  const submitError =
    mutation.error instanceof Error ? mutation.error.message : null;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <label className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
          Assinatura
        </label>
        <input
          type="text"
          maxLength={MAX_AUTHOR_NAME_LEN}
          placeholder="Como você quer assinar (ex.: Família Silva)"
          {...register("author_name")}
          className="border border-gold-muted/50 bg-ivory text-dark-warm text-[0.9rem] px-3 py-2 outline-none focus:border-burgundy transition-colors"
        />
        {errors.author_name && (
          <span className="text-[0.78rem] text-[#7a2e2b]">{errors.author_name.message}</span>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between">
          <label className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
            Mensagem
          </label>
          <span className="text-[0.7rem] text-hint">
            {contentValue.length}/{MAX_CONTENT_LEN}
          </span>
        </div>
        <textarea
          rows={5}
          maxLength={MAX_CONTENT_LEN}
          placeholder="Escreva uma mensagem para os noivos…"
          {...register("content")}
          className="border border-gold-muted/50 bg-ivory text-dark-warm text-[0.9rem] px-3 py-2 outline-none focus:border-burgundy transition-colors resize-y"
        />
        {errors.content && (
          <span className="text-[0.78rem] text-[#7a2e2b]">{errors.content.message}</span>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <label className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
          Mídia (opcional)
        </label>
        <div className="flex items-center gap-3 flex-wrap">
          <label
            htmlFor={`gift-msg-media-${transactionId}`}
            className="inline-flex items-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2 px-3 border border-gold-muted/60 text-hint hover:border-burgundy hover:text-burgundy transition-colors cursor-pointer"
          >
            <Paperclip size={14} />
            {file ? "Trocar arquivo" : "Anexar foto, áudio ou vídeo"}
          </label>
          <input
            id={`gift-msg-media-${transactionId}`}
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp,audio/mpeg,audio/mp4,audio/x-m4a,audio/ogg,video/mp4,video/webm,video/quicktime"
            onChange={handleFileChange}
            className="hidden"
          />
          {file && (
            <button
              type="button"
              onClick={clearFile}
              className="inline-flex items-center gap-1.5 font-heading text-[0.68rem] font-semibold tracking-[0.06em] uppercase text-hint hover:text-burgundy transition-colors cursor-pointer"
            >
              <X size={12} />
              Remover
            </button>
          )}
        </div>
        {fileError && (
          <div className="flex items-center gap-2 rounded border border-[#c25550]/30 bg-[#fef2f1] px-3 py-2">
            <AlertTriangle size={14} className="text-[#c25550] shrink-0" />
            <span className="text-[0.78rem] text-[#7a2e2b]">{fileError}</span>
          </div>
        )}
        {file && filePreviewURL && (
          <div className="border border-gold-muted/40 bg-parchment-dark p-3">
            <p className="text-[0.7rem] text-hint mb-2 truncate">{file.name}</p>
            {mediaKind === "image" && (
              <img src={filePreviewURL} alt="prévia" className="max-h-[180px] object-contain mx-auto" />
            )}
            {mediaKind === "audio" && (
              <audio controls src={filePreviewURL} className="w-full" />
            )}
            {mediaKind === "video" && (
              <video controls src={filePreviewURL} className="max-h-[220px] w-full" />
            )}
          </div>
        )}
      </div>

      {submitError && (
        <div className="flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-3 py-2">
          <AlertTriangle size={14} className="text-[#c25550] shrink-0" />
          <span className="text-[0.8rem] text-[#7a2e2b] flex-1">{submitError}</span>
        </div>
      )}

      <div className="flex items-center justify-end gap-3 pt-2">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={isSubmitting || mutation.isPending}
            className="font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 text-hint hover:text-burgundy transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Cancelar
          </button>
        )}
        <button
          type="submit"
          disabled={isSubmitting || mutation.isPending}
          className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Send size={13} />
          {mutation.isPending ? "Enviando…" : "Enviar mensagem"}
        </button>
      </div>
    </form>
  );
}
