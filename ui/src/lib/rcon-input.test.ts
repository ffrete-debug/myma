import { describe, it, expect } from 'vitest'
import { applyInputEvent, type InputBuffer } from './rcon-input'

const at = (text: string, cursor = text.length): InputBuffer => ({ text, cursor })

describe('applyInputEvent - printable input', () => {
  it('inserts at an empty cursor', () => {
    const r = applyInputEvent(at(''), 'a')
    expect(r.buffer).toEqual(at('a', 1))
    expect(r.submitted).toBeUndefined()
  })

  it('appends at end of line', () => {
    const r = applyInputEvent(at('ab'), 'c')
    expect(r.buffer).toEqual(at('abc', 3))
  })

  it('inserts in the middle of the line at cursor', () => {
    const r = applyInputEvent(at('ac', 1), 'b')
    expect(r.buffer).toEqual(at('abc', 2))
  })

  it('ingests multi-byte UTF-8 data as a single unit', () => {
    const r = applyInputEvent(at(''), 'ç')
    expect(r.buffer).toEqual(at('ç', 1))
  })
})

describe('applyInputEvent - Enter', () => {
  it('submits the current line and clears the buffer', () => {
    const r = applyInputEvent(at('status'), '\r')
    expect(r.submitted).toBe('status')
    expect(r.buffer).toEqual(at('', 0))
  })

  it('submits an empty line too (component decides to swallow)', () => {
    const r = applyInputEvent(at(''), '\r')
    expect(r.submitted).toBe('')
    expect(r.buffer).toEqual(at('', 0))
  })
})

describe('applyInputEvent - Backspace', () => {
  it('deletes the char before cursor', () => {
    const r = applyInputEvent(at('abc', 2), '\x7f')
    expect(r.buffer).toEqual(at('ac', 1))
  })

  it('is a no-op at column 0', () => {
    const r = applyInputEvent(at('abc', 0), '\x7f')
    expect(r.buffer).toEqual(at('abc', 0))
  })
})

describe('applyInputEvent - arrow keys', () => {
  it('moves left', () => {
    expect(applyInputEvent(at('abc', 3), '\x1b[D').buffer).toEqual(at('abc', 2))
  })

  it('does not move left past column 0', () => {
    expect(applyInputEvent(at('abc', 0), '\x1b[D').buffer).toEqual(at('abc', 0))
  })

  it('moves right', () => {
    expect(applyInputEvent(at('abc', 0), '\x1b[C').buffer).toEqual(at('abc', 1))
  })

  it('does not move right past end of line', () => {
    expect(applyInputEvent(at('abc', 3), '\x1b[C').buffer).toEqual(at('abc', 3))
  })

  it('ignores up/down arrows', () => {
    expect(applyInputEvent(at('abc', 1), '\x1b[A').buffer).toEqual(at('abc', 1))
    expect(applyInputEvent(at('abc', 1), '\x1b[B').buffer).toEqual(at('abc', 1))
  })
})

describe('applyInputEvent - Ctrl-A / Ctrl-E / Ctrl-U', () => {
  it('Ctrl-A jumps to start', () => {
    expect(applyInputEvent(at('abc', 2), '\x01').buffer).toEqual(at('abc', 0))
  })

  it('Ctrl-E jumps to end', () => {
    expect(applyInputEvent(at('abc', 0), '\x05').buffer).toEqual(at('abc', 3))
  })

  it('Ctrl-U clears the entire line', () => {
    expect(applyInputEvent(at('abc', 2), '\x15').buffer).toEqual(at('', 0))
  })
})

describe('applyInputEvent - unknown control bytes', () => {
  it('ignores raw ESC alone', () => {
    expect(applyInputEvent(at('abc', 1), '\x1b').buffer).toEqual(at('abc', 1))
  })

  it('ignores other C0 control codes below 0x20 (e.g. tab \\t=0x09)', () => {
    expect(applyInputEvent(at('abc', 1), '\t').buffer).toEqual(at('abc', 1))
  })

  it('handles empty input as a no-op', () => {
    expect(applyInputEvent(at('abc', 1), '').buffer).toEqual(at('abc', 1))
  })
})
