package necore_test

import (
	"sync"
	"testing"

	"necore/ws"

	"github.com/gofiber/fiber/v2"
)

func TestWSHub_RegisterUnregisterStats(t *testing.T) {
	env := setupTestEnv(t)

	hub := &ws.Hub{Clients: make(map[string]*ws.Client)}

	hub.Register(&ws.Client{SessionID: "s1", Identifier: "bot-1", TokenID: 1, TokenName: "tk-1"})
	hub.Register(&ws.Client{SessionID: "s2", Identifier: "bot-2", TokenID: 2, TokenName: "tk-2"})

	clients, logs := hub.GetDashboardStats()
	if len(clients) != 2 {
		t.Fatalf("stats clients = %d, want 2", len(clients))
	}
	if len(logs) == 0 {
		t.Fatal("register should produce a log entry")
	}

	// 返回的副本可被安全修改，不影响 hub 内部状态
	clients[0].Identifier = "mutated"
	if got := hub.Clients["s1"].Identifier; got == "mutated" {
		t.Fatal("dashboard copy must not alias internal client")
	}

	// 注销后计数回落
	hub.Unregister("s1", "unit-test", false)
	clients, _ = hub.GetDashboardStats()
	if len(clients) != 1 {
		t.Fatalf("after unregister clients = %d, want 1", len(clients))
	}
	// 不存在的会话注销不 panic
	hub.Unregister("missing", "unit-test", false)

	_ = env
}

func TestWSHub_BroadcastWithNilConn(t *testing.T) {
	// 断连但未注销的客户端（Conn 为 nil）不应导致广播 panic
	hub := &ws.Hub{Clients: make(map[string]*ws.Client)}
	hub.Register(&ws.Client{SessionID: "s1", Identifier: "bot-1"})

	hub.Broadcast(fiber.Map{"event": "ping"})
	hub.BroadcastToSessions(fiber.Map{"event": "ping"}, []string{"s1"})
	hub.BroadcastToSessions(fiber.Map{"event": "ping"}, nil)
}

func TestWSHub_ConcurrentAccess(t *testing.T) {
	hub := &ws.Hub{Clients: make(map[string]*ws.Client)}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 26 {
				hub.Register(&ws.Client{SessionID: "s", Identifier: "bot"})
			}
			hub.GetDashboardStats()
			hub.Broadcast(fiber.Map{"event": "ping"})
		}()
	}
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.Unregister("s", "test", true)
		}()
	}
	wg.Wait()

	// 竞态检测器（go test -race）会在有数据竞争时失败。
	hub.GetDashboardStats()
}
