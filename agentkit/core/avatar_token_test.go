package core

import "testing"

func TestIsAvatarTokenManagedIncludesSensetime(t *testing.T) {
	if !IsAvatarTokenManaged("sensetime") {
		t.Fatal("expected sensetime avatar to use managed agora_token generation")
	}
}

func TestIsAvatarTokenManagedIncludesSpatius(t *testing.T) {
	if !IsAvatarTokenManaged("spatius") {
		t.Fatal("expected spatius avatar to use managed agora_token generation")
	}
}

func TestBuildPropertiesMapAutoGeneratesSensetimeAvatarToken(t *testing.T) {
	base := &BaseAgent{
		Avatar: map[string]interface{}{
			"enable": true,
			"vendor": "sensetime",
			"params": map[string]interface{}{
				"agora_uid": "2001",
				"appId":     "sensetime-app",
				"app_key":   "sensetime-key",
				"sceneList": []map[string]interface{}{
					{
						"digital_role": map[string]interface{}{
							"face_feature_id": "face-1",
							"position": map[string]interface{}{
								"x": 0,
								"y": 0,
							},
							"url": "https://example.test/model",
						},
					},
				},
			},
		},
	}

	props, err := BuildPropertiesMap(base, ToPropertiesOptions{
		Channel:              "avatar-channel",
		AgentUID:             "1001",
		RemoteUIDs:           []string{"1002"},
		AppID:                "agora-app",
		AppCertificate:       "agora-cert",
		ExpiresIn:            3600,
		SkipVendorValidation: true,
	}, func(GenerateConvoAITokenOptions) (string, error) {
		return "avatar-token", nil
	})
	if err != nil {
		t.Fatalf("BuildPropertiesMap returned error: %v", err)
	}

	avatar := props["avatar"].(map[string]interface{})
	params := avatar["params"].(map[string]interface{})
	if params["agora_token"] != "avatar-token" {
		t.Fatalf("expected auto-generated avatar token, got %#v", params["agora_token"])
	}
}
