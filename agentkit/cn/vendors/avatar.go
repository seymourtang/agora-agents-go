package vendors

type SensetimePosition struct {
	X int
	Y int
}

type SensetimeDigitalRole struct {
	FaceFeatureID string
	Position      SensetimePosition
	URL           string
}

type SensetimeScene struct {
	DigitalRole SensetimeDigitalRole
}

type SensetimeAvatarOptions struct {
	AgoraUID         string
	AgoraToken       string
	AppID            string
	AppKey           string
	SceneList        []SensetimeScene
	Enable           *bool
	AdditionalParams map[string]interface{}
}

type SensetimeAvatar struct {
	options SensetimeAvatarOptions
}

func NewSensetimeAvatar(opts SensetimeAvatarOptions) *SensetimeAvatar {
	if opts.AgoraUID == "" {
		panic("SensetimeAvatar requires AgoraUID")
	}
	if opts.AppID == "" {
		panic("SensetimeAvatar requires AppID")
	}
	if opts.AppKey == "" {
		panic("SensetimeAvatar requires AppKey")
	}
	if len(opts.SceneList) == 0 {
		panic("SensetimeAvatar requires SceneList")
	}
	return &SensetimeAvatar{options: opts}
}

func (s *SensetimeAvatar) RequiredSampleRate() SampleRate {
	return 0
}

func (s *SensetimeAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range s.options.AdditionalParams {
		params[k] = v
	}
	params["agora_uid"] = s.options.AgoraUID
	if s.options.AgoraToken != "" {
		params["agora_token"] = s.options.AgoraToken
	}
	params["appId"] = s.options.AppID
	params["app_key"] = s.options.AppKey
	sceneList := make([]map[string]interface{}, 0, len(s.options.SceneList))
	for _, scene := range s.options.SceneList {
		sceneList = append(sceneList, map[string]interface{}{
			"digital_role": map[string]interface{}{
				"face_feature_id": scene.DigitalRole.FaceFeatureID,
				"position": map[string]interface{}{
					"x": scene.DigitalRole.Position.X,
					"y": scene.DigitalRole.Position.Y,
				},
				"url": scene.DigitalRole.URL,
			},
		})
	}
	params["sceneList"] = sceneList

	enable := true
	if s.options.Enable != nil {
		enable = *s.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "sensetime",
		"params": params,
	}
}
