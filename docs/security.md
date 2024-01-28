# Security & HTML Escaping

Gotmx escapes HTML by default to prevent XSS (Cross-Site Scripting) attacks. Understanding the escaping rules for each attribute ensures your templates handle user input safely.

## Escaping Rules

Gotmx attributes fall into three categories based on how they handle HTML escaping. Each category exists for a specific reason.

### Always Safe (use with untrusted input)

| Attribute | Escaping | Why |
|-----------|----------|-----|
| `data-g-inner-text` | Always escaped | This is the **recommended default** for displaying any text content, especially user input. It escapes unconditionally — even `Unescaped()` cannot disable it. This guarantee means you never have to think about whether a value is safe. |
| `data-g-att-*` | Always escaped | Attribute values must always be escaped to prevent attribute injection attacks. There is no scenario where unescaped attribute values are safe. |

### Context-Dependent (follows the `Escaped` flag)

| Attribute | Escaping | Why |
|-----------|----------|-----|
| `data-g-outer-text` | Follows `ctx.Escaped` | Replaces the entire element with text. Because it replaces the element itself (rather than injecting content into it), it participates in the global escaping setting. This allows `Unescaped()` to affect the output when rendering fully trusted content. By default (escaped mode), it is safe for user input. |
| Text nodes with `[[ .Path ]]` | Follows `ctx.Escaped` | Same reasoning as `data-g-outer-text` — text nodes in the template body follow the global setting. Safe by default. |

### Never Escapes (use only with trusted/sanitized content)

| Attribute | Escaping | Why |
|-----------|----------|-----|
| `data-g-inner-html` | Never escaped | Explicitly designed for injecting pre-sanitized HTML. The name "inner-html" signals that the value is treated as raw HTML. |
| `data-g-as-unsafe-template` | No escaping | Uses Go's `text/template`, which has no auto-escaping. The word "unsafe" in the name is intentional. |

### Why `data-g-inner-text` and `data-g-outer-text` Differ

This is the most common source of confusion. The difference is intentional:

- **`data-g-inner-text`** is the primary attribute for displaying user-facing text. It is designed to be unconditionally safe — no configuration, no flags, no footguns. When you use `data-g-inner-text`, you never have to worry about the escaping state.
- **`data-g-outer-text`** replaces the entire element and is typically used for computed values or template references rendered to strings. It follows the global `Escaped` flag because there are legitimate use cases where the caller controls both the data and the escaping context.

**Rule of thumb:** When in doubt, use `data-g-inner-text`. It is always safe.

## Safe Patterns for User Input

### Always use `data-g-inner-text` for user-generated content

```html
<!-- SAFE: inner-text always escapes, even with Unescaped() -->
<span data-g-inner-text="[[ .UserComment ]]">preview</span>
```

When `UserComment` contains `<script>alert('xss')</script>`:

```html
<!-- Output (default): -->
<span>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</span>

<!-- Output (with Unescaped()): SAME — still escaped -->
<span>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</span>
```

### Use `data-g-att-*` for dynamic attribute values

```html
<!-- SAFE: attribute values are always escaped -->
<a data-g-att-href="[[ .UserUrl ]]"
   data-g-att-title="[[ .UserBio ]]">
    Profile
</a>
```

When `UserUrl` contains `javascript:alert('xss')` or `UserBio` contains `" onclick="alert('xss')`:

```html
<!-- Output: attributes are escaped, injection is neutralized -->
<a href="javascript:alert(&#39;xss&#39;)"
   title="&#34; onclick=&#34;alert(&#39;xss&#39;)">
    Profile
</a>
```

Attribute values are always HTML-escaped, preventing attribute injection attacks.

### Never use `data-g-inner-html` with untrusted content

```html
<!-- DANGEROUS: inner-html never escapes -->
<div data-g-inner-html="[[ .Content ]]">preview</div>
```

When `Content` contains `<script>alert('xss')</script>`:

```html
<!-- Output: the script tag is rendered as-is — XSS vulnerability! -->
<div><script>alert('xss')</script></div>
```

If you must render HTML from user input, sanitize it server-side first (e.g., with [bluemonday](https://github.com/microcosm-cc/bluemonday)).

### Side-by-side: `data-g-inner-text` vs `data-g-outer-text` vs `data-g-inner-html`

Given `data = {"Value": "<b>Hello</b>"}`:

```html
<!-- data-g-inner-text: always escapes -->
<span data-g-inner-text="[[ .Value ]]">preview</span>
<!-- Default:      <span>&lt;b&gt;Hello&lt;/b&gt;</span>  -->
<!-- Unescaped():  <span>&lt;b&gt;Hello&lt;/b&gt;</span>  (same!) -->

<!-- data-g-outer-text: follows global setting -->
<span data-g-outer-text="[[ .Value ]]">preview</span>
<!-- Default:      &lt;b&gt;Hello&lt;/b&gt;               (escaped, element replaced) -->
<!-- Unescaped():  <b>Hello</b>                            (NOT escaped!) -->

<!-- data-g-inner-html: never escapes -->
<span data-g-inner-html="[[ .Value ]]">preview</span>
<!-- Default:      <span><b>Hello</b></span>               (raw HTML) -->
<!-- Unescaped():  <span><b>Hello</b></span>               (same!) -->
```

## The `Unescaped()` Option

By default, all rendering escapes HTML. The `Unescaped()` render option disables escaping for text nodes and `data-g-outer-text`, but critically:

- `data-g-inner-text` **still escapes** (unconditionally safe)
- Attribute values **still escape** (unconditionally safe)
- `data-g-inner-html` **never escapes** regardless

```go
// Default: escaped
engine.Render(ctx, w, "template", data)

// Unescaped: only affects text nodes and g-outer-text
engine.Render(ctx, w, "template", trustedData, gotmx.Unescaped())
```

Use `Unescaped()` only when you control the data source entirely (e.g., rendering a static page with no user input).

## Go Template Integration

When using inline Go templates:

```html
<!-- SAFE: data-g-as-template uses html/template (auto-escaped) -->
<div data-g-as-template="">Hello, {{ .Name }}!</div>

<!-- DANGEROUS: data-g-as-unsafe-template uses text/template (no escaping) -->
<div data-g-as-unsafe-template="">{{ .TrustedHtml }}</div>
```

Prefer `data-g-as-template` over `data-g-as-unsafe-template` unless you specifically need unescaped output from trusted sources.

## Recommendations

1. **Default to `data-g-inner-text`** for displaying any data. It is the safest option and cannot be misconfigured.
2. **Use `data-g-att-*`** for dynamic attributes. They are always escaped.
3. **Use `data-g-inner-html`** only for content you have sanitized server-side (e.g., with [bluemonday](https://github.com/microcosm-cc/bluemonday)).
4. **Never pass `Unescaped()`** when rendering pages with user-generated content.
5. **Sanitize on the server**, not in templates — templates are the last line of defense, not the first.
6. **Use `data-g-as-template`** (not `data-g-as-unsafe-template`) for inline Go templates that include user data.
