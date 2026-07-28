"use client";

import { useEffect, useRef, useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import Cookies from 'js-cookie';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { ChevronDown, ChevronUp, Terminal as TerminalIcon } from 'lucide-react';
import { applyInputEvent, type InputBuffer } from '@/lib/rcon-input';
import { buildWebSocketUrl } from '@/lib/ws-url';

interface RCONConsoleProps {
  serverId: string;
  serverStatus?: 'running' | 'stopped' | 'starting' | 'stopping' | 'restarting';
}

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'error';

const PROMPT = '> ';

// Reconnect backoff. A server that is down for maintenance used to be retried
// every 1.5s for as long as the panel stayed open; now the delay doubles up to
// a ceiling so the browser and the backend both stop being hammered.
const RECONNECT_BASE_MS = 1500;
const RECONNECT_MAX_MS = 30000;
const RECONNECT_MAX_ATTEMPTS = 8;

// Matches one ESC-introduced sequence:
//   - CSI:                ESC [ <params> <intermediates> <final>
//   - string sequences:   ESC ] / P / X / ^ / _ ... terminated by BEL or ST
//   - short escapes:      ESC <single printable>
const ANSI_SEQUENCE =
  /\x1b(?:\[[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]|[\]PX^_][\s\S]*?(?:\x07|\x1b\\|$)|[\x20-\x7e])/g;

// SGR ("select graphic rendition" - colour, bold, reset) is the one escape
// family that cannot move the cursor, repaint the screen or attach an OSC-8
// hyperlink, so it is the only one allowed through.
const SGR_SEQUENCE = /^\x1b\[[0-9;]*m$/;

// C0 controls other than tab/newline/carriage-return, plus DEL and the whole C1
// range (0x80-0x9f), which some terminals decode as single-byte equivalents of
// the ESC sequences stripped above.
const UNSAFE_CONTROL_CHARS = /[\x00-\x08\x0b\x0c\x0e-\x1f\x7f-\x9f]/g;

/**
 * Neutralise terminal control sequences in RCON output before writing it.
 *
 * The output is attacker-influenceable (any player can put text into a game
 * response), and the terminal runs with `allowProposedApi`. This is not DOM
 * XSS - xterm renders to a canvas - but unfiltered escapes still let remote
 * text reposition the cursor, erase or rewrite lines already on screen and
 * emit OSC-8 hyperlinks, i.e. spoof console output the operator then trusts.
 *
 * The backend forwards raw RCON payloads verbatim (see
 * `server/websocket/rcon_handler.go`) and adds no colouring of its own, but
 * SGR is preserved anyway since it is inert.
 */
function sanitizeTerminalOutput(input: string): string {
  let result = '';
  let index = 0;

  ANSI_SEQUENCE.lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = ANSI_SEQUENCE.exec(input)) !== null) {
    result += input.slice(index, match.index).replace(UNSAFE_CONTROL_CHARS, '');
    if (SGR_SEQUENCE.test(match[0])) {
      result += match[0];
    }
    index = match.index + match[0].length;
  }
  result += input.slice(index).replace(UNSAFE_CONTROL_CHARS, '');

  return result;
}

