# Using Gotmx with HTMX

Gotmx works naturally with [HTMX](https://htmx.org/) because both operate on standard HTML. HTMX attributes (`hx-*` or `data-hx-*`) coexist with gotmx attributes (`g-*` or `data-g-*`) on the same elements with no conflicts.

## Key Concepts

### Full Page vs Fragment Rendering

HTMX requests typically expect an HTML fragment, not a full page. Use the `HX-Request` header to detect HTMX requests and render accordingly:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    data := loadData()

    if r.Header.Get("HX-Request") == "true" {
        // HTMX request: render just the fragment
        engine.Render(r.Context(), w, "user-list", data)
    } else {
        // Full page load: wrap in layout
        engine.Render(r.Context(), w, "user-list", data,
            gotmx.WithLayout("main-layout", layoutData),
        )
    }
}
```

### Templates with HTMX Attributes

HTMX attributes pass through gotmx untouched. Use `data-hx-*` for HTML validity:

```html
<div data-g-define="user-list">
    <button data-hx-get="/users"
            data-hx-target="#user-table"
            data-hx-swap="innerHTML">
        Refresh
    </button>
    <table id="user-table">
        <tbody data-g-inner-repeat="[[ .Users ]]">
            <tr>
                <td data-g-inner-text="[[ .Name ]]">John</td>
                <td data-g-inner-text="[[ .Email ]]">john@example.com</td>
            </tr>
        </tbody>
    </table>
</div>
```

### Dynamic HTMX Attributes

Use `data-g-att-*` to set HTMX attributes dynamically from model data:

```html
<button data-g-att-data-hx-delete="[[ .DeleteUrl ]]"
        data-hx-target="#item-list"
        data-hx-swap="outerHTML"
        data-hx-confirm="Are you sure?">
    Delete
</button>
```

### Conditional HTMX Behavior

Use `data-g-attif-*` to conditionally add HTMX attributes:

```html
<div data-g-attif-data-hx-trigger="[[ .AutoRefresh ]]"
     data-hx-get="/status"
     data-hx-swap="innerHTML">
    Status content
</div>
```

When `AutoRefresh` is false, the `data-hx-trigger` attribute is omitted, disabling automatic refresh.

## Patterns

### Loading States

Use `data-g-ignore` for preview content that HTMX replaces:

```html
<div data-g-define="detail-page">
    <div data-hx-get="[[ .DataUrl ]]"
         data-hx-trigger="load"
         data-hx-target="this"
         data-hx-swap="outerHTML">
        <p data-g-ignore="outer">Loading placeholder for browser preview...</p>
        <div class="spinner">Loading...</div>
    </div>
</div>
```

### Out-of-Band Updates

HTMX out-of-band swaps work naturally — just render multiple templates:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Main content
    engine.Render(r.Context(), w, "item-list", listData)
    // OOB update for a counter elsewhere on the page
    engine.Render(r.Context(), w, "item-count", countData)
}
```

```html
<span data-g-define="item-count"
      id="item-count"
      data-hx-swap-oob="true"
      data-g-inner-text="[[ .Count ]]">0</span>
```

### Form Handling

```html
<form data-g-define="edit-form"
      data-hx-post="[[ .SubmitUrl ]]"
      data-hx-target="#content"
      data-hx-swap="innerHTML">
    <input type="text" name="name"
           data-g-att-value="[[ .Name ]]"
           data-g-attif-class="[[ .NameInvalid ]]"
           data-g-att-class="error" />
    <button type="submit">Save</button>
</form>
```

## Browser Preview

Both gotmx and HTMX use `data-*` attributes, so templates remain valid HTML. Preview content in `data-g-define-slot` and `data-g-ignore` elements gives designers a realistic preview without running any server.
