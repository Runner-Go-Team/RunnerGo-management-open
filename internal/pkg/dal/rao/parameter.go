package rao

//type Auth struct {
//	Type   string  `json:"type"`
//	Kv     *KV     `json:"kv"`
//	Bearer *Bearer `json:"bearer"`
//	Basic  *Basic  `json:"basic"`
//}

type Auth struct {
	Type     string    `json:"type"`
	Kv       *KV       `json:"kv"`
	Bearer   *Bearer   `json:"bearer"`
	Basic    *Basic    `json:"basic"`
	Digest   *Digest   `json:"digest"`
	Hawk     *Hawk     `json:"hawk"`
	Awsv4    *AwsV4    `json:"awsv4"`
	Ntlm     *Ntlm     `json:"ntlm"`
	Edgegrid *Edgegrid `json:"edgegrid"`
	Oauth1   *Oauth1   `json:"oauth1"`
}

type Bearer struct {
	Key string `json:"key"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Basic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Digest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Realm     string `json:"realm"`
	Nonce     string `json:"nonce"`
	Algorithm string `json:"algorithm"`
	Qop       string `json:"qop"`
	Nc        string `json:"nc"`
	Cnonce    string `json:"cnonce"`
	Opaque    string `json:"opaque"`
}

type Hawk struct {
	AuthID             string `json:"authId"`
	AuthKey            string `json:"authKey"`
	Algorithm          string `json:"algorithm"`
	User               string `json:"user"`
	Nonce              string `json:"nonce"`
	ExtraData          string `json:"extraData"`
	App                string `json:"app"`
	Delegation         string `json:"delegation"`
	Timestamp          string `json:"timestamp"`
	IncludePayloadHash int    `json:"includePayloadHash"`
}

type AwsV4 struct {
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	SessionToken       string `json:"sessionToken"`
	AddAuthDataToQuery int    `json:"addAuthDataToQuery"`
}

type Ntlm struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Domain              string `json:"domain"`
	Workstation         string `json:"workstation"`
	DisableRetryRequest int    `json:"disableRetryRequest"`
}

type Edgegrid struct {
	AccessToken   string `json:"accessToken"`
	ClientToken   string `json:"clientToken"`
	ClientSecret  string `json:"clientSecret"`
	Nonce         string `json:"nonce"`
	Timestamp     string `json:"timestamp"`
	BaseURi       string `json:"baseURi"`
	HeadersToSign string `json:"headersToSign"`
}

type Oauth1 struct {
	ConsumerKey          string `json:"consumerKey"`
	ConsumerSecret       string `json:"consumerSecret"`
	SignatureMethod      string `json:"signatureMethod"`
	AddEmptyParamsToSign int    `json:"addEmptyParamsToSign"`
	IncludeBodyHash      int    `json:"includeBodyHash"`
	AddParamsToHeader    int    `json:"addParamsToHeader"`
	Realm                string `json:"realm"`
	Version              string `json:"version"`
	Nonce                string `json:"nonce"`
	Timestamp            string `json:"timestamp"`
	Verifier             string `json:"verifier"`
	Callback             string `json:"callback"`
	TokenSecret          string `json:"tokenSecret"`
	Token                string `json:"token"`
}

type Query struct {
	Parameter []*Parameter `json:"parameter"`
}

type Header struct {
	Parameter []*Parameter `json:"parameter"`
}

type Body struct {
	Mode      string       `json:"mode"`
	Parameter []*Parameter `json:"parameter"`
	Raw       string       `json:"raw"`
}

type Parameter struct {
	IsChecked int32  `json:"is_checked"`
	Type      string `json:"type"`
	Key       string `json:"key"`
	//Value       string   `json:"value"`
	Value       interface{} `json:"value"`
	NotNull     int32       `json:"not_null"`
	Description string      `json:"description"`
	FileBase64  []string    `json:"fileBase64"`
	FieldType   string      `json:"field_type"`
}

type Script struct {
	PreScript       string `json:"pre_script"`
	Test            string `json:"test"`
	PreScriptSwitch bool   `json:"pre_script_switch"`
	TestSwitch      bool   `json:"test_switch"`
}

type Event struct {
	PreScript string `json:"pre_script"`
	Test      string `json:"test"`
}

type Cookie struct {
	Parameter []*Parameter `json:"parameter"`
}

type Resful struct {
	Parameter []*Parameter `json:"parameter"`
}

type Request struct {
	PreUrl      string  `json:"pre_url"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Auth        *Auth   `json:"auth"`
	Body        *Body   `json:"body"`
	Header      *Header `json:"header"`
	Query       *Query  `json:"query"`
	Event       *Event  `json:"event"`
	Cookie      *Cookie `json:"cookie"`
	Resful      *Resful `json:"resful"`
}

type Success struct {
	Raw       string       `json:"raw"`
	Parameter []*Parameter `json:"parameter"`
}

type Error struct {
	Raw       string       `json:"raw"`
	Parameter []*Parameter `json:"parameter"`
}

type Response struct {
	Success *Success `json:"success"`
	Error   *Error   `json:"error"`
}

type Assert struct {
	ResponseType int32  `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	IsChecked    int    `json:"is_checked"`
}

type Regex struct {
	Var       string `json:"var"`
	Express   string `json:"express"`
	Val       string `json:"val"`
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Type      int    `json:"type"`       // 0 正则  1 json
}
