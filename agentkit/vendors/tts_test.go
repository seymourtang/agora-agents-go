package vendors

import (
	"reflect"
	"testing"
)

func TestTTSVendorParamsMatchGeneratedCoreShapes(t *testing.T) {
	sampleRate := SampleRate24kHz

	cases := []struct {
		name   string
		params map[string]interface{}
		want   map[string]interface{}
	}{
		{
			name: "amazon",
			params: NewAmazonTTS(AmazonTTSOptions{
				AccessKey: "access",
				SecretKey: "secret",
				Region:    "us-east-1",
				VoiceID:   "Joanna",
				Engine:    "neural",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"aws_access_key_id":     "access",
				"aws_secret_access_key": "secret",
				"region_name":           "us-east-1",
				"voice":                 "Joanna",
				"engine":                "neural",
			},
		},
		{
			name: "google",
			params: NewGoogleTTS(GoogleTTSOptions{
				Key:          "{}",
				VoiceName:    "en-US-JennyNeural",
				LanguageCode: "en-US",
				SampleRate:   &sampleRate,
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"credentials":          "{}",
				"VoiceSelectionParams": map[string]interface{}{"name": "en-US-JennyNeural", "language_code": "en-US"},
				"AudioConfig":          map[string]interface{}{"sample_rate_hertz": 24000},
			},
		},
		{
			name: "cartesia",
			params: NewCartesiaTTS(CartesiaTTSOptions{
				APIKey:     "cartesia-key",
				VoiceID:    "voice",
				ModelID:    "sonic-2",
				SampleRate: &sampleRate,
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":       "cartesia-key",
				"model_id":      "sonic-2",
				"voice":         map[string]interface{}{"mode": "id", "id": "voice"},
				"output_format": map[string]interface{}{"container": "raw", "sample_rate": 24000},
			},
		},
		{
			name: "rime",
			params: NewRimeTTS(RimeTTSOptions{
				Key:     "rime-key",
				Speaker: "speaker",
				ModelID: "mist",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "rime-key", "speaker": "speaker", "modelId": "mist"},
		},
		{
			name: "fish",
			params: NewFishAudioTTS(FishAudioTTSOptions{
				Key:         "fish-key",
				ReferenceID: "ref",
				Backend:     "speech-1.5",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "fish-key", "reference_id": "ref", "backend": "speech-1.5"},
		},
		{
			name: "sarvam",
			params: NewSarvamTTS(SarvamTTSOptions{
				Key:                "sarvam-key",
				Speaker:            "anushka",
				TargetLanguageCode: "en-IN",
				SampleRate:         ptrInt(24000),
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_subscription_key": "sarvam-key",
				"speaker":              "anushka",
				"target_language_code": "en-IN",
				"sample_rate":          24000,
			},
		},
		{
			name: "murf",
			params: NewMurfTTS(MurfTTSOptions{
				Key:     "murf-key",
				VoiceID: "Ariana",
				BaseURL: "wss://murf.example/ws",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "murf-key", "base_url": "wss://murf.example/ws", "voiceId": "Ariana"},
		},
	}

	for _, tc := range cases {
		if !reflect.DeepEqual(tc.params, tc.want) {
			t.Fatalf("%s params mismatch\nwant: %#v\n got: %#v", tc.name, tc.want, tc.params)
		}
	}
}

func ptrInt(v int) *int {
	return &v
}
