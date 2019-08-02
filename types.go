package govalidator

import (
	"context"
	"reflect"
	"regexp"
	"sort"
	"sync"
)

type Validator interface {
	ValidateStruct(interface{}) (bool, error)
	ValidateStructCtx(context.Context, interface{}) (bool, error)
	AddCustomTypeTagFn(string, CustomTypeValidator)
}

type validator struct {
	// CustomTypeTagMap is a map of functions that can be used as tags for ValidateStruct function.
	// Use this to validate compound or custom types that need to be handled as a whole, e.g.
	// `type UUID [16]byte` (this would be handled as an array of bytes).
	CustomTypeTagMap *customTypeTagMap
}

func New() Validator {
	return &validator{
		CustomTypeTagMap: &customTypeTagMap{validators: make(map[string]CustomTypeValidator)},
	}
}

// ValidatorFn is a wrapper for a validator function that returns bool and accepts string.
type ValidatorFn func(str string) bool

// CustomTypeValidator is a wrapper for validator functions that returns bool and accepts any type.
// The second parameter should be the context (in the case of validating a struct: the whole object being validated).
type CustomTypeValidator func(ctx context.Context, i interface{}, o interface{}) bool

// ParamValidator is a wrapper for validator functions that accepts additional parameters.
type ParamValidator func(str string, params ...string) bool
type tagOptionsMap map[string]tagOption

func (t tagOptionsMap) orderedKeys() []string {
	var keys []string
	for k := range t {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(a, b int) bool {
		return t[keys[a]].order < t[keys[b]].order
	})

	return keys
}

type tagOption struct {
	name               string
	customErrorMessage string
	order              int
}

// UnsupportedTypeError is a wrapper for reflect.Type
type UnsupportedTypeError struct {
	Type reflect.Type
}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements the methods to sort by string.
type stringValues []reflect.Value

// ParamTagMap is a map of functions accept variants parameters
var ParamTagMap = map[string]ParamValidator{
	"length":       ByteLength,
	"range":        Range,
	"runelength":   RuneLength,
	"stringlength": StringLength,
	"matches":      StringMatches,
	"in":           isInRaw,
	"rsapub":       IsRsaPub,
}

// ParamTagRegexMap maps param tags to their respective regexes.
var ParamTagRegexMap = map[string]*regexp.Regexp{
	"range":        regexp.MustCompile("^range\\((\\d+)\\|(\\d+)\\)$"),
	"length":       regexp.MustCompile("^length\\((\\d+)\\|(\\d+)\\)$"),
	"runelength":   regexp.MustCompile("^runelength\\((\\d+)\\|(\\d+)\\)$"),
	"stringlength": regexp.MustCompile("^stringlength\\((\\d+)\\|(\\d+)\\)$"),
	"in":           regexp.MustCompile(`^in\((.*)\)`),
	"matches":      regexp.MustCompile(`^matches\((.+)\)$`),
	"rsapub":       regexp.MustCompile("^rsapub\\((\\d+)\\)$"),
}

type customTypeTagMap struct {
	validators map[string]CustomTypeValidator

	sync.RWMutex
}

func (tm *customTypeTagMap) Get(name string) (CustomTypeValidator, bool) {
	tm.RLock()
	defer tm.RUnlock()
	v, ok := tm.validators[name]
	return v, ok
}

func (tm *customTypeTagMap) Set(name string, ctv CustomTypeValidator) {
	tm.Lock()
	defer tm.Unlock()
	tm.validators[name] = ctv
}

// Tagmap is a map of functions, that can be used as tags for ValidateStruct function.
var TagMap = map[string]ValidatorFn{
	"email":              IsEmail,
	"url":                IsURL,
	"dialstring":         IsDialString,
	"requrl":             IsRequestURL,
	"requri":             IsRequestURI,
	"alpha":              IsAlpha,
	"utfletter":          IsUTFLetter,
	"alphanum":           IsAlphanumeric,
	"utfletternum":       IsUTFLetterNumeric,
	"numeric":            IsNumeric,
	"utfnumeric":         IsUTFNumeric,
	"utfdigit":           IsUTFDigit,
	"hexadecimal":        IsHexadecimal,
	"hexcolor":           IsHexcolor,
	"rgbcolor":           IsRGBcolor,
	"lowercase":          IsLowerCase,
	"uppercase":          IsUpperCase,
	"int":                IsInt,
	"float":              IsFloat,
	"null":               IsNull,
	"uuid":               IsUUID,
	"uuidv3":             IsUUIDv3,
	"uuidv4":             IsUUIDv4,
	"uuidv5":             IsUUIDv5,
	"creditcard":         IsCreditCard,
	"isbn10":             IsISBN10,
	"isbn13":             IsISBN13,
	"json":               IsJSON,
	"multibyte":          IsMultibyte,
	"ascii":              IsASCII,
	"printableascii":     IsPrintableASCII,
	"fullwidth":          IsFullWidth,
	"halfwidth":          IsHalfWidth,
	"variablewidth":      IsVariableWidth,
	"base64":             IsBase64,
	"datauri":            IsDataURI,
	"ip":                 IsIP,
	"port":               IsPort,
	"ipv4":               IsIPv4,
	"ipv6":               IsIPv6,
	"dns":                IsDNSName,
	"host":               IsHost,
	"mac":                IsMAC,
	"latitude":           IsLatitude,
	"longitude":          IsLongitude,
	"ssn":                IsSSN,
	"semver":             IsSemver,
	"rfc3339":            IsRFC3339,
	"rfc3339WithoutZone": IsRFC3339WithoutZone,
	"ISO3166Alpha2":      IsISO3166Alpha2,
	"ISO3166Alpha3":      IsISO3166Alpha3,
	"ISO4217":            IsISO4217,
}

// ISO3166Entry stores country codes
type ISO3166Entry struct {
	EnglishShortName string
	FrenchShortName  string
	Alpha2Code       string
	Alpha3Code       string
	Numeric          string
}

