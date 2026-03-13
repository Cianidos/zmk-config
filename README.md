# ZMK config — Lily58 (nice!nano v2)

Miryoku-inspired layout with home row mods, mouse control, and a gaming mode.

## Layers

| # | Name | Activated by |
|---|------|-------------|
| 0 | Base | always |
| 1 | Nums | right thumb hold |
| 2 | Arrows | left inner thumb hold |
| 3 | Mouse | left outer thumb hold |
| 4–7 | Mouse Slow/Medium/Fast/Faster | bottom-row hold while in Mouse |
| 8 | Gaming | `&tog 8` (Z-column key) |

## Base layout (Colemak-DH)

Home row mods: `GUI / ALT / CTRL / SHIFT` on both hands (home row + bottom row).
Hold-tap with `hold-trigger-key-positions` — each hand's mods only fire on opposite-hand keys.

## Mouse layer

- Right hand: move (IJKL-style) + scroll
- Left hand: modifiers (no speed change)
- Bottom-left row: hold `&mo 4/5/6/7` to change cursor/scroll speed while staying in mouse mode

## Bluetooth / output

BT controls on the Arrows layer (top row, right side): `OUT_TOG`, `BT_SEL 0–3`, `BT_CLR`.

## Flashing

1. Trigger a GitHub Actions build (push a change or re-run the workflow)
2. Download the `firmware` artifact
3. Double-tap RESET on each half, drag the matching `.uf2` onto the `NICENANO` drive

If the keyboard misbehaves after reflashing, flash `settings_reset.uf2` to both halves first, then reflash normally.

## Keymap editor

[nickcoutsos.github.io/keymap-editor](https://nickcoutsos.github.io/keymap-editor/) — authorize your GitHub fork for a visual editor.
