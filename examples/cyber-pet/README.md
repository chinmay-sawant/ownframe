# Cyber Pet companion

Virtual companion in a cyberpunk handheld frame with dynamic expressions, vital decay, real-time animation, and interactive care actions.

```sh
go run ./examples/cyber-pet          # native window (520x640)
go run ./examples/cyber-pet -web     # browser preview on 127.0.0.1:8142
```

## Features

- **Pet Vitals & State Machine**:
  - `Energy` (0–100): Powers activities; recovers while sleeping.
  - `Hunger` (0–100): Food fullness; replenished by feeding.
  - `Happiness` (0–100): Boosted through play, petting, and chats.
  - `Moods`: Transitions between `happy`, `hungry`, `tired`, `sleeping`, `sad`, `playful`, and `purring`.
  - **Decay Engine**: Vitals decay gradually over time, prompting player care.

- **Interactive Actions**:
  - **FEED** (`1`): Satisfies hunger (+25) and cheers up the pet (+5).
  - **PLAY** (`2`): Costs energy (-15) and hunger (-10) to boost happiness (+20).
  - **PET** (`3`, `Space`, or Face Click/Drag): Triggers purring reaction with floating hearts.
  - **SLEEP** (`4`): Toggles sleep mode for rapid energy recovery (+6/sec).
  - **TALK** (`5`): Displays cyber dialogue responses and boosts affection.
  - **Petting Gesture**: Click or drag across the pet face to trigger purring and heart particles.

- **Display List Animation Loop**:
  - Retained display operations animate inside `Page.SetTick` without triggering template re-parses.
  - **Eye Blinks**: Natural periodic blinking pulse while awake; closed lids when sleeping.
  - **Mouth Oscillation**: Animated speech and chew oscillations tailored to active behavior.
  - **Floating Hearts**: Floats upward with alpha fading whenever the companion is petted.

## Testing

Run unit and headless integration tests:

```sh
go test -p 1 ./examples/cyber-pet/...
```