//ISO3166List based on https://www.iso.org/obp/ui/#search/code/ Code Type "Officially Assigned Codes"
var ISO3166List = []ISO3166Entry{
	{"Afghanistan", "Afghanistan (l')", "AF", "AFG", "004"},
	{"Albania", "Albanie (l')", "AL", "ALB", "008"},
	{"Antarctica", "Antarctique (l')", "AQ", "ATA", "010"},
	{"Algeria", "Algérie (l')", "DZ", "DZA", "012"},
	{"American Samoa", "Samoa américaines (les)", "AS", "ASM", "016"},
	{"Andorra", "Andorre (l')", "AD", "AND", "020"},
	{"Angola", "Angola (l')", "AO", "AGO", "024"},
	{"Antigua and Barbuda", "Antigua-et-Barbuda", "AG", "ATG", "028"},
	{"Azerbaijan", "Azerbaïdjan (l')", "AZ", "AZE", "031"},
	{"Argentina", "Argentine (l')", "AR", "ARG", "032"},
	{"Australia", "Australie (l')", "AU", "AUS", "036"},
	{"Austria", "Autriche (l')", "AT", "AUT", "040"},
	{"Bahamas (the)", "Bahamas (les)", "BS", "BHS", "044"},
	{"Bahrain", "Bahreïn", "BH", "BHR", "048"},
	{"Bangladesh", "Bangladesh (le)", "BD", "BGD", "050"},
	{"Armenia", "Arménie (l')", "AM", "ARM", "051"},
	{"Barbados", "Barbade (la)", "BB", "BRB", "052"},
	{"Belgium", "Belgique (la)", "BE", "BEL", "056"},
	{"Bermuda", "Bermudes (les)", "BM", "BMU", "060"},
	{"Bhutan", "Bhoutan (le)", "BT", "BTN", "064"},
	{"Bolivia (Plurinational State of)", "Bolivie (État plurinational de)", "BO", "BOL", "068"},
	{"Bosnia and Herzegovina", "Bosnie-Herzégovine (la)", "BA", "BIH", "070"},
	{"Botswana", "Botswana (le)", "BW", "BWA", "072"},
	{"Bouvet Island", "Bouvet (l'Île)", "BV", "BVT", "074"},
	{"Brazil", "Brésil (le)", "BR", "BRA", "076"},
	{"Belize", "Belize (le)", "BZ", "BLZ", "084"},
	{"British Indian Ocean Territory (the)", "Indien (le Territoire britannique de l'océan)", "IO", "IOT", "086"},
	{"Solomon Islands", "Salomon (Îles)", "SB", "SLB", "090"},
	{"Virgin Islands (British)", "Vierges britanniques (les Îles)", "VG", "VGB", "092"},
	{"Brunei Darussalam", "Brunéi Darussalam (le)", "BN", "BRN", "096"},
	{"Bulgaria", "Bulgarie (la)", "BG", "BGR", "100"},
	{"Myanmar", "Myanmar (le)", "MM", "MMR", "104"},
	{"Burundi", "Burundi (le)", "BI", "BDI", "108"},
	{"Belarus", "Bélarus (le)", "BY", "BLR", "112"},
	{"Cambodia", "Cambodge (le)", "KH", "KHM", "116"},
	{"Cameroon", "Cameroun (le)", "CM", "CMR", "120"},
	{"Canada", "Canada (le)", "CA", "CAN", "124"},
	{"Cabo Verde", "Cabo Verde", "CV", "CPV", "132"},
	{"Cayman Islands (the)", "Caïmans (les Îles)", "KY", "CYM", "136"},
	{"Central African Republic (the)", "République centrafricaine (la)", "CF", "CAF", "140"},
	{"Sri Lanka", "Sri Lanka", "LK", "LKA", "144"},
	{"Chad", "Tchad (le)", "TD", "TCD", "148"},
	{"Chile", "Chili (le)", "CL", "CHL", "152"},
	{"China", "Chine (la)", "CN", "CHN", "156"},
	{"Taiwan (Province of China)", "Taïwan (Province de Chine)", "TW", "TWN", "158"},
	{"Christmas Island", "Christmas (l'Île)", "CX", "CXR", "162"},
	{"Cocos (Keeling) Islands (the)", "Cocos (les Îles)/ Keeling (les Îles)", "CC", "CCK", "166"},
	{"Colombia", "Colombie (la)", "CO", "COL", "170"},
	{"Comoros (the)", "Comores (les)", "KM", "COM", "174"},
	{"Mayotte", "Mayotte", "YT", "MYT", "175"},
	{"Congo (the)", "Congo (le)", "CG", "COG", "178"},
	{"Congo (the Democratic Republic of the)", "Congo (la République démocratique du)", "CD", "COD", "180"},
	{"Cook Islands (the)", "Cook (les Îles)", "CK", "COK", "184"},
	{"Costa Rica", "Costa Rica (le)", "CR", "CRI", "188"},
	{"Croatia", "Croatie (la)", "HR", "HRV", "191"},
	{"Cuba", "Cuba", "CU", "CUB", "192"},
	{"Cyprus", "Chypre", "CY", "CYP", "196"},
	{"Czech Republic (the)", "tchèque (la République)", "CZ", "CZE", "203"},
	{"Benin", "Bénin (le)", "BJ", "BEN", "204"},
	{"Denmark", "Danemark (le)", "DK", "DNK", "208"},
	{"Dominica", "Dominique (la)", "DM", "DMA", "212"},
	{"Dominican Republic (the)", "dominicaine (la République)", "DO", "DOM", "214"},
	{"Ecuador", "Équateur (l')", "EC", "ECU", "218"},
	{"El Salvador", "El Salvador", "SV", "SLV", "222"},
	{"Equatorial Guinea", "Guinée équatoriale (la)", "GQ", "GNQ", "226"},
	{"Ethiopia", "Éthiopie (l')", "ET", "ETH", "231"},
	{"Eritrea", "Érythrée (l')", "ER", "ERI", "232"},
	{"Estonia", "Estonie (l')", "EE", "EST", "233"},
	{"Faroe Islands (the)", "Féroé (les Îles)", "FO", "FRO", "234"},
	{"Falkland Islands (the) [Malvinas]", "Falkland (les Îles)/Malouines (les Îles)", "FK", "FLK", "238"},
	{"South Georgia and the South Sandwich Islands", "Géorgie du Sud-et-les Îles Sandwich du Sud (la)", "GS", "SGS", "239"},
	{"Fiji", "Fidji (les)", "FJ", "FJI", "242"},
	{"Finland", "Finlande (la)", "FI", "FIN", "246"},
	{"Åland Islands", "Åland(les Îles)", "AX", "ALA", "248"},
	{"France", "France (la)", "FR", "FRA", "250"},
	{"French Guiana", "Guyane française (la )", "GF", "GUF", "254"},
	{"French Polynesia", "Polynésie française (la)", "PF", "PYF", "258"},
	{"French Southern Territories (the)", "Terres australes françaises (les)", "TF", "ATF", "260"},
	{"Djibouti", "Djibouti", "DJ", "DJI", "262"},
	{"Gabon", "Gabon (le)", "GA", "GAB", "266"},
	{"Georgia", "Géorgie (la)", "GE", "GEO", "268"},
	{"Gambia (the)", "Gambie (la)", "GM", "GMB", "270"},
	{"Palestine, State of", "Palestine, État de", "PS", "PSE", "275"},
	{"Germany", "Allemagne (l')", "DE", "DEU", "276"},
	{"Ghana", "Ghana (le)", "GH", "GHA", "288"},
	{"Gibraltar", "Gibraltar", "GI", "GIB", "292"},
	{"Kiribati", "Kiribati", "KI", "KIR", "296"},
	{"Greece", "Grèce (la)", "GR", "GRC", "300"},
	{"Greenland", "Groenland (le)", "GL", "GRL", "304"},
	{"Grenada", "Grenade (la)", "GD", "GRD", "308"},
	{"Guadeloupe", "Guadeloupe (la)", "GP", "GLP", "312"},
	{"Guam", "Guam", "GU", "GUM", "316"},
	{"Guatemala", "Guatemala (le)", "GT", "GTM", "320"},
	{"Guinea", "Guinée (la)", "GN", "GIN", "324"},
	{"Guyana", "Guyana (le)", "GY", "GUY", "328"},
	{"Haiti", "Haïti", "HT", "HTI", "332"},
	{"Heard Island and McDonald Islands", "Heard-et-Îles MacDonald (l'Île)", "HM", "HMD", "334"},
	{"Holy See (the)", "Saint-Siège (le)", "VA", "VAT", "336"},
	{"Honduras", "Honduras (le)", "HN", "HND", "340"},
	{"Hong Kong", "Hong Kong", "HK", "HKG", "344"},
	{"Hungary", "Hongrie (la)", "HU", "HUN", "348"},
	{"Iceland", "Islande (l')", "IS", "ISL", "352"},
	{"India", "Inde (l')", "IN", "IND", "356"},
	{"Indonesia", "Indonésie (l')", "ID", "IDN", "360"},
	{"Iran (Islamic Republic of)", "Iran (République Islamique d')", "IR", "IRN", "364"},
	{"Iraq", "Iraq (l')", "IQ", "IRQ", "368"},
	{"Ireland", "Irlande (l')", "IE", "IRL", "372"},
	{"Israel", "Israël", "IL", "ISR", "376"},
	{"Italy", "Italie (l')", "IT", "ITA", "380"},
	{"Côte d'Ivoire", "Côte d'Ivoire (la)", "CI", "CIV", "384"},
	{"Jamaica", "Jamaïque (la)", "JM", "JAM", "388"},
	{"Japan", "Japon (le)", "JP", "JPN", "392"},
	{"Kazakhstan", "Kazakhstan (le)", "KZ", "KAZ", "398"},
	{"Jordan", "Jordanie (la)", "JO", "JOR", "400"},
	{"Kenya", "Kenya (le)", "KE", "KEN", "404"},
	{"Korea (the Democratic People's Republic of)", "Corée (la République populaire démocratique de)", "KP", "PRK", "408"},
	{"Korea (the Republic of)", "Corée (la République de)", "KR", "KOR", "410"},
	{"Kuwait", "Koweït (le)", "KW", "KWT", "414"},
	{"Kyrgyzstan", "Kirghizistan (le)", "KG", "KGZ", "417"},
	{"Lao People's Democratic Republic (the)", "Lao, République démocratique populaire", "LA", "LAO", "418"},
	{"Lebanon", "Liban (le)", "LB", "LBN", "422"},
	{"Lesotho", "Lesotho (le)", "LS", "LSO", "426"},
	{"Latvia", "Lettonie (la)", "LV", "LVA", "428"},
	{"Liberia", "Libéria (le)", "LR", "LBR", "430"},
	{"Libya", "Libye (la)", "LY", "LBY", "434"},
	{"Liechtenstein", "Liechtenstein (le)", "LI", "LIE", "438"},
	{"Lithuania", "Lituanie (la)", "LT", "LTU", "440"},
	{"Luxembourg", "Luxembourg (le)", "LU", "LUX", "442"},
	{"Macao", "Macao", "MO", "MAC", "446"},
	{"Madagascar", "Madagascar", "MG", "MDG", "450"},
	{"Malawi", "Malawi (le)", "MW", "MWI", "454"},
	{"Malaysia", "Malaisie (la)", "MY", "MYS", "458"},
	{"Maldives", "Maldives (les)", "MV", "MDV", "462"},
	{"Mali", "Mali (le)", "ML", "MLI", "466"},
	{"Malta", "Malte", "MT", "MLT", "470"},
	{"Martinique", "Martinique (la)", "MQ", "MTQ", "474"},
	{"Mauritania", "Mauritanie (la)", "MR", "MRT", "478"},
	{"Mauritius", "Maurice", "MU", "MUS", "480"},
	{"Mexico", "Mexique (le)", "MX", "MEX", "484"},
	{"Monaco", "Monaco", "MC", "MCO", "492"},
	{"Mongolia", "Mongolie (la)", "MN", "MNG", "496"},
	{"Moldova (the Republic of)", "Moldova , République de", "MD", "MDA", "498"},
	{"Montenegro", "Monténégro (le)", "ME", "MNE", "499"},
	{"Montserrat", "Montserrat", "MS", "MSR", "500"},
	{"Morocco", "Maroc (le)", "MA", "MAR", "504"},
	{"Mozambique", "Mozambique (le)", "MZ", "MOZ", "508"},
	{"Oman", "Oman", "OM", "OMN", "512"},
	{"Namibia", "Namibie (la)", "NA", "NAM", "516"},
	{"Nauru", "Nauru", "NR", "NRU", "520"},
	{"Nepal", "Népal (le)", "NP", "NPL", "524"},
	{"Netherlands (the)", "Pays-Bas (les)", "NL", "NLD", "528"},
	{"Curaçao", "Curaçao", "CW", "CUW", "531"},
	{"Aruba", "Aruba", "AW", "ABW", "533"},
	{"Sint Maarten (Dutch part)", "Saint-Martin (partie néerlandaise)", "SX", "SXM", "534"},
	{"Bonaire, Sint Eustatius and Saba", "Bonaire, Saint-Eustache et Saba", "BQ", "BES", "535"},
	{"New Caledonia", "Nouvelle-Calédonie (la)", "NC", "NCL", "540"},
	{"Vanuatu", "Vanuatu (le)", "VU", "VUT", "548"},
	{"New Zealand", "Nouvelle-Zélande (la)", "NZ", "NZL", "554"},
	{"Nicaragua", "Nicaragua (le)", "NI", "NIC", "558"},
	{"Niger (the)", "Niger (le)", "NE", "NER", "562"},
	{"Nigeria", "Nigéria (le)", "NG", "NGA", "566"},
	{"Niue", "Niue", "NU", "NIU", "570"},
	{"Norfolk Island", "Norfolk (l'Île)", "NF", "NFK", "574"},
	{"Norway", "Norvège (la)", "NO", "NOR", "578"},
	{"Northern Mariana Islands (the)", "Mariannes du Nord (les Îles)", "MP", "MNP", "580"},
	{"United States Minor Outlying Islands (the)", "Îles mineures éloignées des États-Unis (les)", "UM", "UMI", "581"},
	{"Micronesia (Federated States of)", "Micronésie (États fédérés de)", "FM", "FSM", "583"},
	{"Marshall Islands (the)", "Marshall (Îles)", "MH", "MHL", "584"},
	{"Palau", "Palaos (les)", "PW", "PLW", "585"},
	{"Pakistan", "Pakistan (le)", "PK", "PAK", "586"},
	{"Panama", "Panama (le)", "PA", "PAN", "591"},
	{"Papua New Guinea", "Papouasie-Nouvelle-Guinée (la)", "PG", "PNG", "598"},
	{"Paraguay", "Paraguay (le)", "PY", "PRY", "600"},
	{"Peru", "Pérou (le)", "PE", "PER", "604"},
	{"Philippines (the)", "Philippines (les)", "PH", "PHL", "608"},
	{"Pitcairn", "Pitcairn", "PN", "PCN", "612"},
	{"Poland", "Pologne (la)", "PL", "POL", "616"},
	{"Portugal", "Portugal (le)", "PT", "PRT", "620"},
	{"Guinea-Bissau", "Guinée-Bissau (la)", "GW", "GNB", "624"},
	{"Timor-Leste", "Timor-Leste (le)", "TL", "TLS", "626"},
	{"Puerto Rico", "Porto Rico", "PR", "PRI", "630"},
	{"Qatar", "Qatar (le)", "QA", "QAT", "634"},
	{"Réunion", "Réunion (La)", "RE", "REU", "638"},
	{"Romania", "Roumanie (la)", "RO", "ROU", "642"},
	{"Russian Federation (the)", "Russie (la Fédération de)", "RU", "RUS", "643"},
	{"Rwanda", "Rwanda (le)", "RW", "RWA", "646"},
	{"Saint Barthélemy", "Saint-Barthélemy", "BL", "BLM", "652"},
	{"Saint Helena, Ascension and Tristan da Cunha", "Sainte-Hélène, Ascension et Tristan da Cunha", "SH", "SHN", "654"},
	{"Saint Kitts and Nevis", "Saint-Kitts-et-Nevis", "KN", "KNA", "659"},
	{"Anguilla", "Anguilla", "AI", "AIA", "660"},
	{"Saint Lucia", "Sainte-Lucie", "LC", "LCA", "662"},
	{"Saint Martin (French part)", "Saint-Martin (partie française)", "MF", "MAF", "663"},
	{"Saint Pierre and Miquelon", "Saint-Pierre-et-Miquelon", "PM", "SPM", "666"},
	{"Saint Vincent and the Grenadines", "Saint-Vincent-et-les Grenadines", "VC", "VCT", "670"},
	{"San Marino", "Saint-Marin", "SM", "SMR", "674"},
	{"Sao Tome and Principe", "Sao Tomé-et-Principe", "ST", "STP", "678"},
	{"Saudi Arabia", "Arabie saoudite (l')", "SA", "SAU", "682"},
	{"Senegal", "Sénégal (le)", "SN", "SEN", "686"},
	{"Serbia", "Serbie (la)", "RS", "SRB", "688"},
	{"Seychelles", "Seychelles (les)", "SC", "SYC", "690"},
	{"Sierra Leone", "Sierra Leone (la)", "SL", "SLE", "694"},
	{"Singapore", "Singapour", "SG", "SGP", "702"},
	{"Slovakia", "Slovaquie (la)", "SK", "SVK", "703"},
	{"Viet Nam", "Viet Nam (le)", "VN", "VNM", "704"},
	{"Slovenia", "Slovénie (la)", "SI", "SVN", "705"},
	{"Somalia", "Somalie (la)", "SO", "SOM", "706"},
	{"South Africa", "Afrique du Sud (l')", "ZA", "ZAF", "710"},
	{"Zimbabwe", "Zimbabwe (le)", "ZW", "ZWE", "716"},
	{"Spain", "Espagne (l')", "ES", "ESP", "724"},
	{"South Sudan", "Soudan du Sud (le)", "SS", "SSD", "728"},
	{"Sudan (the)", "Soudan (le)", "SD", "SDN", "729"},
	{"Western Sahara*", "Sahara occidental (le)*", "EH", "ESH", "732"},
	{"Suriname", "Suriname (le)", "SR", "SUR", "740"},
	{"Svalbard and Jan Mayen", "Svalbard et l'Île Jan Mayen (le)", "SJ", "SJM", "744"},
	{"Swaziland", "Swaziland (le)", "SZ", "SWZ", "748"},
	{"Sweden", "Suède (la)", "SE", "SWE", "752"},
	{"Switzerland", "Suisse (la)", "CH", "CHE", "756"},
	{"Syrian Arab Republic", "République arabe syrienne (la)", "SY", "SYR", "760"},
	{"Tajikistan", "Tadjikistan (le)", "TJ", "TJK", "762"},
	{"Thailand", "Thaïlande (la)", "TH", "THA", "764"},
	{"Togo", "Togo (le)", "TG", "TGO", "768"},
	{"Tokelau", "Tokelau (les)", "TK", "TKL", "772"},
	{"Tonga", "Tonga (les)", "TO", "TON", "776"},
	{"Trinidad and Tobago", "Trinité-et-Tobago (la)", "TT", "TTO", "780"},
	{"United Arab Emirates (the)", "Émirats arabes unis (les)", "AE", "ARE", "784"},
	{"Tunisia", "Tunisie (la)", "TN", "TUN", "788"},
	{"Turkey", "Turquie (la)", "TR", "TUR", "792"},
	{"Turkmenistan", "Turkménistan (le)", "TM", "TKM", "795"},
	{"Turks and Caicos Islands (the)", "Turks-et-Caïcos (les Îles)", "TC", "TCA", "796"},
	{"Tuvalu", "Tuvalu (les)", "TV", "TUV", "798"},
	{"Uganda", "Ouganda (l')", "UG", "UGA", "800"},
	{"Ukraine", "Ukraine (l')", "UA", "UKR", "804"},
	{"Macedonia (the former Yugoslav Republic of)", "Macédoine (l'ex‑République yougoslave de)", "MK", "MKD", "807"},
	{"Egypt", "Égypte (l')", "EG", "EGY", "818"},
	{"United Kingdom of Great Britain and Northern Ireland (the)", "Royaume-Uni de Grande-Bretagne et d'Irlande du Nord (le)", "GB", "GBR", "826"},
	{"Guernsey", "Guernesey", "GG", "GGY", "831"},
	{"Jersey", "Jersey", "JE", "JEY", "832"},
	{"Isle of Man", "Île de Man", "IM", "IMN", "833"},
	{"Tanzania, United Republic of", "Tanzanie, République-Unie de", "TZ", "TZA", "834"},
	{"United States of America (the)", "États-Unis d'Amérique (les)", "US", "USA", "840"},
	{"Virgin Islands (U.S.)", "Vierges des États-Unis (les Îles)", "VI", "VIR", "850"},
	{"Burkina Faso", "Burkina Faso (le)", "BF", "BFA", "854"},
	{"Uruguay", "Uruguay (l')", "UY", "URY", "858"},
	{"Uzbekistan", "Ouzbékistan (l')", "UZ", "UZB", "860"},
	{"Venezuela (Bolivarian Republic of)", "Venezuela (République bolivarienne du)", "VE", "VEN", "862"},
	{"Wallis and Futuna", "Wallis-et-Futuna", "WF", "WLF", "876"},
	{"Samoa", "Samoa (le)", "WS", "WSM", "882"},
	{"Yemen", "Yémen (le)", "YE", "YEM", "887"},
	{"Zambia", "Zambie (la)", "ZM", "ZMB", "894"},
}

