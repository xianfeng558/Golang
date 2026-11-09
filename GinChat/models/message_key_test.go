package models

import "testing"

// TestMsgKeyOrdersPair 保证聊天记录的键不看谁先开口。
// 9 和 2 无论谁是发送者，键都是 msg_2_9，这样双方打开会话读到的是同一份有序集合。
func TestMsgKeyOrdersPair(t *testing.T) {
	if MsgKey(9, 2) != "msg_2_9" {
		t.Fatalf("got %s", MsgKey(9, 2))
	}
	if MsgKey(2, 9) != MsgKey(9, 2) {
		t.Fatal("两个人的聊天记录必须是同一把钥匙")
	}
}
