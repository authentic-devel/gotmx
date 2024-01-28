package gotmx

// booleanHTMLAttributes lists HTML attributes that are boolean according to the HTML spec.
// Boolean attributes are rendered without a value (e.g., <button disabled> instead of
// <button disabled="true">) when their resolved value is empty or "true".
//
// Reference: https://html.spec.whatwg.org/multipage/indices.html#attributes-3
var booleanHTMLAttributes = map[string]bool{
	"allowfullscreen": true,
	"async":           true,
	"autofocus":       true,
	"autoplay":        true,
	"checked":         true,
	"controls":        true,
	"default":         true,
	"defer":           true,
	"disabled":        true,
	"formnovalidate":  true,
	"hidden":          true,
	"inert":           true,
	"ismap":           true,
	"itemscope":       true,
	"loop":            true,
	"multiple":        true,
	"muted":           true,
	"nomodule":        true,
	"novalidate":      true,
	"open":            true,
	"playsinline":     true,
	"readonly":        true,
	"required":        true,
	"reversed":        true,
	"selected":        true,
}