// ISO4217List is the list of ISO currency codes
var ISO4217List = []string{
	"AED", "AFN", "ALL", "AMD", "ANG", "AOA", "ARS", "AUD", "AWG", "AZN",
	"BAM", "BBD", "BDT", "BGN", "BHD", "BIF", "BMD", "BND", "BOB", "BOV", "BRL", "BSD", "BTN", "BWP", "BYN", "BZD",
	"CAD", "CDF", "CHE", "CHF", "CHW", "CLF", "CLP", "CNY", "COP", "COU", "CRC", "CUC", "CUP", "CVE", "CZK",
	"DJF", "DKK", "DOP", "DZD",
	"EGP", "ERN", "ETB", "EUR",
	"FJD", "FKP",
	"GBP", "GEL", "GHS", "GIP", "GMD", "GNF", "GTQ", "GYD",
	"HKD", "HNL", "HRK", "HTG", "HUF",
	"IDR", "ILS", "INR", "IQD", "IRR", "ISK",
	"JMD", "JOD", "JPY",
	"KES", "KGS", "KHR", "KMF", "KPW", "KRW", "KWD", "KYD", "KZT",
	"LAK", "LBP", "LKR", "LRD", "LSL", "LYD",
	"MAD", "MDL", "MGA", "MKD", "MMK", "MNT", "MOP", "MRO", "MUR", "MVR", "MWK", "MXN", "MXV", "MYR", "MZN",
	"NAD", "NGN", "NIO", "NOK", "NPR", "NZD",
	"OMR",
	"PAB", "PEN", "PGK", "PHP", "PKR", "PLN", "PYG",
	"QAR",
	"RON", "RSD", "RUB", "RWF",
	"SAR", "SBD", "SCR", "SDG", "SEK", "SGD", "SHP", "SLL", "SOS", "SRD", "SSP", "STD", "SVC", "SYP", "SZL",
	"THB", "TJS", "TMT", "TND", "TOP", "TRY", "TTD", "TWD", "TZS",
	"UAH", "UGX", "USD", "USN", "UYI", "UYU", "UZS",
	"VEF", "VND", "VUV",
	"WST",
	"XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XOF", "XPD", "XPF", "XPT", "XSU", "XTS", "XUA", "XXX",
	"YER",
	"ZAR", "ZMW", "ZWL",
}

