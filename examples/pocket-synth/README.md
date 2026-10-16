# Pocket Synthesizer

A retro hardware-inspired mini-synthesizer with real-time waveform oscilloscope, 16-step arpeggiator, and procedural 16-bit audio engine.

```sh
go run ./examples/pocket-synth          # desktop window, 540x680
go run ./examples/pocket-synth -web     # browser preview on 127.0.0.1:8141
```

## Features

- **Procedural Audio Engine**: Real-time 16-bit PCM waveform generation at 48 kHz wrapped in RIFF/WAVE streams played through Ebiten audio. Gracefully falls back in headless tests and web mode.
- **Waveform Presets**: Four distinct classic synth waveforms:
  - `8-BIT SQUARE`: Snappy chiptune pulse wave
  - `ACID SAW`: Sharp, harmonic-rich sawtooth wave
  - `WARM TRIANGLE`: Soft, mellow triangular wave
  - `PURE SINE`: Clean fundamental sine tone
- **16-Step Arpeggiator**: Built-in rhythmic sequencer at 120 BPM with pulsing beat LED and animated oscilloscope display.
- **Octave Shifting**: Shift keyboard octaves across range 2 to 6.
- **Oscilloscope**: Animated 16-bar frequency display reflecting waveform shapes and active notes.

## Controls

### Keyboard Mappings

| Key | Function |
| --- | --- |
| `A`, `S`, `D`, `F`, `G`, `H`, `J`, `K` | White keys: C, D, E, F, G, A, B, C+ |
| `W`, `E`, `T`, `Y`, `U` | Black keys: C#, D#, F#, G#, A# |
| `Z` / `X` | Octave Down / Octave Up |
| `Tab` | Cycle Waveform Preset |
| `Space` / `P` | Toggle Arpeggiator Play / Pause |

### Mouse & Touch

- Tap or click any piano key to trigger note playback.
- Click `PRESET` to cycle through available waveforms.
- Click `OCT -` / `OCT +` to change pitch register.
- Click `PLAY` / `PAUSE` to start or pause the sequencer.

## Testing

```sh
go test -p 1 -v ./examples/pocket-synth/...
```
