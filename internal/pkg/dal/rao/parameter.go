package rao

//type Auth struct {
//	Type   string  `json:"type"`
//	Kv     *KV     `json:"kv"`
//	Bearer *Bearer `json:"bearer"`
//	Basic  *Basic  `json:"basic"`
//}

type Auth struct {
	Type          string   `json:"type"`
	Kv            KV       `json:"kv"`
	Bearer        Bearer   `json:"bearer"`
	Basic         Basic    `json:"basic"`
	Digest        Digest   `json:"digest"`
	Hawk          Hawk     `json:"hawk"`
	Awsv4         AwsV4    `json:"awsv4"`
	Ntlm          Ntlm     `json:"ntlm"`
	Edgegrid      Edgegrid `json:"edgegrid"`
	Oauth1        Oauth1   `json:"oauth1"`
	Bidirectional TLS      `json:"bidirectional"`
}
type TLS struct {
	CaCert     string `json:"ca_cert"`
	CaCertName string `json:"ca_cert_name"`
}
type Bearer struct {
	Key string `json:"key" bson:"key"`
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
	Username  string `json:"username" bson:"username"`
	Password  string `json:"password" bson:"password"`
	Realm     string `json:"realm" bson:"realm"`
	Nonce     string `json:"nonce" bson:"nonce"`
	Algorithm string `json:"algorithm" bson:"algorithm"`
	Qop       string `json:"qop" bson:"qop"`
	Nc        string `json:"nc" bson:"nc"`
	Cnonce    string `json:"cnonce" bson:"cnonce"`
	Opaque    string `json:"opaque" bson:"opaque"`
}

type Hawk struct {
	AuthID             string `json:"authId" bson:"authID"`
	AuthKey            string `json:"authKey" bson:"authKey"`
	Algorithm          string `json:"algorithm" bson:"algorithm"`
	User               string `json:"user" bson:"user"`
	Nonce              string `json:"nonce" bson:"nonce"`
	ExtraData          string `json:"extraData" bson:"extraData"`
	App                string `json:"app" bson:"app"`
	Delegation         string `json:"delegation" bson:"delegation"`
	Timestamp          string `json:"timestamp" bson:"timestamp"`
	IncludePayloadHash int    `json:"includePayloadHash" bson:"includePayloadHash"`
}

type AwsV4 struct {
	AccessKey          string `json:"accessKey" bson:"accessKey"`
	SecretKey          string `json:"secretKey" bson:"secretKey"`
	Region             string `json:"region" bson:"region"`
	Service            string `json:"service" bson:"service"`
	SessionToken       string `json:"sessionToken" bson:"sessionToken"`
	AddAuthDataToQuery int    `json:"addAuthDataToQuery" bson:"addAuthDataToQuery"`
}

type Ntlm struct {
	Username            string `json:"username" bson:"username"`
	Password            string `json:"password" bson:"password"`
	Domain              string `json:"domain" bson:"domain"`
	Workstation         string `json:"workstation" bson:"workstation"`
	DisableRetryRequest int    `json:"disableRetryRequest" bson:"disableRetryRequest"`
}

type Edgegrid struct {
	AccessToken   string `json:"accessToken" bson:"accessToken"`
	ClientToken   string `json:"clientToken" bson:"clientToken"`
	ClientSecret  string `json:"clientSecret" bson:"clientSecret"`
	Nonce         string `json:"nonce" bson:"nonce"`
	Timestamp     string `json:"timestamp" bson:"timestamp"`
	BaseURi       string `json:"baseURi" bson:"baseURi"`
	HeadersToSign string `json:"headersToSign" bson:"headersToSign"`
}

type Oauth1 struct {
	ConsumerKey          string `json:"consumerKey" bson:"consumerKey"`
	ConsumerSecret       string `json:"consumerSecret" bson:"consumerSecret"`
	SignatureMethod      string `json:"signatureMethod" bson:"signatureMethod"`
	AddEmptyParamsToSign int    `json:"addEmptyParamsToSign" bson:"addEmptyParamsToSign"`
	IncludeBodyHash      int    `json:"includeBodyHash" bson:"includeBodyHash"`
	AddParamsToHeader    int    `json:"addParamsToHeader" bson:"addParamsToHeader"`
	Realm                string `json:"realm" bson:"realm"`
	Version              string `json:"version" bson:"version"`
	Nonce                string `json:"nonce" bson:"nonce"`
	Timestamp            string `json:"timestamp" bson:"timestamp"`
	Verifier             string `json:"verifier" bson:"verifier"`
	Callback             string `json:"callback" bson:"callback"`
	TokenSecret          string `json:"tokenSecret" bson:"tokenSecret"`
	Token                string `json:"token" bson:"token"`
}

type Query struct {
	Parameter []Parameter `json:"parameter"`
}

type Header struct {
	Parameter []Parameter `json:"parameter" bson:"parameter"`
}

type Body struct {
	Mode      string      `json:"mode"`
	Parameter []Parameter `json:"parameter"`
	Raw       string      `json:"raw"`
}

type Parameter struct {
	IsChecked   int32       `json:"is_checked"`
	Type        string      `json:"type"`
	Key         string      `json:"key"`
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
	Parameter []Parameter `json:"parameter"`
}

type Resful struct {
	Parameter []Parameter `json:"parameter"`
}

type Request struct {
	PreUrl       string       `json:"pre_url"`
	URL          string       `json:"url"`
	Method       string       `json:"method"`
	Description  string       `json:"description"`
	Auth         Auth         `json:"auth"`
	Body         Body         `json:"body"`
	Header       Header       `json:"header"`
	Query        Query        `json:"query"`
	Event        Event        `json:"event"`
	Cookie       Cookie       `json:"cookie"`
	Assert       []Assert     `json:"assert"`
	Regex        []Regex      `json:"regex"`
	HttpApiSetup HttpApiSetup `json:"http_api_setup"`
}

type Assert struct {
	ResponseType int32  `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	IsChecked    int    `json:"is_checked"`
	Index        int    `json:"index"` // 正则时提取第几个值
}

type Regex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Type      int    `json:"type"`       // 0 正则  1 json
	Var       string `json:"var"`
	Val       string `json:"val"`
	Express   string `json:"express"`
	Index     int    `json:"index"` // 正则时提取第几个值
}