// ISO693Entry stores ISO language codes
type ISO693Entry struct {
	Alpha3bCode string
	Alpha2Code  string
	English     string
}

//ISO693List based on http://data.okfn.org/data/core/language-codes/r/language-codes-3b2.json
var ISO693List = []ISO693Entry{
	{Alpha3bCode: "aar", Alpha2Code: "aa", English: "Afar"},
	{Alpha3bCode: "abk", Alpha2Code: "ab", English: "Abkhazian"},
	{Alpha3bCode: "afr", Alpha2Code: "af", English: "Afrikaans"},
	{Alpha3bCode: "aka", Alpha2Code: "ak", English: "Akan"},
	{Alpha3bCode: "alb", Alpha2Code: "sq", English: "Albanian"},
	{Alpha3bCode: "amh", Alpha2Code: "am", English: "Amharic"},
	{Alpha3bCode: "ara", Alpha2Code: "ar", English: "Arabic"},
	{Alpha3bCode: "arg", Alpha2Code: "an", English: "Aragonese"},
	{Alpha3bCode: "arm", Alpha2Code: "hy", English: "Armenian"},
	{Alpha3bCode: "asm", Alpha2Code: "as", English: "Assamese"},
	{Alpha3bCode: "ava", Alpha2Code: "av", English: "Avaric"},
	{Alpha3bCode: "ave", Alpha2Code: "ae", English: "Avestan"},
	{Alpha3bCode: "aym", Alpha2Code: "ay", English: "Aymara"},
	{Alpha3bCode: "aze", Alpha2Code: "az", English: "Azerbaijani"},
	{Alpha3bCode: "bak", Alpha2Code: "ba", English: "Bashkir"},
	{Alpha3bCode: "bam", Alpha2Code: "bm", English: "Bambara"},
	{Alpha3bCode: "baq", Alpha2Code: "eu", English: "Basque"},
	{Alpha3bCode: "bel", Alpha2Code: "be", English: "Belarusian"},
	{Alpha3bCode: "ben", Alpha2Code: "bn", English: "Bengali"},
	{Alpha3bCode: "bih", Alpha2Code: "bh", English: "Bihari languages"},
	{Alpha3bCode: "bis", Alpha2Code: "bi", English: "Bislama"},
	{Alpha3bCode: "bos", Alpha2Code: "bs", English: "Bosnian"},
	{Alpha3bCode: "bre", Alpha2Code: "br", English: "Breton"},
	{Alpha3bCode: "bul", Alpha2Code: "bg", English: "Bulgarian"},
	{Alpha3bCode: "bur", Alpha2Code: "my", English: "Burmese"},
	{Alpha3bCode: "cat", Alpha2Code: "ca", English: "Catalan; Valencian"},
	{Alpha3bCode: "cha", Alpha2Code: "ch", English: "Chamorro"},
	{Alpha3bCode: "che", Alpha2Code: "ce", English: "Chechen"},
	{Alpha3bCode: "chi", Alpha2Code: "zh", English: "Chinese"},
	{Alpha3bCode: "chu", Alpha2Code: "cu", English: "Church Slavic; Old Slavonic; Church Slavonic; Old Bulgarian; Old Church Slavonic"},
	{Alpha3bCode: "chv", Alpha2Code: "cv", English: "Chuvash"},
	{Alpha3bCode: "cor", Alpha2Code: "kw", English: "Cornish"},
	{Alpha3bCode: "cos", Alpha2Code: "co", English: "Corsican"},
	{Alpha3bCode: "cre", Alpha2Code: "cr", English: "Cree"},
	{Alpha3bCode: "cze", Alpha2Code: "cs", English: "Czech"},
	{Alpha3bCode: "dan", Alpha2Code: "da", English: "Danish"},
	{Alpha3bCode: "div", Alpha2Code: "dv", English: "Divehi; Dhivehi; Maldivian"},
	{Alpha3bCode: "dut", Alpha2Code: "nl", English: "Dutch; Flemish"},
	{Alpha3bCode: "dzo", Alpha2Code: "dz", English: "Dzongkha"},
	{Alpha3bCode: "eng", Alpha2Code: "en", English: "English"},
	{Alpha3bCode: "epo", Alpha2Code: "eo", English: "Esperanto"},
	{Alpha3bCode: "est", Alpha2Code: "et", English: "Estonian"},
	{Alpha3bCode: "ewe", Alpha2Code: "ee", English: "Ewe"},
	{Alpha3bCode: "fao", Alpha2Code: "fo", English: "Faroese"},
	{Alpha3bCode: "fij", Alpha2Code: "fj", English: "Fijian"},
	{Alpha3bCode: "fin", Alpha2Code: "fi", English: "Finnish"},
	{Alpha3bCode: "fre", Alpha2Code: "fr", English: "French"},
	{Alpha3bCode: "fry", Alpha2Code: "fy", English: "Western Frisian"},
	{Alpha3bCode: "ful", Alpha2Code: "ff", English: "Fulah"},
	{Alpha3bCode: "geo", Alpha2Code: "ka", English: "Georgian"},
	{Alpha3bCode: "ger", Alpha2Code: "de", English: "German"},
	{Alpha3bCode: "gla", Alpha2Code: "gd", English: "Gaelic; Scottish Gaelic"},
	{Alpha3bCode: "gle", Alpha2Code: "ga", English: "Irish"},
	{Alpha3bCode: "glg", Alpha2Code: "gl", English: "Galician"},
	{Alpha3bCode: "glv", Alpha2Code: "gv", English: "Manx"},
	{Alpha3bCode: "gre", Alpha2Code: "el", English: "Greek, Modern (1453-)"},
	{Alpha3bCode: "grn", Alpha2Code: "gn", English: "Guarani"},
	{Alpha3bCode: "guj", Alpha2Code: "gu", English: "Gujarati"},
	{Alpha3bCode: "hat", Alpha2Code: "ht", English: "Haitian; Haitian Creole"},
	{Alpha3bCode: "hau", Alpha2Code: "ha", English: "Hausa"},
	{Alpha3bCode: "heb", Alpha2Code: "he", English: "Hebrew"},
	{Alpha3bCode: "her", Alpha2Code: "hz", English: "Herero"},
	{Alpha3bCode: "hin", Alpha2Code: "hi", English: "Hindi"},
	{Alpha3bCode: "hmo", Alpha2Code: "ho", English: "Hiri Motu"},
	{Alpha3bCode: "hrv", Alpha2Code: "hr", English: "Croatian"},
	{Alpha3bCode: "hun", Alpha2Code: "hu", English: "Hungarian"},
	{Alpha3bCode: "ibo", Alpha2Code: "ig", English: "Igbo"},
	{Alpha3bCode: "ice", Alpha2Code: "is", English: "Icelandic"},
	{Alpha3bCode: "ido", Alpha2Code: "io", English: "Ido"},
	{Alpha3bCode: "iii", Alpha2Code: "ii", English: "Sichuan Yi; Nuosu"},
	{Alpha3bCode: "iku", Alpha2Code: "iu", English: "Inuktitut"},
	{Alpha3bCode: "ile", Alpha2Code: "ie", English: "Interlingue; Occidental"},
	{Alpha3bCode: "ina", Alpha2Code: "ia", English: "Interlingua (International Auxiliary Language Association)"},
	{Alpha3bCode: "ind", Alpha2Code: "id", English: "Indonesian"},
	{Alpha3bCode: "ipk", Alpha2Code: "ik", English: "Inupiaq"},
	{Alpha3bCode: "ita", Alpha2Code: "it", English: "Italian"},
	{Alpha3bCode: "jav", Alpha2Code: "jv", English: "Javanese"},
	{Alpha3bCode: "jpn", Alpha2Code: "ja", English: "Japanese"},
	{Alpha3bCode: "kal", Alpha2Code: "kl", English: "Kalaallisut; Greenlandic"},
	{Alpha3bCode: "kan", Alpha2Code: "kn", English: "Kannada"},
	{Alpha3bCode: "kas", Alpha2Code: "ks", English: "Kashmiri"},
	{Alpha3bCode: "kau", Alpha2Code: "kr", English: "Kanuri"},
	{Alpha3bCode: "kaz", Alpha2Code: "kk", English: "Kazakh"},
	{Alpha3bCode: "khm", Alpha2Code: "km", English: "Central Khmer"},
	{Alpha3bCode: "kik", Alpha2Code: "ki", English: "Kikuyu; Gikuyu"},
	{Alpha3bCode: "kin", Alpha2Code: "rw", English: "Kinyarwanda"},
	{Alpha3bCode: "kir", Alpha2Code: "ky", English: "Kirghiz; Kyrgyz"},
	{Alpha3bCode: "kom", Alpha2Code: "kv", English: "Komi"},
	{Alpha3bCode: "kon", Alpha2Code: "kg", English: "Kongo"},
	{Alpha3bCode: "kor", Alpha2Code: "ko", English: "Korean"},
	{Alpha3bCode: "kua", Alpha2Code: "kj", English: "Kuanyama; Kwanyama"},
	{Alpha3bCode: "kur", Alpha2Code: "ku", English: "Kurdish"},
	{Alpha3bCode: "lao", Alpha2Code: "lo", English: "Lao"},
	{Alpha3bCode: "lat", Alpha2Code: "la", English: "Latin"},
	{Alpha3bCode: "lav", Alpha2Code: "lv", English: "Latvian"},
	{Alpha3bCode: "lim", Alpha2Code: "li", English: "Limburgan; Limburger; Limburgish"},
	{Alpha3bCode: "lin", Alpha2Code: "ln", English: "Lingala"},
	{Alpha3bCode: "lit", Alpha2Code: "lt", English: "Lithuanian"},
	{Alpha3bCode: "ltz", Alpha2Code: "lb", English: "Luxembourgish; Letzeburgesch"},
	{Alpha3bCode: "lub", Alpha2Code: "lu", English: "Luba-Katanga"},
	{Alpha3bCode: "lug", Alpha2Code: "lg", English: "Ganda"},
	{Alpha3bCode: "mac", Alpha2Code: "mk", English: "Macedonian"},
	{Alpha3bCode: "mah", Alpha2Code: "mh", English: "Marshallese"},
	{Alpha3bCode: "mal", Alpha2Code: "ml", English: "Malayalam"},
	{Alpha3bCode: "mao", Alpha2Code: "mi", English: "Maori"},
	{Alpha3bCode: "mar", Alpha2Code: "mr", English: "Marathi"},
	{Alpha3bCode: "may", Alpha2Code: "ms", English: "Malay"},
	{Alpha3bCode: "mlg", Alpha2Code: "mg", English: "Malagasy"},
	{Alpha3bCode: "mlt", Alpha2Code: "mt", English: "Maltese"},
	{Alpha3bCode: "mon", Alpha2Code: "mn", English: "Mongolian"},
	{Alpha3bCode: "nau", Alpha2Code: "na", English: "Nauru"},
	{Alpha3bCode: "nav", Alpha2Code: "nv", English: "Navajo; Navaho"},
	{Alpha3bCode: "nbl", Alpha2Code: "nr", English: "Ndebele, South; South Ndebele"},
	{Alpha3bCode: "nde", Alpha2Code: "nd", English: "Ndebele, North; North Ndebele"},
	{Alpha3bCode: "ndo", Alpha2Code: "ng", English: "Ndonga"},
	{Alpha3bCode: "nep", Alpha2Code: "ne", English: "Nepali"},
	{Alpha3bCode: "nno", Alpha2Code: "nn", English: "Norwegian Nynorsk; Nynorsk, Norwegian"},
	{Alpha3bCode: "nob", Alpha2Code: "nb", English: "Bokmål, Norwegian; Norwegian Bokmål"},
	{Alpha3bCode: "nor", Alpha2Code: "no", English: "Norwegian"},
	{Alpha3bCode: "nya", Alpha2Code: "ny", English: "Chichewa; Chewa; Nyanja"},
	{Alpha3bCode: "oci", Alpha2Code: "oc", English: "Occitan (post 1500); Provençal"},
	{Alpha3bCode: "oji", Alpha2Code: "oj", English: "Ojibwa"},
	{Alpha3bCode: "ori", Alpha2Code: "or", English: "Oriya"},
	{Alpha3bCode: "orm", Alpha2Code: "om", English: "Oromo"},
	{Alpha3bCode: "oss", Alpha2Code: "os", English: "Ossetian; Ossetic"},
	{Alpha3bCode: "pan", Alpha2Code: "pa", English: "Panjabi; Punjabi"},
	{Alpha3bCode: "per", Alpha2Code: "fa", English: "Persian"},
	{Alpha3bCode: "pli", Alpha2Code: "pi", English: "Pali"},
	{Alpha3bCode: "pol", Alpha2Code: "pl", English: "Polish"},
	{Alpha3bCode: "por", Alpha2Code: "pt", English: "Portuguese"},
	{Alpha3bCode: "pus", Alpha2Code: "ps", English: "Pushto; Pashto"},
	{Alpha3bCode: "que", Alpha2Code: "qu", English: "Quechua"},
	{Alpha3bCode: "roh", Alpha2Code: "rm", English: "Romansh"},
	{Alpha3bCode: "rum", Alpha2Code: "ro", English: "Romanian; Moldavian; Moldovan"},
	{Alpha3bCode: "run", Alpha2Code: "rn", English: "Rundi"},
	{Alpha3bCode: "rus", Alpha2Code: "ru", English: "Russian"},
	{Alpha3bCode: "sag", Alpha2Code: "sg", English: "Sango"},
	{Alpha3bCode: "san", Alpha2Code: "sa", English: "Sanskrit"},
	{Alpha3bCode: "sin", Alpha2Code: "si", English: "Sinhala; Sinhalese"},
	{Alpha3bCode: "slo", Alpha2Code: "sk", English: "Slovak"},
	{Alpha3bCode: "slv", Alpha2Code: "sl", English: "Slovenian"},
	{Alpha3bCode: "sme", Alpha2Code: "se", English: "Northern Sami"},
	{Alpha3bCode: "smo", Alpha2Code: "sm", English: "Samoan"},
	{Alpha3bCode: "sna", Alpha2Code: "sn", English: "Shona"},
	{Alpha3bCode: "snd", Alpha2Code: "sd", English: "Sindhi"},
	{Alpha3bCode: "som", Alpha2Code: "so", English: "Somali"},
	{Alpha3bCode: "sot", Alpha2Code: "st", English: "Sotho, Southern"},
	{Alpha3bCode: "spa", Alpha2Code: "es", English: "Spanish; Castilian"},
	{Alpha3bCode: "srd", Alpha2Code: "sc", English: "Sardinian"},
	{Alpha3bCode: "srp", Alpha2Code: "sr", English: "Serbian"},
	{Alpha3bCode: "ssw", Alpha2Code: "ss", English: "Swati"},
	{Alpha3bCode: "sun", Alpha2Code: "su", English: "Sundanese"},
	{Alpha3bCode: "swa", Alpha2Code: "sw", English: "Swahili"},
	{Alpha3bCode: "swe", Alpha2Code: "sv", English: "Swedish"},
	{Alpha3bCode: "tah", Alpha2Code: "ty", English: "Tahitian"},
	{Alpha3bCode: "tam", Alpha2Code: "ta", English: "Tamil"},
	{Alpha3bCode: "tat", Alpha2Code: "tt", English: "Tatar"},
	{Alpha3bCode: "tel", Alpha2Code: "te", English: "Telugu"},
	{Alpha3bCode: "tgk", Alpha2Code: "tg", English: "Tajik"},
	{Alpha3bCode: "tgl", Alpha2Code: "tl", English: "Tagalog"},
	{Alpha3bCode: "tha", Alpha2Code: "th", English: "Thai"},
	{Alpha3bCode: "tib", Alpha2Code: "bo", English: "Tibetan"},
	{Alpha3bCode: "tir", Alpha2Code: "ti", English: "Tigrinya"},
	{Alpha3bCode: "ton", Alpha2Code: "to", English: "Tonga (Tonga Islands)"},
	{Alpha3bCode: "tsn", Alpha2Code: "tn", English: "Tswana"},
	{Alpha3bCode: "tso", Alpha2Code: "ts", English: "Tsonga"},
	{Alpha3bCode: "tuk", Alpha2Code: "tk", English: "Turkmen"},
	{Alpha3bCode: "tur", Alpha2Code: "tr", English: "Turkish"},
	{Alpha3bCode: "twi", Alpha2Code: "tw", English: "Twi"},
	{Alpha3bCode: "uig", Alpha2Code: "ug", English: "Uighur; Uyghur"},
	{Alpha3bCode: "ukr", Alpha2Code: "uk", English: "Ukrainian"},
	{Alpha3bCode: "urd", Alpha2Code: "ur", English: "Urdu"},
	{Alpha3bCode: "uzb", Alpha2Code: "uz", English: "Uzbek"},
	{Alpha3bCode: "ven", Alpha2Code: "ve", English: "Venda"},
	{Alpha3bCode: "vie", Alpha2Code: "vi", English: "Vietnamese"},
	{Alpha3bCode: "vol", Alpha2Code: "vo", English: "Volapük"},
	{Alpha3bCode: "wel", Alpha2Code: "cy", English: "Welsh"},
	{Alpha3bCode: "wln", Alpha2Code: "wa", English: "Walloon"},
	{Alpha3bCode: "wol", Alpha2Code: "wo", English: "Wolof"},
	{Alpha3bCode: "xho", Alpha2Code: "xh", English: "Xhosa"},
	{Alpha3bCode: "yid", Alpha2Code: "yi", English: "Yiddish"},
	{Alpha3bCode: "yor", Alpha2Code: "yo", English: "Yoruba"},
	{Alpha3bCode: "zha", Alpha2Code: "za", English: "Zhuang; Chuang"},
	{Alpha3bCode: "zul", Alpha2Code: "zu", English: "Zulu"},
}
