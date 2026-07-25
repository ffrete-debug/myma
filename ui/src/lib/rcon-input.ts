// Pure input-line buffer logic for the RCON console.
//
// Extracted from RCONConsole so the keystroke handling can be unit-tested
// without spinning up xterm or a WebSocket. The terminal side (re-rendering
// the visible line, moving the caret) is intentionally NOT part of this
// module — the buffer is data-only, the rendering is the component's job.

export interface InputBuffer {
  text: string
  cursor: number
}

export interface ApplyResult {
  /** The new buffer state after applying the input event. */
  buffer: InputBuffer
  /**
   * A sentinel value indicating the user pressed Enter. The component uses
   * this to flush the line to the WebSocket. The returned buffer is empty
   * (cursor 0) in that case.
   */
  submitted?: string
}

/**
 * Apply a raw xterm `onData` payload to an input buffer.
 *
 * Supports:
 *  - Printable characters (insert at cursor)
 *  - Backspace (0x7F)
 *  - Enter (\r, 0x0D)
 *  - Arrow left / right (ESC [ C / ESC [ D)
 *  - Ctrl-A (home, 0x01), Ctrl-E (end, 0x05), Ctrl-U (clear line, 0x15)
 *
 * All other control sequences are ignored (the component decides whether to
 * swallow them at the terminal level).
 */
export function applyInputEvent(prev: InputBuffer, data: string): ApplyResult {
  if (data.length === 0) return { buffer: prev }

  const code = data.charCodeAt(0)

  // Enter
  if (code === 0x0d) {
    const submitted = prev.text
    return { buffer: { text: '', cursor: 0 }, submitted }
  }

  // Backspace
  if (code === 0x7f) {
    if (prev.cursor > 0) {
      const text = prev.text.slice(0, prev.cursor - 1) + prev.text.slice(prev.cursor)
      return { buffer: { text, cursor: prev.cursor - 1 } }
    }
    return { buffer: prev }
  }

  // Ctrl-A: home
  if (code === 0x01) {
    return { buffer: { text: prev.text, cursor: 0 } }
  }

  // Ctrl-E: end
  if (code === 0x05) {
    return { buffer: { text: prev.text, cursor: prev.text.length } }
  }

  // Ctrl-U: kill line
  if (code === 0x15) {
    return { buffer: { text: '', cursor: 0 } }
  }

  // Arrow keys: ESC [ <letter>
  if (code === 0x1b && data.length === 3) {
    const seq = data[2]
    if (seq === 'D' && prev.cursor > 0) {
      return { buffer: { text: prev.text, cursor: prev.cursor - 1 } }
    }
    if (seq === 'C' && prev.cursor < prev.text.length) {
      return { buffer: { text: prev.text, cursor: prev.cursor + 1 } }
    }
    // up/down (A/B) and anything else: no-op
    return { buffer: prev }
  }

  // Printable
  if (code >= 0x20) {
    const text = prev.text.slice(0, prev.cursor) + data + prev.text.slice(prev.cursor)
    return { buffer: { text, cursor: prev.cursor + data.length } }
  }

  // Any other control byte: ignore
  return { buffer: prev }
}
