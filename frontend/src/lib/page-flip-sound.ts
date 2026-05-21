let audioCtx: AudioContext | null = null;

function getCtx(): AudioContext | null {
  try {
    if (!audioCtx) {
      const Ctor =
        window.AudioContext ??
        (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      if (!Ctor) return null;
      audioCtx = new Ctor();
    }
    if (audioCtx.state === "suspended") void audioCtx.resume();
    return audioCtx;
  } catch {
    return null;
  }
}

// Must be invoked inside a user gesture (click, keydown). Browsers gate
// AudioContext.resume() on a gesture — without this, the first batch of
// page-flip sounds plays into a suspended context and is silently dropped.
export function unlockAudio(): void {
  const ctx = getCtx();
  if (!ctx) return;
  try {
    const buf = ctx.createBuffer(1, 1, ctx.sampleRate);
    const src = ctx.createBufferSource();
    src.buffer = buf;
    src.connect(ctx.destination);
    src.start(0);
  } catch {
    // noop
  }
}

export function playPageFlip(volume = 0.12): void {
  const ctx = getCtx();
  if (!ctx) return;
  try {
    const bufferSize = Math.floor(ctx.sampleRate * 0.25);
    const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
    const data = buffer.getChannelData(0);
    for (let i = 0; i < bufferSize; i++) {
      const t = i / bufferSize;
      // sine-curved envelope: gentle attack and gentle decay, no percussive front
      const env = Math.sin(t * Math.PI) * (1 - t * 0.4);
      data[i] = (Math.random() * 2 - 1) * env;
    }
    const src = ctx.createBufferSource();
    src.buffer = buffer;

    // Paper-formant band: highpass strips rumble, lowpass tames hiss.
    const hp = ctx.createBiquadFilter();
    hp.type = "highpass";
    hp.frequency.value = 800;
    hp.Q.value = 0.5;

    const lp = ctx.createBiquadFilter();
    lp.type = "lowpass";
    // Slight per-call detune so consecutive flips don't sound mechanical.
    lp.frequency.value = 5500 + (Math.random() - 0.5) * 800;
    lp.Q.value = 0.5;

    const gain = ctx.createGain();
    gain.gain.value = volume;

    src.connect(hp).connect(lp).connect(gain).connect(ctx.destination);
    src.start();
  } catch {
    // silent fail in unsupported envs
  }
}
