package vendors

import "testing"

func TestSensetimeAvatarOmitsEmptySceneList(t *testing.T) {
	config := NewSensetimeAvatar(SensetimeAvatarOptions{
		AgoraUID: "2001",
		AppID:    "sensetime-app",
		AppKey:   "sensetime-key",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if _, ok := params["sceneList"]; ok {
		t.Fatalf("expected empty sceneList to be omitted, got %#v", params["sceneList"])
	}
}

func TestSpatiusAvatarSerializesExpectedParams(t *testing.T) {
	sr := SampleRate24kHz
	expire := 30
	config := NewSpatiusAvatar(SpatiusAvatarOptions{
		SpatiusAPIKey:        "spatius-key",
		SpatiusAppID:         "spatius-app",
		SpatiusAvatarID:      "avatar-1",
		AgoraUID:             "2001",
		Region:               "cn-beijing",
		SampleRate:           &sr,
		SessionExpireMinutes: &expire,
	}).ToConfig()

	if config["vendor"] != "spatius" {
		t.Fatalf("unexpected vendor: %v", config["vendor"])
	}
	params := config["params"].(map[string]interface{})
	if params["spatius_api_key"] != "spatius-key" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["spatius_app_id"] != "spatius-app" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["spatius_avatar_id"] != "avatar-1" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["agora_uid"] != "2001" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["region"] != "cn-beijing" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["sample_rate"] != 24000 {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["session_expire_minutes"] != 30 {
		t.Fatalf("unexpected params: %#v", params)
	}
}
