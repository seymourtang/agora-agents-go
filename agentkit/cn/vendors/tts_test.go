package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestCNTTSVendorParams(t *testing.T) {
	cases := []struct {
		name   string
		params map[string]interface{}
		want   map[string]interface{}
	}{
		{
			name: "minimax byok full",
			params: NewMiniMaxTTS(MiniMaxTTSOptions{
				Key:   "minimax-key",
				Model: "speech-01-turbo",
				VoiceSetting: &MiniMaxVoiceSetting{
					VoiceID:              "female-shaonv",
					Speed:                ptrInt(1),
					Volume:               ptrInt(1),
					Pitch:                ptrInt(0),
					Emotion:              "happy",
					LatexRead:            boolPtr(true),
					EnglishNormalization: boolPtr(true),
				},
				AudioSetting: &MiniMaxAudioSetting{
					SampleRate: 16000,
				},
				PronunciationDict: &MiniMaxPronunciationDict{
					Tone: []string{"alpha/(ae1)(l)(f)(ah0)", "beta/(b)(ey1)(t)(ah0)"},
				},
				LanguageBoost: "auto",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":   "minimax-key",
				"model": "speech-01-turbo",
				"voice_setting": map[string]interface{}{
					"voice_id":              "female-shaonv",
					"speed":                 1,
					"vol":                   1,
					"pitch":                 0,
					"emotion":               "happy",
					"latex_read":            true,
					"english_normalization": true,
				},
				"audio_setting": map[string]interface{}{
					"sample_rate": 16000,
				},
				"pronunciation_dict": map[string]interface{}{
					"tone": []string{"alpha/(ae1)(l)(f)(ah0)", "beta/(b)(ey1)(t)(ah0)"},
				},
				"language_boost": "auto",
			},
		},
		{
			name: "microsoft",
			params: NewMicrosoftTTS(MicrosoftTTSOptions{
				Key:       "ms-key",
				Region:    "eastus",
				VoiceName: "zh-CN-XiaoxiaoNeural",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":        "ms-key",
				"region":     "eastus",
				"voice_name": "zh-CN-XiaoxiaoNeural",
			},
		},
		{
			name: "tencent",
			params: NewTencentTTS(TencentTTSOptions{
				AppID:     "app-id",
				SecretID:  "secret-id",
				SecretKey: "secret-key",
				AdditionalParams: map[string]interface{}{
					"app_id":     "override-app",
					"voice_type": 999,
					"codec":      "pcm",
				},
				VoiceType:       ptrInt(601005),
				Volume:          ptrFloat(0),
				Speed:           ptrFloat(0),
				EmotionCategory: "happy",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"app_id":           "app-id",
				"secret_id":        "secret-id",
				"secret_key":       "secret-key",
				"voice_type":       601005,
				"codec":            "pcm",
				"volume":           float64(0),
				"speed":            float64(0),
				"emotion_category": "happy",
			},
		},
		{
			name: "bytedance",
			params: NewBytedanceTTS(BytedanceTTSOptions{
				Token:       "token",
				AppID:       "app-id",
				Cluster:     "volcano_tts",
				VoiceType:   "BV700_streaming",
				SpeedRatio:  ptrFloat(1),
				VolumeRatio: ptrFloat(1),
				PitchRatio:  ptrFloat(1),
				Emotion:     "happy",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"token":        "token",
				"app_id":       "app-id",
				"cluster":      "volcano_tts",
				"voice_type":   "BV700_streaming",
				"speed_ratio":  float64(1),
				"volume_ratio": float64(1),
				"pitch_ratio":  float64(1),
				"emotion":      "happy",
			},
		},
		{
			name: "cosyvoice",
			params: NewCosyVoiceTTS(CosyVoiceTTSOptions{
				APIKey:     "api-key",
				Model:      "cosyvoice-v1",
				Voice:      "longxiaochun",
				SampleRate: ptrInt(16000),
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":     "api-key",
				"model":       "cosyvoice-v1",
				"voice":       "longxiaochun",
				"sample_rate": 16000,
			},
		},
		{
			name: "bytedance duplex",
			params: NewBytedanceDuplexTTS(BytedanceDuplexTTSOptions{
				AppID:   "app-id",
				Token:   "token",
				Speaker: "zh_female_shuangkuaisisi_moon_bigtts",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"app_id":  "app-id",
				"token":   "token",
				"speaker": "zh_female_shuangkuaisisi_moon_bigtts",
			},
		},
		{
			name: "stepfun",
			params: NewStepFunTTS(StepFunTTSOptions{
				APIKey:  "step-key",
				Model:   "step-tts-mini",
				VoiceID: "cixingnansheng",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":  "step-key",
				"model":    "step-tts-mini",
				"voice_id": "cixingnansheng",
			},
		},
		{
			name: "generic",
			params: NewGenericTTS(GenericTTSOptions{
				URL:            "https://tts.example.com/v1/audio/speech",
				Headers:        map[string]string{"Authorization": "Bearer token"},
				APIKey:         "generic-key",
				Model:          "gpt-4o-mini-tts",
				Voice:          "alloy",
				ResponseFormat: "pcm",
				AdditionalParams: map[string]interface{}{
					"api_key":         "additional-key",
					"model":           "additional-model",
					"voice":           "additional-voice",
					"response_format": "mp3",
					"custom_param":    "custom-value",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":         "generic-key",
				"model":           "gpt-4o-mini-tts",
				"voice":           "alloy",
				"response_format": "pcm",
				"custom_param":    "custom-value",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.params, tc.want) {
				t.Fatalf("params mismatch\nwant: %#v\n got: %#v", tc.want, tc.params)
			}
		})
	}
}

func TestCNGenericTTSRequiresOnlyURLAndMatchesGeneratedHTTPVendor(t *testing.T) {
	t.Run("URL only", func(t *testing.T) {
		config := NewGenericTTS(GenericTTSOptions{
			URL: "https://tts.example.com/v1/audio/speech",
		}).ToConfig()

		if got := config["vendor"]; got != "generic_http" {
			t.Fatalf("vendor = %v, want generic_http", got)
		}
		if _, ok := config["headers"]; ok {
			t.Fatalf("headers should be omitted when empty: %#v", config)
		}
		wantParams := map[string]interface{}{}
		if got := config["params"]; !reflect.DeepEqual(got, wantParams) {
			t.Fatalf("params mismatch\nwant: %#v\n got: %#v", wantParams, got)
		}

		payload, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}
		var generated Agora.Tts
		if err := json.Unmarshal(payload, &generated); err != nil {
			t.Fatalf("unmarshal into generated TTS union: %v", err)
		}
		if generated.GenericHTTP == nil {
			t.Fatalf("generated GenericHTTP vendor is nil: %#v", generated)
		}
		if _, err := json.Marshal(generated); err != nil {
			t.Fatalf("marshal generated TTS union: %v", err)
		}
	})

	t.Run("missing URL", func(t *testing.T) {
		assertPanicValue(t, "GenericTTS requires URL", func() {
			NewGenericTTS(GenericTTSOptions{})
		})
	})
}

func TestCNGenericTTSURLValidation(t *testing.T) {
	validURLs := []struct {
		name string
		url  string
	}{
		{name: "HTTP", url: "http://tts.example.com/v1/audio/speech"},
		{name: "HTTPS", url: "https://tts.example.com/v1/audio/speech"},
	}
	for _, tc := range validURLs {
		t.Run(tc.name, func(t *testing.T) {
			config := NewGenericTTS(GenericTTSOptions{URL: tc.url}).ToConfig()
			if got := config["vendor"]; got != "generic_http" {
				t.Fatalf("vendor = %v, want generic_http", got)
			}
		})
	}

	invalidURLs := []struct {
		name      string
		url       string
		wantPanic string
	}{
		{name: "missing scheme", url: "tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "missing host", url: "https:///v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "malformed", url: "https://tts.example.com/%zz", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "WebSocket", url: "ws://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "secure WebSocket", url: "wss://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "FTP", url: "ftp://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
	}
	for _, tc := range invalidURLs {
		t.Run(tc.name, func(t *testing.T) {
			assertPanicValue(t, tc.wantPanic, func() {
				NewGenericTTS(GenericTTSOptions{URL: tc.url})
			})
		})
	}
}

func assertPanicValue(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %s", got, want)
		}
	}()
	fn()
}

func ptrInt(v int) *int           { return &v }
func ptrFloat(v float64) *float64 { return &v }
func boolPtr(v bool) *bool        { return &v }
