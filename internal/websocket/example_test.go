package websocket

import (
	"testing"
	"time"
)

// TestHub 测试 Hub 基本功能
func TestHub(t *testing.T) {
	// 创建 Hub
	hub := NewHub(nil)

	// 启动 Hub
	hub.Start()
	defer hub.Stop()

	// 验证 Hub 状态
	if !hub.running {
		t.Error("Hub 应该处于运行状态")
	}

	// 获取统计信息
	stats := hub.GetStats()
	if stats["total_clients"].(int) != 0 {
		t.Error("初始客户端数量应该为 0")
	}

	if stats["total_rooms"].(int) != 0 {
		t.Error("初始房间数量应该为 0")
	}
}

// TestBroadcastMessage 测试消息广播
func TestBroadcastMessage(t *testing.T) {
	hub := NewHub(nil)
	hub.Start()
	defer hub.Stop()

	// 测试广播到不存在的房间
	hub.BroadcastToRoom("test-room", "test-event", map[string]interface{}{
		"message": "hello",
	})

	// 应该不会崩溃
	time.Sleep(100 * time.Millisecond)
}

// TestRoomManagement 测试房间管理
func TestRoomManagement(t *testing.T) {
	hub := NewHub(nil)
	hub.Start()
	defer hub.Stop()

	// 获取不存在房间的用户列表
	users := hub.GetRoomUsers("non-existent-room")
	if len(users) != 0 {
		t.Error("不存在的房间应该返回空用户列表")
	}
}
