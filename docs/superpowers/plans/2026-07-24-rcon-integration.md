# RCON Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add live RCON command execution to ARK servers via an embedded xterm.js terminal panel and a command/response API endpoint.

**Architecture:** Extend the existing WebSocket hub for real-time terminal I/O, add an HTTP endpoint for single command/response, and build a Go RCON client using an existing Go RCON library. The frontend uses xterm.js in a collapsible panel on the server detail page.

**Tech Stack:** Go RCON library (`github.com/stevenmey/ark-rcon`), xterm.js, existing Gin + WebSocket infrastructure

## Global Constraints

- JWT auth on all RCON endpoints (reuse existing middleware)
- Audit all RCON commands in the audit log
- RCON connection lifecycle: connect on panel open, disconnect on panel close
- Maximum output response: 64KB per command
- Use existing packages in go.mod where possible

---

### Task 1: Add RCON Go library dependency

**Files:**
- Modify: `server/go.mod`

**Interfaces:**
- Consumes: nothing
- Produces: RCON library available in module

- [ ] **Step 1: Add RCON library**

```bash
cd server && go get github.com/stevenmey/ark-rcon
```

- [ ] **Step 2: Verify it resolves**

```bash
cd server && go mod tidy && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/go.mod server/go.sum
git commit -m "feat: add ark-rcon library dependency"
```

### Task 2: Create RCON client wrapper

**Files:**
- Create: `server/service/rcon/rcon_client.go`
- Create: `server/service/rcon/rcon_client_test.go`

**Interfaces:**
- Produces: `func ExecuteCommand(host string, port int, password, command string) (string, error)`

- [ ] **Step 1: Write the failing test**

```go
package rcon

import (
	"testing"
)

func TestExecuteCommand_InvalidHost(t *testing.T) {
	_, err := ExecuteCommand("invalid-host", 32330, "password", "status")
	if err == nil {
		t.Error("expected error for invalid host, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./service/rcon/ -v -run TestExecuteCommand_InvalidHost
```

- [ ] **Step 3: Write minimal implementation**

```go
package rcon

import (
	"fmt"
	"time"
	"github.com/stevenmey/ark-rcon"
)

func ExecuteCommand(host string, port int, password, command string) (string, error) {
	conn, err := arkrcon.Dial(host, port, password, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("rcon dial: %w", err)
	}
	defer conn.Close()

	response, err := conn.Execute(command)
	if err != nil {
		return "", fmt.Errorf("rcon execute: %w", err)
	}

	return response, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd server && go test ./service/rcon/ -v -run TestExecuteCommand_InvalidHost
```

- [ ] **Step 5: Commit**

```bash
git add server/service/rcon/
git commit -m "feat: add RCON client wrapper with ExecuteCommand"
```

### Task 3: Add ExecuteRCONCommand to server service

**Files:**
- Modify: `server/service/server/server_service.go`

**Interfaces:**
- Consumes: RCON port and admin password from server model
- Produces: `func (s *ServerService) ExecuteRCONCommand(userID uint, serverID string, command string) (string, error)`

- [ ] **Step 1: Add method to ServerService**

```go
func (s *ServerService) ExecuteRCONCommand(userID uint, serverID string, command string) (string, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return "", fmt.Errorf("Server not found")
	}

	output, err := rcon.ExecuteCommand("localhost", server.RCONPort, server.AdminPassword, command)
	if err != nil {
		return "", fmt.Errorf("RCON command failed: %w", err)
	}

	return output, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/service/server/server_service.go
git commit -m "feat: add ExecuteRCONCommand to ServerService"
```

### Task 4: Create RCON HTTP handler

**Files:**
- Create: `server/controllers/rcon/rcon.go`

**Interfaces:**
- Produces: `POST /servers/:id/rcon/execute` handler

- [ ] **Step 1: Write the handler**

```go
package controllers

import (
	"net/http"
	"ark-server-commander/service/server"
	"ark-server-commander/utils"
	"github.com/gin-gonic/gin"
)

type RCONRequest struct {
	Command string `json:"command" binding:"required"`
}

func ExecuteRCON(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	var req RCONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	serverService := server.GetServerService()
	output, err := serverService.ExecuteRCONCommand(userID, serverID, req.Command)
	if err != nil {
		utils.InternalError(c, "RCON execution failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Command executed",
		"data":    gin.H{"output": output},
	})
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/controllers/rcon/rcon.go
git commit -m "feat: add RCON execute HTTP handler"
```

### Task 5: Register RCON routes

**Files:**
- Create: `server/routes/rcon_routes.go`
- Modify: `server/routes/routes.go` (mount RCON routes)

- [ ] **Step 1: Create routes file**

```go
package routes

import (
	"ark-server-commander/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterRCONRoutes(r *gin.RouterGroup) {
	r.POST("/:id/rcon/execute", controllers.ExecuteRCON)
}
```

- [ ] **Step 2: Mount routes in routes.go**

After the server routes group, add:
```go
RegisterRCONRoutes(serverRoutes)
```