export function RCONConsole({ serverId, serverStatus }: RCONConsoleProps) {
  const t = useTranslations('servers.rcon');
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const inputBufferRef = useRef<string>('');
  const cursorPosRef = useRef<number>(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isClosingRef = useRef(false);

  const [collapsed, setCollapsed] = useState(true);
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');

  const isServerReachable = serverStatus === 'running' || serverStatus === undefined;

  // The terminal is created exactly once per expand and lives across renders,
  // so anything it closes over is read through a ref. Keeping `t` out of the
  // callback dependencies means a new translator identity cannot tear down a
  // live session (and lets the effects below list honest dependencies).
  const tRef = useRef(t);
  tRef.current = t;
  const isServerReachableRef = useRef(isServerReachable);
  isServerReachableRef.current = isServerReachable;

  const writePrompt = useCallback((term: Terminal) => {
    term.write('\r\n' + PROMPT);
  }, []);

  const refreshInputLine = useCallback((term: Terminal) => {
    const buf = inputBufferRef.current;
    let line = PROMPT;
    if (buf.length === 0) {
      term.write('\r' + line);
      return;
    }
    const cursor = cursorPosRef.current;
    line += buf;
    term.write('\r' + line);
    if (cursor < buf.length) {
      // Move cursor back to logical position relative to prompt.
      const backspaces = buf.length - cursor;
      term.write(`\x1b[${backspaces}D`);
    }
  }, []);

  const connectWS = useCallback(() => {
    if (isClosingRef.current) return;

    // A pending reconnect is now superseded by this attempt.
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }

    const previous = wsRef.current;
    if (previous) {
      // Detach the old socket before closing it. Its `onclose` fires
      // asynchronously, i.e. after `wsRef.current` already points at the new
      // socket, so leaving the handlers attached would let the dead socket null
      // out the live reference and queue a second, duplicate connection.
      previous.onopen = null;
      previous.onmessage = null;
      previous.onerror = null;
      previous.onclose = null;
      previous.close();
      wsRef.current = null;
    }

    const token = Cookies.get('auth-token');
    if (!token) {
      setConnectionState('disconnected');
      termRef.current?.writeln('\r\n' + tRef.current('connectionError'));
      return;
    }

    const url = buildWebSocketUrl(`/api/ws/rcon/${serverId}`, token);
    setConnectionState('connecting');
    const ws = new WebSocket(url);
    wsRef.current = ws;

    // Every handler bails out unless this socket is still the current one, so a
    // socket that has been replaced can neither write to the terminal, move the
    // connection state, clear the ref nor schedule a reconnect.
    const isCurrent = () => wsRef.current === ws;

    ws.onopen = () => {
      if (!isCurrent()) return;
      reconnectAttemptsRef.current = 0;
      setConnectionState('connected');
      const term = termRef.current;
      if (!term) return;
      term.writeln('\r\n[' + new Date().toLocaleTimeString() + '] ' + tRef.current('connected'));
      writePrompt(term);
    };

    ws.onmessage = (event) => {
      if (!isCurrent()) return;
      const term = termRef.current;
      if (!term) return;
      if (typeof event.data !== 'string') {
        term.writeln('\r\n' + tRef.current('invalidResponse'));
        writePrompt(term);
        return;
      }
      const text = sanitizeTerminalOutput(event.data).replace(/\n/g, '\r\n');
      term.write('\r\n' + text);
      writePrompt(term);
    };

    ws.onerror = () => {
      if (!isCurrent()) return;
      setConnectionState('error');
      termRef.current?.writeln('\r\n' + tRef.current('connectionError'));
    };

    ws.onclose = () => {
      if (!isCurrent()) return;
      wsRef.current = null;
      setConnectionState((prev) => (prev === 'connected' ? 'disconnected' : prev));
      if (isClosingRef.current) return;

      // Exponential backoff, capped in both delay and attempts, to survive
      // transient network drops without retrying a dead server forever.
      // Once the attempts are exhausted the panel stays in `error` until it is
      // collapsed and re-expanded, rather than dialling a dead server forever.
      const attempt = reconnectAttemptsRef.current;
      if (attempt >= RECONNECT_MAX_ATTEMPTS) {
        setConnectionState('error');
        termRef.current?.writeln('\r\n' + tRef.current('connectionError'));
        return;
      }
      reconnectAttemptsRef.current = attempt + 1;
      const delay = Math.min(RECONNECT_BASE_MS * 2 ** attempt, RECONNECT_MAX_MS);
      reconnectTimerRef.current = setTimeout(() => {
        reconnectTimerRef.current = null;
        if (!isClosingRef.current) connectWS();
      }, delay);
    };
  }, [serverId, writePrompt]);

  // Initialize terminal once when the panel is expanded.
  useEffect(() => {
    if (collapsed) return;
    const el = containerRef.current;
    if (!el) return;
    isClosingRef.current = false;
    reconnectAttemptsRef.current = 0;

    const term = new Terminal({
      theme: { background: '#0b0e14', foreground: '#cdd6e0', cursor: '#cdd6e0' },
      cursorBlink: true,
      fontSize: 13,
      fontFamily: 'ui-monospace, Menlo, Monaco, "Cascadia Mono", "Courier New", monospace',
      scrollback: 1000,
      convertEol: true,
      // `disableStdin` must stay false: xterm drops every key event when stdin
      // is disabled, so the `onData` handler below would never fire and the
      // console would be unusable for typing commands.
      disableStdin: false,
      allowProposedApi: true,
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(el);
    fit.fit();
    termRef.current = term;
    fitRef.current = fit;

    term.writeln(tRef.current('connectingHint'));
    if (!isServerReachableRef.current) {
      term.writeln('\r\n' + tRef.current('serverOffline'));
      setConnectionState('disconnected');
    } else {
      connectWS();
    }

    const handleResize = () => {
      try { fit.fit(); } catch { /* ignore until container has layout */ }
    };
    window.addEventListener('resize', handleResize);
    const ro = new ResizeObserver(handleResize);
    ro.observe(el);

    term.onData((data) => {
      const tic = termRef.current;
      if (!tic) return;

      const prevInput: InputBuffer = {
        text: inputBufferRef.current,
        cursor: cursorPosRef.current,
      };
      const result = applyInputEvent(prevInput, data);

      // Detect a cursor-delta for arrow / home / end movements so we can
      // emit the proper ANSI cursor-move sequence instead of re-rendering
      // the whole line (avoiding flicker at the caret).
      if (result.buffer.cursor !== prevInput.cursor && result.buffer.text === prevInput.text) {
        const delta = result.buffer.cursor - prevInput.cursor;
        if (delta < 0) {
          tic.write(`\x1b[${-delta}D`)
        } else if (delta > 0) {
          tic.write(`\x1b[${delta}C`)
        }
      } else if (result.buffer.text !== prevInput.text || result.buffer.cursor !== prevInput.cursor) {
        // Text changed — redraw the visible line and clear any trailing char.
        refreshInputLine(tic)
        if (result.buffer.text.length < prevInput.text.length) {
          tic.write('\x1b[K')
        }
      }

      // Enter: flush the line to the WS. The input buffer is already cleared.
      if (typeof result.submitted !== 'undefined') {
        const line = result.submitted
        tic.write('\r\n')
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
          if (line.trim().length === 0) {
            writePrompt(tic)
            return
          }
          try {
            wsRef.current.send(line)
          } catch {
            tic.writeln(tRef.current('sendFailed'))
          }
        } else {
          tic.writeln(tRef.current('connectionError'))
          writePrompt(tic)
        }
      }

      inputBufferRef.current = result.buffer.text
      cursorPosRef.current = result.buffer.cursor
    });

    return () => {
      isClosingRef.current = true;
      window.removeEventListener('resize', handleResize);
      ro.disconnect();
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      reconnectAttemptsRef.current = 0;
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      term.dispose();
      termRef.current = null;
      fitRef.current = null;
      inputBufferRef.current = '';
      cursorPosRef.current = 0;
    };
  }, [collapsed, connectWS, writePrompt, refreshInputLine]);

  // Reconnect when the server becomes reachable after being offline.
  useEffect(() => {
    if (collapsed) return;
    if (
      serverStatus === 'running' &&
      connectionState === 'disconnected' &&
      wsRef.current === null &&
      // Do not race the backoff timer: if a reconnect is already scheduled,
      // let it run instead of dialling immediately.
      reconnectTimerRef.current === null
    ) {
      reconnectAttemptsRef.current = 0;
      connectWS();
    }
  }, [collapsed, connectionState, serverStatus, connectWS]);

  const dotColor =
    connectionState === 'connected' ? 'bg-emerald-500' :
    connectionState === 'connecting' ? 'bg-amber-500 animate-pulse' :
    connectionState === 'error' ? 'bg-red-500' :
    'bg-slate-400';

  const statusLabel =
    connectionState === 'connected' ? t('connected') :
    connectionState === 'connecting' ? t('connecting') :
    connectionState === 'error' ? t('connectionError') :
    t('disconnected');

  return (
    <div className="border border-border rounded-lg overflow-hidden bg-card">
      <button
        type="button"
        onClick={() => setCollapsed((v) => !v)}
        className="w-full px-4 py-2.5 flex items-center gap-2 bg-muted/40 hover:bg-muted/70 transition-colors"
        aria-expanded={!collapsed}
        aria-label={collapsed ? t('expand') : t('collapse')}
      >
        <TerminalIcon className="h-4 w-4 text-primary" />
        <span className="text-sm font-medium text-foreground">{t('console')}</span>
        <span className={`ml-2 h-2 w-2 rounded-full inline-block ${dotColor}`} data-state={connectionState} />
        <span className="text-xs text-muted-foreground">{statusLabel}</span>
        <span className="ml-auto">
          {collapsed ? <ChevronDown className="h-4 w-4" /> : <ChevronUp className="h-4 w-4" />}
        </span>
      </button>
      {!collapsed && (
        <div className="relative">
          <div ref={containerRef} className="h-72 w-full bg-[#0b0e14]" />
          {!isServerReachable && (
            <div className="absolute inset-0 bg-background/70 flex items-center justify-center pointer-events-none">
              <span className="text-sm text-muted-foreground">{t('serverOffline')}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
