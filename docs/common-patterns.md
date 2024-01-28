# Common Gotmx Patterns

Recipes for typical template tasks.

## Dynamic CSS Classes

Use `data-g-class` (shortcut for `data-g-att-class`) to set classes dynamically:

```html
<div data-g-define="alert" data-g-class="[[ .AlertClass ]]">
    <span data-g-inner-text="[[ .Message ]]">Alert text</span>
</div>
```

Conditionally add a class with `data-g-attif-class`:

```html
<tr data-g-outer-repeat="[[ .Rows ]]"
    class="row"
    data-g-attif-class="[[ .IsHighlighted ]]">
    <td data-g-inner-text="[[ .Name ]]">Name</td>
</tr>
```

## Dynamic Attributes

Set any attribute with `data-g-att-*`:

```html
<a data-g-att-href="[[ .Url ]]" data-g-inner-text="[[ .Label ]]">Link</a>
<img data-g-att-src="[[ .ImageUrl ]]" data-g-att-alt="[[ .AltText ]]" />
<div data-g-att-data-id="[[ .Id ]]" data-g-att-data-type="[[ .Type ]]">Content</div>
```

## Boolean Attributes

Boolean HTML attributes like `disabled`, `checked`, `hidden`, and `required` render without a value when present:

```html
<button data-g-attif-disabled="[[ .IsLoading ]]">Submit</button>
<!-- When IsLoading=true:  <button disabled>Submit</button> -->
<!-- When IsLoading=false: <button>Submit</button> -->

<input type="checkbox" data-g-attif-checked="[[ .IsSelected ]]" />
<!-- When IsSelected=true: <input type="checkbox" checked /> -->
```

## Conditional Rendering

Show or hide elements:

```html
<div data-g-if="[[ .IsAdmin ]]">
    <h2>Admin Panel</h2>
</div>
```

Toggle between alternatives:

```html
<span data-g-if="[[ .IsOnline ]]" class="text-green">Online</span>
<span data-g-if="[[ .IsOffline ]]" class="text-red">Offline</span>
```

## Iteration

### Repeat the element (most common)

```html
<ul>
    <li data-g-outer-repeat="[[ .Items ]]" data-g-inner-text="[[ .Name ]]">
        Sample Item
    </li>
</ul>
```

### Repeat children only (keep wrapper)

```html
<tbody data-g-inner-repeat="[[ .Rows ]]">
    <tr>
        <td data-g-inner-text="[[ .Name ]]">Name</td>
        <td data-g-inner-text="[[ .Email ]]">email@example.com</td>
    </tr>
</tbody>
```

### Iterate maps

When iterating over a map, each item has `.Key` and `.Value` properties:

```html
<dl data-g-inner-repeat="[[ .Settings ]]">
    <dt data-g-inner-text="[[ .Key ]]">Setting Name</dt>
    <dd data-g-inner-text="[[ .Value ]]">Setting Value</dd>
</dl>
```

## Nested Components

Define a card, then use it from another template:

```html
<!-- card.htm -->
<div data-g-define="card" class="card">
    <div class="card-header" data-g-define-slot="header">Default</div>
    <div class="card-body" data-g-define-slot="">Body</div>
</div>

<!-- dashboard.htm -->
<div data-g-define="dashboard">
    <div data-g-use="card">
        <h2 data-g-use-slot="header" data-g-inner-text="[[ .Title ]]">Title</h2>
        <p data-g-inner-text="[[ .Summary ]]">Summary text</p>
    </div>
</div>
```

## Layout Composition

### Via render option (recommended)

Wrap any page in a layout without modifying the template:

```go
engine.Render(r.Context(), w, "dashboard-page", pageData,
    gotmx.WithLayout("main-layout", layoutData),
)
```

### Via g-use in templates

```html
<!-- layout.htm -->
<html data-g-define="layout">
<head><title data-g-inner-text="[[ .Title ]]">Title</title></head>
<body>
    <nav>Navigation</nav>
    <main data-g-define-slot="content">Default</main>
</body>
</html>

<!-- page.htm -->
<div data-g-define="page">
    <div data-g-use="layout">
        <article data-g-use-slot="content">
            <h1>Page Content</h1>
        </article>
    </div>
</div>
```

## Surrogate Content for Browser Preview

Use `data-g-ignore` to include content visible in browser preview but excluded from rendering:

```html
<ul data-g-define="nav-menu">
    <li data-g-outer-repeat="[[ .Items ]]" data-g-inner-text="[[ .Label ]]">
        Real Item
    </li>
    <!-- These only show in browser preview, not at runtime -->
    <li data-g-ignore="outer">Preview Item 2</li>
    <li data-g-ignore="outer">Preview Item 3</li>
    <li data-g-ignore="outer">Preview Item 4</li>
</ul>
```

## Template Constants

Define constants for template references to get IDE autocomplete and compile-time safety:

```go
package dashboard

const (
    TemplateDashboardPage = "modules/dashboard/dashboard.htm#dashboard-page"
    TemplateWidgetCard    = "modules/dashboard/dashboard.htm#widget-card"
)
```

Use them in handlers:

```go
engine.Render(r.Context(), w, dashboard.TemplateDashboardPage, data)
```

## Context Switching

Use `data-g-with` to avoid repeating long paths:

```html
<!-- Without g-with -->
<div>
    <span data-g-inner-text="[[ .User.Profile.FirstName ]]">First</span>
    <span data-g-inner-text="[[ .User.Profile.LastName ]]">Last</span>
    <span data-g-inner-text="[[ .User.Profile.Email ]]">Email</span>
</div>

<!-- With g-with -->
<div data-g-with="[[ .User.Profile ]]">
    <span data-g-inner-text="[[ .FirstName ]]">First</span>
    <span data-g-inner-text="[[ .LastName ]]">Last</span>
    <span data-g-inner-text="[[ .Email ]]">Email</span>
</div>
```

## Escaping Rules

| What you want | Use this |
|---------------|----------|
| Safe text (user input) | `data-g-inner-text` (always escapes) |
| Trusted HTML content | `data-g-inner-html` (never escapes) |
| Replace element with text | `data-g-outer-text` (escapes by default) |
| Dynamic attribute value | `data-g-att-*` (always escapes) |

`data-g-inner-text` is unconditionally safe — it escapes even when the `Unescaped()` render option is used. Use it for any user-supplied content.