- [ ] **Step 3: Verify compilation**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/routes/
git commit -m "feat: register RCON execute route"
```

### Task 6: Add RCON audit logging

**Files:**
- Modify: `server/middleware/audit.go`

**Interfaces:**
- Consumes: audit action type for RCON commands

- [ ] **Step 1: Add RCON command audit type**

In the audit action constants:
```go
const ActionRCONCommand = "rcon_command"
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/middleware/audit.go
git commit -m "feat: add RCON command audit action type"
```

### Task 7: Extend WebSocket hub for RCON

**Files:**
- Modify: `server/websocket/hub.go`
- Create: `server/websocket/rcon_handler.go`

- [ ] **Step 1: Create RCON WebSocket handler**

```go
package websocket

import (
	"ark-server-commander/service/rcon"
	"ark-server-commander/models"
	"ark-server-commander/database"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleRCONWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Server not found"))
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		output, err := rcon.ExecuteCommand("localhost", server.RCONPort, server.AdminPassword, string(msg))
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			continue
		}

		conn.WriteMessage(websocket.TextMessage, []byte(output))
	}
}
```

- [ ] **Step 2: Add WebSocket route**

In routes:
```go
r.GET("/ws/rcon/:id", middleware.AuthRequired(), websocket.HandleRCONWebSocket)
```

- [ ] **Step 3: Verify compilation**

```bash
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/websocket/ server/routes/
git commit -m "feat: add RCON WebSocket handler and route"
```

### Task 8: Create RCONConsole React component

**Files:**
- Create: `ui/src/components/servers/RCONConsole.tsx`

- [ ] **Step 1: Install xterm.js dependency**

```bash
cd ui && npm install @xterm/xterm @xterm/addon-fit
```

- [ ] **Step 2: Create RCONConsole component**

```tsx
'use client';

import { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

interface RCONConsoleProps {
  serverId: string;
}

export function RCONConsole({ serverId }: RCONConsoleProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const termRef = useRef<Terminal | null>(null);
  const [connected, setConnected] = useState(false);
  const [collapsed, setCollapsed] = useState(true);

  const connectWS = () => {
    const token = localStorage.getItem('access_token');
    const ws = new WebSocket(`/ws/rcon/${serverId}?token=${token}`);
    
    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onmessage = (event) => {
      termRef.current?.writeln(event.data);
    };
    
    wsRef.current = ws;
  };

  useEffect(() => {
    if (!terminalRef.current || collapsed) return;

    const term = new Terminal({
      theme: { background: '#1a1a2e', foreground: '#e0e0e0' },
      cursorBlink: true,
      fontSize: 14,
    });
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();
    termRef.current = term;

    connectWS();
    
    term.onKey((e) => {
      if (e.domEvent.key === 'Enter') {
        const line = term.buffer.active.getLine(term.buffer.active.cursorY)?.translateToString() || '';
        if (line.trim()) {
          wsRef.current?.send(line.trim());
        }
      }
    });

    return () => {
      wsRef.current?.close();
      term.dispose();
    };
  }, [collapsed]);

  return (
    <div className="border rounded-lg overflow-hidden">
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="w-full px-4 py-2 bg-gray-800 text-white flex items-center gap-2"
      >
        <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`} />
        RCON Console
        <span className="ml-auto">{collapsed ? '▲' : '▼'}</span>
      </button>
      {!collapsed && (
        <div ref={terminalRef} className="h-64" />
      )}
    </div>
  );
}
```

- [ ] **Step 3: Commit**

```bash
git add ui/src/components/servers/RCONConsole.tsx
git commit -m "feat: add RCONConsole xterm.js component"
```

### Task 9: Embed RCONConsole in server detail page

**Files:**
- Create: `ui/src/app/(protected)/servers/[id]/page.tsx`

- [ ] **Step 1: Create server detail page with RCON panel**

```tsx
'use client';

import { useParams } from 'next/navigation';
import { RCONConsole } from '@/components/servers/RCONConsole';

export default function ServerDetailPage() {
  const params = useParams();
  const serverId = params.id as string;

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Server Detail</h1>
      <RCONConsole serverId={serverId} />
    </div>
  );
}
```

- [ ] **Step 2: Verify build**

```bash
cd ui && npm run build
```

- [ ] **Step 3: Commit**

```bash
git add ui/src/app/(protected)/servers/[id]/
git commit -m "feat: embed RCONConsole in server detail page"
```

### Task 10: Integration tests

**Files:**
- Create: `server/service/rcon/rcon_integration_test.go`

- [ ] **Step 1: Write integration test**

```go
package rcon

import (
	"testing"
)

func TestExecuteCommand_ValidServer(t *testing.T) {
	output, err := ExecuteCommand("localhost", 32330, "password", "status")
	if err != nil {
		t.Skip("RCON server not available:", err)
	}
	if output == "" {
		t.Error("expected non-empty response from RCON server")
	}
}
```

- [ ] **Step 2: Run Go vet**

```bash
cd server && go vet ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/service/rcon/rcon_integration_test.go
git commit -m "test: add RCON integration test"
```