package vendors

import "testing"

func TestAnamAvatarSerializesAvatarID(t *testing.T) {
	config := NewAnamAvatar(AnamAvatarOptions{
		APIKey:   "anam-key",
		AvatarID: "avatar-1",
	}).ToConfig()

	if config["vendor"] != "anam" {
		t.Fatalf("unexpected vendor: %v", config["vendor"])
	}

	params := config["params"].(map[string]interface{})
	if params["api_key"] != "anam-key" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["avatar_id"] != "avatar-1" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if _, ok := params["persona_id"]; ok {
		t.Fatalf("unexpected legacy param present: %#v", params)
	}
}
