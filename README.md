# ZMK config — Lily58 (nice!nano v2)

Miryoku-inspired layout with home row mods, Russian phonetic layer, mouse control, and a gaming mode.

## Layers

| # | Name | Activated by |
|---|------|-------------|
| 0 | Base | always |
| 1 | Ru | `&layer_ru` (outer-left middle-row key on Base) |
| 2 | Nums | right thumb hold |
| 3 | Arrows | left inner thumb hold |
| 4 | Mouse | left outer thumb hold |
| 5–8 | Mouse Slow/Medium/Fast/Faster | bottom-row hold while in Mouse |
| 9 | Gaming | `&tog GAMING` (Z-column key) |

## Base layout (Colemak-DH)

Home row mods: `GUI / ALT / CTRL / SHIFT` on both hands (home row + bottom row).
Hold-tap with `hold-trigger-key-positions` — each hand's mods only fire on opposite-hand keys.

Outer-left column: top = `CapsLock`, middle = `&layer_ru` (switches to Russian layer).

## Russian layer

Colemak-DH phonetic mapping for ЙЦУКЕН input method (OS must be switched to Russian).

- `&layer_ru` on Base activates the layer and sends CapsLock to toggle the OS input method.
- `&layer_en` on Ru switches back to Base and sends CapsLock to switch OS back to English.
- Punctuation (`,` `.` `/` `?` `'` `"`) uses native ЙЦУКЕН keycodes — no latency.
- `en_key` macro temporarily flips to English for a single keypress (used for `'`).
- Right thumb NUM hold uses `lt_lang` / `mo_lang` to toggle OS language while the layer is held.

## Mouse layer

- Right hand: move (IJKL-style) + scroll
- Left hand: modifiers (no speed change)
- Bottom-left row: hold `&mo MS_SLOW/MS_MED/MS_FAST/MS_FASTER` to change cursor/scroll speed while staying in mouse mode

## Bluetooth / output

BT controls on the Arrows layer (top row, right side): `OUT_TOG`, `BT_SEL 0–3`, `BT_CLR`.

## Flashing

### Manual (drag-and-drop)

1. Trigger a GitHub Actions build (push a change or re-run the workflow)
2. Download the `firmware` artifact
3. Double-tap RESET on each half, drag the matching `.uf2` onto the `NICENANO` drive

If the keyboard misbehaves after reflashing, flash `settings_reset.uf2` to both halves first, then reflash normally.

### Automated (`scripts/flash-zmk.go`)

```
go run scripts/flash-zmk.go [flags] <firmware-dir-or.zip>
```

Flashes in sequence: reset → left → reset → right, waiting for the drive to appear/disappear at each step.

```
Flags:
  -mount <path>   exact mount path of the keyboard drive (skips auto-detection)
  -t <ms>         poll interval in milliseconds (default 150)
  -q              quiet mode
```

Auto-detection by OS:
- **Linux** — `/proc/mounts`, `/media/<user>/NICENANO`, `/run/media/<user>/NICENANO`
- **macOS** — `/Volumes/NICENANO`
- **Windows** — drive letters D–Z (reads `INFO_UF2.TXT` to confirm nice!nano)

## Keymap editor

[nickcoutsos.github.io/keymap-editor](https://nickcoutsos.github.io/keymap-editor/) — authorize your GitHub fork for a visual editor.
