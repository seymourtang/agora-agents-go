package vendors

type FengmingSTT struct{}

func NewFengmingSTT() *FengmingSTT {
	return &FengmingSTT{}
}

func (f *FengmingSTT) ToConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vendor": "fengming",
	}
	return config
}

type TencentSTTOptions struct {
	Key              string
	AppID            string
	Secret           string
	EngineModelType  string
	VoiceID          string
	AdditionalParams map[string]interface{}
}

type TencentSTT struct {
	options TencentSTTOptions
}

func NewTencentSTT(opts TencentSTTOptions) *TencentSTT {
	if opts.Key == "" {
		panic("TencentSTT requires Key")
	}
	if opts.AppID == "" {
		panic("TencentSTT requires AppID")
	}
	if opts.Secret == "" {
		panic("TencentSTT requires Secret")
	}
	return &TencentSTT{options: opts}
}

func (t *TencentSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range t.options.AdditionalParams {
		params[k] = v
	}
	params["key"] = t.options.Key
	params["app_id"] = t.options.AppID
	params["secret"] = t.options.Secret
	if t.options.EngineModelType != "" {
		params["engine_model_type"] = t.options.EngineModelType
	}
	if t.options.VoiceID != "" {
		params["voice_id"] = t.options.VoiceID
	}
	return map[string]interface{}{
		"vendor": "tencent",
		"params": params,
	}
}

type MicrosoftSTTOptions struct {
	Key              string
	Region           string
	Language         string
	PhraseList       []string
	AdditionalParams map[string]interface{}
}

type MicrosoftSTT struct {
	options MicrosoftSTTOptions
}

func NewMicrosoftSTT(opts MicrosoftSTTOptions) *MicrosoftSTT {
	if opts.Key == "" {
		panic("MicrosoftSTT requires Key")
	}
	if opts.Region == "" {
		panic("MicrosoftSTT requires Region")
	}
	if opts.Language == "" {
		panic("MicrosoftSTT requires Language")
	}
	return &MicrosoftSTT{options: opts}
}

func (m *MicrosoftSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range m.options.AdditionalParams {
		params[k] = v
	}
	params["key"] = m.options.Key
	params["region"] = m.options.Region
	params["language"] = m.options.Language
	if len(m.options.PhraseList) > 0 {
		params["phrase_list"] = append([]string(nil), m.options.PhraseList...)
	}

	return map[string]interface{}{
		"vendor": "microsoft",
		"params": params,
	}
}

type XfyunSTTOptions struct {
	APIKey           string
	AppID            string
	APISecret        string
	Language         string
	AdditionalParams map[string]interface{}
}

type XfyunSTT struct {
	options XfyunSTTOptions
}

func NewXfyunSTT(opts XfyunSTTOptions) *XfyunSTT {
	if opts.APIKey == "" {
		panic("XfyunSTT requires APIKey")
	}
	if opts.AppID == "" {
		panic("XfyunSTT requires AppID")
	}
	if opts.APISecret == "" {
		panic("XfyunSTT requires APISecret")
	}
	return &XfyunSTT{options: opts}
}

func (x *XfyunSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range x.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = x.options.APIKey
	params["app_id"] = x.options.AppID
	params["api_secret"] = x.options.APISecret
	if x.options.Language != "" {
		params["language"] = x.options.Language
	}
	return map[string]interface{}{
		"vendor": "xfyun",
		"params": params,
	}
}

type XfyunBigModelSTTOptions struct {
	APIKey           string
	AppID            string
	APISecret        string
	LanguageName     string
	Language         string
	AdditionalParams map[string]interface{}
}

type XfyunBigModelSTT struct {
	options XfyunBigModelSTTOptions
}

func NewXfyunBigModelSTT(opts XfyunBigModelSTTOptions) *XfyunBigModelSTT {
	if opts.APIKey == "" {
		panic("XfyunBigModelSTT requires APIKey")
	}
	if opts.AppID == "" {
		panic("XfyunBigModelSTT requires AppID")
	}
	if opts.APISecret == "" {
		panic("XfyunBigModelSTT requires APISecret")
	}
	return &XfyunBigModelSTT{options: opts}
}

func (x *XfyunBigModelSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range x.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = x.options.APIKey
	params["app_id"] = x.options.AppID
	params["api_secret"] = x.options.APISecret
	if x.options.LanguageName != "" {
		params["language_name"] = x.options.LanguageName
	}
	if x.options.Language != "" {
		params["language"] = x.options.Language
	}
	return map[string]interface{}{
		"vendor": "xfyun_bigmodel",
		"params": params,
	}
}

type XfyunDialectSTTOptions struct {
	AppID            string
	AccessKeyID      string
	AccessKeySecret  string
	Language         string
	AdditionalParams map[string]interface{}
}

type XfyunDialectSTT struct {
	options XfyunDialectSTTOptions
}

func NewXfyunDialectSTT(opts XfyunDialectSTTOptions) *XfyunDialectSTT {
	if opts.AppID == "" {
		panic("XfyunDialectSTT requires AppID")
	}
	if opts.AccessKeyID == "" {
		panic("XfyunDialectSTT requires AccessKeyID")
	}
	if opts.AccessKeySecret == "" {
		panic("XfyunDialectSTT requires AccessKeySecret")
	}
	return &XfyunDialectSTT{options: opts}
}

func (x *XfyunDialectSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range x.options.AdditionalParams {
		params[k] = v
	}
	params["app_id"] = x.options.AppID
	params["access_key_id"] = x.options.AccessKeyID
	params["access_key_secret"] = x.options.AccessKeySecret
	if x.options.Language != "" {
		params["language"] = x.options.Language
	}
	return map[string]interface{}{
		"vendor": "xfyun_dialect",
		"params": params,
	}
}
