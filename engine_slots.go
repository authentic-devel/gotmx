package gotmx

import "fmt"

// Slots is a map of slot names to content.
// Content can be:
//   - string: rendered as HTML content
//   - Renderable: rendered using its Render method
//   - []Renderable: all items rendered in sequence
//
// Slots supports a fluent API for building slot content:
//
//	slots := gotmx.Slots{}.
//	    Set("header", headerComponent).
//	    Set("content", bodyComponent).
//	    Add("sidebar", widget1).
//	    Add("sidebar", widget2)
type Slots map[string]any

// Set replaces the content of a named slot. Returns the Slots map for chaining.
func (s Slots) Set(name string, content any) Slots {
	s[name] = content
	return s
}

// Add appends a Renderable to a named slot. Unlike Set, this preserves existing
// content in the slot. Returns the Slots map for chaining.
func (s Slots) Add(name string, r Renderable) Slots {
	existing, ok := s[name]
	if !ok {
		s[name] = r
		return s
	}
	switch v := existing.(type) {
	case []Renderable:
		s[name] = append(v, r)
	case Renderable:
		s[name] = []Renderable{v, r}
	default:
		// Convert existing to renderable first, then combine
		s[name] = []Renderable{toRenderables(existing)[0], r}
	}
	return s
}

// WithSlots provides slot content for the render.
// Slot values can be:
//   - string: rendered as HTML content
//   - Renderable: rendered using its Render method
//   - []Renderable: all items rendered in sequence
func WithSlots(slots Slots) RenderOption {
	return func(cfg *renderConfig) {
		if cfg.slots == nil {
			cfg.slots = make(SlottedRenderables)
		}
		for name, content := range slots {
			cfg.slots[name] = toRenderables(content)
		}
	}
}

// Slot is a convenience function for creating a single-slot Slots map.
func Slot(name string, content any) RenderOption {
	return WithSlots(Slots{name: content})
}

// toRenderables converts various types to []Renderable.
func toRenderables(content any) []Renderable {
	switch v := content.(type) {
	case string:
		tpl := NewStringLiteralTemplate("slot-content", v, "")
		r, _ := tpl.NewRenderable(nil)
		return []Renderable{r}
	case Renderable:
		return []Renderable{v}
	case []Renderable:
		return v
	default:
		// Convert to string as fallback
		tpl := NewStringLiteralTemplate("slot-content", fmt.Sprintf("%v", v), "")
		r, _ := tpl.NewRenderable(nil)
		return []Renderable{r}
	}
}
