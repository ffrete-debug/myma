"use client";

import { useEffect, useRef, useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import Cookies from 'js-cookie';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { ChevronDown, ChevronUp, Terminal as TerminalIcon } from 'lucide-react';
import { applyInputEvent, type InputBuffer } from '@/lib/rcon-input';

interface RCONConsoleProps {
  serverId: string;
  serverStatus?: 'running' | 'stopped' | 'starting' | 'stopping' | 'restarting';
}

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'error';

const PROMPT = '> ';

export function RCONConsole({ serverId, serverStatus }: RCONConsoleProps) {
  const t = useTranslations('servers.rcon');
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const inputBufferRef = useRef<string>('');
  const cursorPosRef = useRef<number>(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isClosingRef = useRef(false);

  const [collapsed, setCollapsed] = useState(true);
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');

  const isServerReachable = serverStatus === 'running' || serverStatus === undefined;

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
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    const token = Cookies.get('auth-token');
    if (!token) {
      setConnectionState('disconnected');
      termRef.current?.writeln('\r\n' + t('connectionError'));
      return;
    }

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const url = `${proto}://${window.location.host}/api/ws/rcon/${serverId}?token=${encodeURIComponent(token)}`;
    setConnectionState('connecting');
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnectionState('connected');
      termRef.current?.writeln('\r\n[' + new Date().toLocaleTimeString() + '] ' + t('connected'));
      writePrompt(termRef.current!);
    };

    ws.onmessage = (event) => {
      if (typeof event.data !== 'string') {
        termRef.current?.writeln('\r\n' + t('invalidResponse'));
        writePrompt(termRef.current!);
        return;
      }
      const text = event.data.replace(/\n/g, '\r\n');
      termRef.current?.write('\r\n' + text);
      writePrompt(termRef.current!);
    };

    ws.onerror = () => {
      setConnectionState('error');
      termRef.current?.writeln('\r\n' + t('connectionError'));
    };

    ws.onclose = () => {
      wsRef.current = null;
      setConnectionState((prev) => (prev === 'connected' ? 'disconnected' : prev));
      if (isClosingRef.current) return;
      // Light retry to survive transient network drops on first attempt.
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = setTimeout(() => {
        if (!isClosingRef.current) connectWS();
      }, 1500);
    };
  }, [serverId, t, writePrompt]);

  // Initialize terminal once when the panel is expanded.
  useEffect(() => {
    if (collapsed) return;
    const el = containerRef.current;
    if (!el) return;
    isClosingRef.current = false;

    const term = new Terminal({
      theme: { background: '#0b0e14', foreground: '#cdd6e0', cursor: '#cdd6e0' },
      cursorBlink: true,
      fontSize: 13,
      fontFamily: 'ui-monospace, Menlo, Monaco, "Cascadia Mono", "Courier New", monospace',
      scrollback: 1000,
      convertEol: true,
      disableStdin: true,
      allowProposedApi: true,
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(el);
    fit.fit();
    termRef.current = term;
    fitRef.current = fit;

    term.writeln(t('connectingHint'));
    if (!isServerReachable) {
      term.writeln('\r\n' + t('serverOffline'));
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
            tic.writeln(t('sendFailed'))
          }
        } else {
          tic.writeln(t('connectionError'))
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
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [collapsed]);

  // Reconnect when server becomes reachable after being offline.
  useEffect(() => {
    if (collapsed) return;
    if (serverStatus === 'running' && connectionState === 'disconnected' && wsRef.current === null) {
      connectWS();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [serverStatus]);

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
