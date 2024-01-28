# Gotmx Template Attributes

The Gotmx HTML template engine uses plain HTML attributes to control template rendering. All attributes have two versions:

- A **long version** that starts with `data-` (e.g., `data-g-if`)
- A **short version** without the `data-` prefix (e.g., `g-if`)

If your editor complains about unknown attributes, use the `data-` version. Since `data-` attributes are part of the HTML5 standard, editors should not warn about them.

When both versions are present on the same element, the long version takes priority, except for [`g-define`](#g-define) where the first attribute encountered is used.

## Attribute Categories

Gotmx attributes are organized into the following categories:

| Category | Attributes | Purpose |
|----------|------------|---------|
| **Template Definition** | [g-define](#g-define) | Define named templates |
| **Control Flow** | [g-if](#g-if), [g-ignore](#g-ignore), [g-with](#g-with) | Control rendering flow and data context |
| **Iteration** | [g-outer-repeat](#g-outer-repeat), [g-inner-repeat](#g-inner-repeat) | Repeat elements or content |
| **Content** | [g-inner-text](#g-inner-text), [g-inner-html](#g-inner-html), [g-outer-text](#g-outer-text), [g-define-slot](#g-define-slot) | Control element content |
| **Composition** | [g-use](#g-use), [g-inner-use](#g-inner-use), [g-use-slot](#g-use-slot), [g-override-att](#g-override-att), [g-as-template](#g-as-template), [g-as-unsafe-template](#g-as-unsafe-template) | Compose and reuse templates |
| **Transformation** | [g-trans](#g-trans), [g-att-?](#g-att), [g-attif-?](#g-attif), [g-class](#g-class), [g-href](#g-href), [g-src](#g-src) | Transform elements and attributes |

## Expression Types

Gotmx attributes support three types of values:

### Literal Values
Plain text values used as-is:
```html
<div g-class="active">Content</div>
<div g-if="true">Always renders</div>
```

### Model Path Expressions
Access data model properties using `[[ .Path ]]` syntax:
```html
<div g-if="[[ .IsVisible ]]">Conditional content</div>
<div g-inner-text="[[ .User.Name ]]">Name placeholder</div>
<div g-outer-repeat="[[ .Items ]]">Repeated content</div>
```

Model path features:
- Property access: `[[ .User.Name ]]`
- Array indexing: `[[ .Items[0] ]]`
- Map access: `[[ .Settings.Theme ]]`
- Current data context: `[[ . ]]`

The default implementation uses the [empaths](https://github.com/maddiesch/empaths) library for model 
path resolution. See the empaths documentation for additional path syntax features and capabilities. 
Custom model path resolvers can be provided when configuring Gotmx to change how model path expressions look any way you want.

### Go Template Expressions
Native Go template syntax using `{{ }}`:
```html
<div g-as-template>
    Hello, {{ .Name }}!
    {{ if .IsAdmin }}You are an admin.{{ end }}
</div>
```

Go template expressions require [`g-as-template`](#g-as-template) or [`g-as-unsafe-template`](#g-as-unsafe-template) on the element or an ancestor.

## Attribute Processing Order

Attributes are processed in this order during rendering:

1. **g-ignore** - Early exit if element should be skipped
2. **g-with** - Switch data context
3. **g-if** - Evaluate condition
4. **g-attif-\*** - Process conditional attributes
5. **g-outer-repeat** - Repeat element (if present, handles remaining phases per iteration)
6. **Element rendering** - Process remaining attributes and render

## All Attributes Reference

| Long Version | Short Version | Description |
|--------------|---------------|-------------|
| [data-g-as-template](#g-as-template) | [g-as-template](#g-as-template) | Treats the innerHTML as a Go HTML template |
| [data-g-as-unsafe-template](#g-as-unsafe-template) | [g-as-unsafe-template](#g-as-unsafe-template) | Treats the innerHTML as a Go text template (no escaping) |
| [data-g-att-?](#g-att) | [g-att-?](#g-att) | Dynamically sets an attribute |
| [data-g-attif-?](#g-attif) | [g-attif-?](#g-attif) | Conditionally adds or removes an attribute |
| [data-g-class](#g-class) | [g-class](#g-class) | Sets the class attribute (shortcut for g-att-class) |
| [data-g-define](#g-define) | [g-define](#g-define) | Defines a named template |
| [data-g-define-slot](#g-define-slot) | [g-define-slot](#g-define-slot) | Defines a slot where content can be injected |
| [data-g-href](#g-href) | [g-href](#g-href) | Sets the href attribute (shortcut for g-att-href) |
| [data-g-if](#g-if) | [g-if](#g-if) | Conditionally renders the element |
| [data-g-ignore](#g-ignore) | [g-ignore](#g-ignore) | Controls whether element/children are rendered |
| [data-g-inner-html](#g-inner-html) | [g-inner-html](#g-inner-html) | Replaces innerHTML with unescaped HTML |
| [data-g-inner-repeat](#g-inner-repeat) | [g-inner-repeat](#g-inner-repeat) | Repeats the innerHTML for each item in a collection |
| [data-g-inner-text](#g-inner-text) | [g-inner-text](#g-inner-text) | Replaces innerHTML with escaped text |
| [data-g-inner-use](#g-inner-use) | [g-inner-use](#g-inner-use) | Renders a template's content without its outer element |
| [data-g-outer-repeat](#g-outer-repeat) | [g-outer-repeat](#g-outer-repeat) | Repeats the element for each item in a collection |
| [data-g-outer-text](#g-outer-text) | [g-outer-text](#g-outer-text) | Replaces the entire element with escaped text |
| [data-g-override-att](#g-override-att) | [g-override-att](#g-override-att) | Specifies which attributes to pass to a component via g-use |
| [data-g-src](#g-src) | [g-src](#g-src) | Sets the src attribute (shortcut for g-att-src) |
| [data-g-trans](#g-trans) | [g-trans](#g-trans) | Transforms the element tag name |
| [data-g-use](#g-use) | [g-use](#g-use) | Renders a different template instead of this element |
| [data-g-use-slot](#g-use-slot) | [g-use-slot](#g-use-slot) | Specifies which slot to place content in |
| [data-g-with](#g-with) | [g-with](#g-with) | Changes the data context for this element and children |

---

## Template Definition

### g-define

Defines a named template that can be referenced and reused.

#### When to Use

- When creating reusable component templates
- At the root of template files
- On the `<html>` element when the entire file should be a single template

#### Values

- **String**: The template name (e.g., `"my-button"`)
- **Empty string**: Falls back to the element's `id` attribute

#### Interactions

- When both `g-define` and `data-g-define` are present, the **first one encountered** is used (unlike other attributes where the long form takes priority)
- [`g-if`](#g-if) and [`g-ignore`](#g-ignore) on the same element are ignored when the template is rendered directly as the root

#### Examples

```html
<!-- Define a simple template -->
<div data-g-define="greeting">
    Hello, World!
</div>

<!-- Define a template using the id attribute as fallback -->
<div data-g-define="" id="my-component">
    Component content
</div>

<!-- Define a full-page template on the html element -->
<!DOCTYPE html>
<html lang="en" data-g-define="page-layout">
<head>...</head>
<body>...</body>
</html>
```

---

## Control Flow

### g-if

Conditionally renders an element based on a boolean expression.

#### When to Use

- When you need to show or hide elements based on data
- For authentication or authorization checks
- For feature flags or conditional UI

#### Values

- **Literal**: `"true"` or `"false"` (case-sensitive)
- **Model path**: `[[ .IsVisible ]]`
- **Go template**: `{{ .IsVisible }}`

The expression must evaluate to `"true"` (case-insensitive) or `"!false"` to render. Any other value is treated as false.

#### Interactions

- Evaluated **after** [`g-with`](#g-with) - so conditions can use the new context
- **Ignored** on root template elements when rendered directly
- **Ignored** when [`g-outer-repeat`](#g-outer-repeat) is present on the same element

#### Examples

```html
<!-- Literal boolean values -->
<div data-g-if="true">Always renders</div>
<div data-g-if="false">Never renders</div>

<!-- Using model path -->
<div data-g-if="[[ .IsLoggedIn ]]">
    Welcome back!
</div>

<!-- Using Go template expression -->
<div data-g-as-template data-g-if="{{ .User.IsAdmin }}">
    Admin controls
</div>

<!-- Combined with g-with -->
<div data-g-with="[[ .User ]]" data-g-if="[[ .IsActive ]]">
    <!-- g-with is evaluated first, then g-if uses the new context -->
    Active user: <span data-g-inner-text="[[ .Name ]]"></span>
</div>
```

---

### g-ignore

Controls whether an element and/or its children are rendered.

#### When to Use

- Including sample content or documentation that should not be rendered
- Using wrapper elements for structure without rendering them
- Dynamically hiding content

#### Values

- **`"outer"`** or **empty string**: Skip the element and all children
- **`"inner"`**: Render element tags but skip children
- **`"outer-only"`**: Skip element tags but render children
- **`"none"`**: Override inherited ignore (render normally)
- **Model path or Go template**: Dynamic ignore behavior

#### Interactions

- **Highest priority** - checked before all other attributes
- [`g-inner-text`](#g-inner-text) and [`g-outer-text`](#g-outer-text) respect this attribute
- **Overridden to `"none"`** when template is rendered directly as root

#### Examples

```html
<!-- Skip element and children entirely -->
<p data-g-ignore="outer">
    This documentation text will not be rendered.
</p>

<!-- Render element but skip children -->
<div data-g-ignore="inner">
    <p>These children will not render</p>
</div>
<!-- Result: <div></div> -->

<!-- Skip element wrapper but render children -->
<div data-g-ignore="outer-only" data-g-as-template>
    This text renders without the div wrapper: {{ .DynamicValue }}
</div>
<!-- Result: This text renders without the div wrapper: [value] -->

<!-- Include sample content in preview -->
<ul data-g-define="nav-menu">
    <li data-g-outer-repeat="[[ .MenuItems ]]" data-g-inner-text="[[ .Label ]]">
        Sample Item
    </li>
    <li data-g-ignore="outer">Sample Item 2 (preview only)</li>
    <li data-g-ignore="outer">Sample Item 3 (preview only)</li>
</ul>
```

---

### g-with

Changes the data context for an element and all its children.

#### When to Use

- Reducing verbosity when accessing nested object properties
- Scoping data access to a specific object
- Working with iteration items

#### Values

- **Model path only**: `[[ .NestedObject ]]`
- Cannot be a literal value or Go template expression

If the specified path does not exist or is nil, the original context is preserved.

#### Interactions

- Evaluated **before** [`g-if`](#g-if) - so conditions can use the new context
- All descendant elements inherit the new context

#### Examples

```html
<!-- Access nested object properties more easily -->
<div data-g-with="[[ .User ]]">
    <h2 data-g-inner-text="[[ .Name ]]">Username</h2>
    <p data-g-inner-text="[[ .Email ]]">user@example.com</p>
    <span data-g-inner-text="[[ .Role ]]">Role</span>
</div>
<!-- Instead of: .User.Name, .User.Email, .User.Role -->

<!-- Context stays unchanged if path doesn't exist -->
<div data-g-with="[[ .NonExistent ]]">
    <!-- Original context is still available -->
    <span data-g-inner-text="[[ .Name ]]">Fallback</span>
</div>

<!-- Combine with repetition -->
<div data-g-outer-repeat="[[ .Products ]]" data-g-with="[[ . ]]">
    <h3 data-g-inner-text="[[ .Title ]]">Product</h3>
    <p data-g-inner-text="[[ .Description ]]">Description</p>
</div>
```

---

## Iteration

### g-outer-repeat

Repeats the entire element for each item in a collection.

#### When to Use

- Rendering lists where each item needs the same element structure
- Creating repeated components
- Building dynamic tables, lists, or grids

#### Values

- **Model path only**: `[[ .Items ]]`
- The path must point to an array, slice, or map
- For non-iterable values, the element renders once with the value as context

#### Interactions

- Cannot be used with [`g-inner-repeat`](#g-inner-repeat) on the same element
- [`g-ignore`](#g-ignore), [`g-with`](#g-with), [`g-if`](#g-if), [`g-attif-*`](#g-attif) are evaluated for each iteration
- Each iteration has the current item as its data context

#### Map Iteration

When iterating over maps, each item is a `MapEntry` with `.Key` and `.Value` properties.

#### Examples

```html
<!-- Repeat list items -->
<ul data-g-define="todo-list">
    <li data-g-outer-repeat="[[ .Tasks ]]" data-g-inner-text="[[ .Title ]]">
        Sample task
    </li>
</ul>
<!-- Result: <ul><li>Task 1</li><li>Task 2</li>...</ul> -->

<!-- Repeat with nested structure -->
<div data-g-outer-repeat="[[ .Users ]]" class="user-card">
    <h3 data-g-inner-text="[[ .Name ]]">User Name</h3>
    <p data-g-inner-text="[[ .Email ]]">user@example.com</p>
</div>

<!-- Iterate over a map -->
<div data-g-outer-repeat="[[ .Settings ]]">
    <span data-g-inner-text="[[ .Key ]]">Key</span>:
    <span data-g-inner-text="[[ .Value ]]">Value</span>
</div>

<!-- Conditional rendering per item -->
<div data-g-outer-repeat="[[ .Items ]]" data-g-if="[[ .IsActive ]]">
    <!-- Only renders for items where .IsActive is true -->
    <span data-g-inner-text="[[ .Name ]]">Item</span>
</div>
```

---

### g-inner-repeat

Repeats the innerHTML for each item in a collection, keeping the element wrapper.

#### When to Use

- When you need a single container element with repeated content
- Table rows within a table body
- Options within a select element

#### Values

- **Model path only**: `[[ .Items ]]`
- The path must point to an array, slice, or map

#### Interactions

- Cannot be used with [`g-outer-repeat`](#g-outer-repeat) on the same element
- Element tags render once, content repeats for each item
- Each iteration has the current item as its data context

#### Examples

```html
<!-- Single container with repeated content -->
<select data-g-inner-repeat="[[ .Options ]]">
    <option data-g-att-value="[[ .Value ]]" data-g-inner-text="[[ .Label ]]">
        Sample Option
    </option>
</select>
<!-- Result: <select><option>...</option><option>...</option>...</select> -->

<!-- Table with repeated rows -->
<tbody data-g-inner-repeat="[[ .Rows ]]">
    <tr>
        <td data-g-inner-text="[[ .Name ]]">Name</td>
        <td data-g-inner-text="[[ .Value ]]">Value</td>
    </tr>
</tbody>
```

---

## Content

### g-inner-text

Replaces the innerHTML with escaped text content.

#### When to Use

- Displaying user-generated content safely
- Dynamic text values that should never contain HTML
- Safe default for any dynamic text

#### Values

- **Literal**: `"Static text"`
- **Model path**: `[[ .TextValue ]]`
- **Go template**: `{{ .TextValue }}`

Content is always HTML-escaped, making it safe for untrusted data.

#### Interactions

- Respects [`g-ignore`](#g-ignore):
  - With `"outer"`: nothing rendered
  - With `"inner"`: the inner-text is skipped (it IS the inner content)
  - With `"outer-only"`: only text rendered, no element wrapper
- Works with [`g-use`](#g-use) - text is placed in the default slot
- Takes precedence over child elements (children are replaced)

#### Examples

```html
<!-- Static text -->
<div data-g-inner-text="Hello, World!">
    This placeholder text is replaced
</div>
<!-- Result: <div>Hello, World!</div> -->

<!-- Dynamic text with escaping -->
<div data-g-inner-text="[[ .UserInput ]]">Placeholder</div>
<!-- If .UserInput is "<script>alert('xss')</script>" -->
<!-- Result: <div>&lt;script&gt;alert('xss')&lt;/script&gt;</div> -->

<!-- Combined with g-ignore="outer-only" -->
<span data-g-ignore="outer-only" data-g-inner-text="[[ .Value ]]">Placeholder</span>
<!-- Result: [value] (no span wrapper) -->
```

---

### g-inner-html

Replaces the innerHTML with unescaped HTML content.

#### When to Use

- Rendering trusted HTML content (e.g., from a CMS)
- Markdown-to-HTML conversion results
- Template-generated HTML fragments

**Security Warning**: Never use with untrusted content. This can lead to XSS vulnerabilities.

#### Values

- **Literal**: `"<strong>Bold</strong>"`
- **Model path**: `[[ .HtmlContent ]]`
- **Go template**: `{{ .HtmlContent }}`

#### Interactions

- Checked **before** [`g-inner-text`](#g-inner-text) - if both present, `g-inner-html` wins
- Same [`g-ignore`](#g-ignore) rules as `g-inner-text`

#### Examples

```html
<!-- Render HTML content -->
<div data-g-inner-html="[[ .RichContent ]]">
    <p>Placeholder content</p>
</div>
<!-- If .RichContent is "<strong>Bold</strong> and <em>italic</em>" -->
<!-- Result: <div><strong>Bold</strong> and <em>italic</em></div> -->

<!-- Static HTML -->
<div data-g-inner-html="<span class='badge'>New</span>">Placeholder</div>
<!-- Result: <div><span class='badge'>New</span></div> -->
```

---

### g-outer-text

Replaces the entire element (including tags) with escaped text.

#### When to Use

- When you need a wrapper element for attributes but not in output
- Simple text output without element structure

#### Values

- **Literal**: `"Static text"`
- **Model path**: `[[ .TextValue ]]`
- **Go template**: `{{ .TextValue }}`

Content is HTML-escaped by default. When the `Unescaped()` render option is used, `g-outer-text` follows the global escaping setting (unlike `g-inner-text` which always escapes).

#### Interactions

- If [`g-ignore`](#g-ignore) is present (any value), nothing is rendered
- Respects [`g-if`](#g-if) (evaluated before)
- Works with [`g-outer-repeat`](#g-outer-repeat)
- Does **not** work with [`g-use`](#g-use)

#### Examples

```html
<!-- Replace element with text -->
<span data-g-outer-text="[[ .Name ]]">Placeholder</span>
<!-- Result: [name value] (no span element) -->

<!-- With iteration -->
<span data-g-outer-repeat="[[ .Tags ]]" data-g-outer-text="[[ . ]]">tag</span>
<!-- Result: tag1 tag2 tag3 -->
```

---

### g-define-slot

Defines a named slot where content can be injected when using the template.

#### When to Use

- Building component composition with named slots
- Creating reusable layouts with customizable sections
- Providing placeholder locations for content injection

#### Values

- **String**: Slot name (e.g., `"header"`, `"footer"`)
- **Empty string**: Default slot for unspecified content
- **Model path or Go template**: Dynamic slot name

#### Interactions

- Children of the slot element are ignored (used as sample/preview content)
- Multiple slots with the same name render the slotted content multiple times
- Use with [`g-use-slot`](#g-use-slot) on parent's children to fill slots

#### Examples

```html
<!-- Define a card component with named slots -->
<div data-g-define="card" class="card">
    <div class="card-header" data-g-define-slot="header">
        <h3>Sample Header</h3>
    </div>
    <div class="card-body" data-g-define-slot>
        <!-- Default slot (empty name) -->
        <p>Sample body content</p>
    </div>
    <div class="card-footer" data-g-define-slot="footer">
        <button>Sample Action</button>
    </div>
</div>

<!-- Use the card component with slot content -->
<div data-g-use="card">
    <h3 data-g-use-slot="header">Custom Header</h3>
    <p>This goes to the default slot</p>
    <button data-g-use-slot="footer">Custom Action</button>
</div>
```

---

## Composition

### g-use

Renders a different template instead of the current element.

#### When to Use

- Including reusable components
- Template composition and reuse
- Dynamic component selection based on data

#### Values

- **Template name**: `"my-component"`
- **Model path**: `[[ .ComponentName ]]`
- **Go template**: `{{ .ComponentName }}`
- **Fully-qualified**: `"components/button.html#button"`

#### Interactions

- Supports [`g-override-att`](#g-override-att) for passing attributes
- Children are placed in the target template's default slot
- [`g-inner-text`](#g-inner-text) is placed in the default slot if present
- Cannot be used with [`g-outer-text`](#g-outer-text)
- Cannot be used with [`g-inner-use`](#g-inner-use) on the same element

#### Examples

```html
<!-- Define a button component -->
<button data-g-define="primary-button" class="btn btn-primary">
    <span data-g-define-slot>Button Text</span>
</button>

<!-- Use the button component -->
<div data-g-use="primary-button">Click Me</div>
<!-- Result: <button class="btn btn-primary"><span>Click Me</span></button> -->

<!-- Dynamic component selection -->
<div data-g-use="[[ .ComponentType ]]">Content</div>

<!-- Pass children to slots -->
<div data-g-use="card">
    <h3 data-g-use-slot="header">Card Title</h3>
    <p>Card body content (goes to default slot)</p>
</div>
```

---

### g-inner-use

Renders a template's content without its outer element tags.

#### When to Use

- When you need to keep the current element but use another template's content
- Reusing content templates without their container

#### Values

Same as [`g-use`](#g-use): template name, model path, or Go template expression.

#### Interactions

- Current element tags are kept, target template's outer tags are removed
- Cannot be used with [`g-use`](#g-use) on the same element

#### Examples

```html
<!-- Define a template -->
<div data-g-define="content-template">
    <p>Template content</p>
    <p>More content</p>
</div>

<!-- Use just the content -->
<section data-g-inner-use="content-template">
    Original content (replaced)
</section>
<!-- Result: <section><p>Template content</p><p>More content</p></section> -->
```

---

### g-use-slot

Specifies which slot to place child content in when using a component.

#### When to Use

- Directing content to specific named slots in a component
- Organizing complex component composition
- Filling multiple slots in one component

#### Values

- **String**: Slot name
- **Empty string**: Default slot
- **Model path or Go template**: Dynamic slot name

Only meaningful on children of elements with [`g-use`](#g-use).

#### Interactions

- Element with `g-use-slot` is not rendered; only its content goes to the slot
- Multiple children can target the same slot (content is concatenated)
- Content without `g-use-slot` goes to the default slot

#### Examples

```html
<!-- Define a modal component -->
<div data-g-define="modal" class="modal">
    <div class="modal-header" data-g-define-slot="header">Header</div>
    <div class="modal-body" data-g-define-slot>Body</div>
    <div class="modal-footer" data-g-define-slot="footer">Footer</div>
</div>

<!-- Use with slot assignments -->
<div data-g-use="modal">
    <h2 data-g-use-slot="header">Confirm Action</h2>
    <p>Are you sure you want to proceed?</p>  <!-- Goes to default slot -->
    <div data-g-use-slot="footer">
        <button>Cancel</button>
        <button>Confirm</button>
    </div>
</div>
```

---

### g-override-att

Specifies which attributes to pass from the current element to a component via [`g-use`](#g-use).

#### When to Use

- Customizing component behavior through attributes
- Passing styling or data attributes to components
- Making components configurable from the outside

#### Values

- **Comma-separated list**: `"class,style,data-id"`
- **Model path**: `[[ .AttributeList ]]`
- **Go template**: `{{ .AttributeList }}`

#### Interactions

- Only used with [`g-use`](#g-use) or [`g-inner-use`](#g-inner-use)
- Matching [`g-attif-*`](#g-attif) attributes are also passed
- Overrides the component's original attribute values

#### Examples

```html
<!-- Define a button component -->
<button data-g-define="btn" class="btn" type="button">
    <span data-g-define-slot>Click</span>
</button>

<!-- Override the class attribute -->
<div data-g-use="btn" data-g-override-att="class" class="btn btn-primary btn-large">
    Submit
</div>
<!-- Result: <button class="btn btn-primary btn-large" type="button"><span>Submit</span></button> -->

<!-- Override multiple attributes -->
<div data-g-use="btn" data-g-override-att="class,type" class="btn-danger" type="submit">
    Delete
</div>

<!-- With conditional attribute -->
<div data-g-use="btn" data-g-override-att="class,disabled"
     class="btn-secondary" data-g-attif-disabled="[[ .IsLoading ]]">
    Save
</div>
```

---

### g-as-template

Treats the innerHTML as a Go HTML template with automatic escaping.

#### When to Use

- When you need Go template expressions like `{{ .Field }}`
- For dynamic content requiring Go template syntax
- When mixing Go templates with Gotmx directives

#### Values

- **Empty string or no value**: Auto-generates a template name
- **String**: Specific template name to use

Child elements inherit this setting unless they have their own `g-as-template` attribute.

#### Interactions

- Cannot be used with [`g-as-unsafe-template`](#g-as-unsafe-template) on the same element
- Attribute values are always escaped regardless of this setting

#### Examples

```html
<!-- Enable Go template processing -->
<div data-g-define="greeting" data-g-as-template>
    Hello, {{ .Name }}!
    {{ if .IsAdmin }}
        <span class="badge">Admin</span>
    {{ end }}
</div>

<!-- Use Go template functions -->
<div data-g-as-template>
    Items: {{ len .Items }}
    {{ range .Items }}
        <p>{{ . }}</p>
    {{ end }}
</div>

<!-- Escaping is automatic -->
<div data-g-as-template>
    User input: {{ .UserInput }}
    <!-- If UserInput is "<script>", renders as "&lt;script&gt;" -->
</div>
```

---

### g-as-unsafe-template

Treats the innerHTML as a Go text template without HTML escaping.

#### When to Use

- When you need to output raw HTML from Go templates
- Rendering trusted template content that intentionally contains HTML

**Security Warning**: Never use with untrusted content. This can lead to XSS vulnerabilities.

#### Values

Same as [`g-as-template`](#g-as-template).

#### Examples

```html
<!-- Render raw HTML (dangerous!) -->
<div data-g-define="raw-content" data-g-as-unsafe-template>
    {{ .TrustedHtmlContent }}
</div>
```

---

## Transformation

### g-trans

Transforms the element's tag name at render time.

#### When to Use

- Rendering different elements based on data
- HTMX out-of-band (OOB) updates where tag needs to change
- Dynamic semantic HTML based on context

#### Values

- **HTML tag name**: `"span"`, `"article"`, `"section"`
- **Model path**: `[[ .TagName ]]`
- **Go template**: `{{ .TagName }}`

#### Examples

```html
<!-- Static transformation -->
<div data-g-trans="article">Content</div>
<!-- Result: <article>Content</article> -->

<!-- Dynamic transformation -->
<div data-g-trans="[[ .HeadingLevel ]]" data-g-inner-text="[[ .Title ]]">
    Title
</div>
<!-- If .HeadingLevel is "h2", result: <h2>Title Text</h2> -->

<!-- HTMX OOB update example -->
<title data-g-trans="title" hx-swap-oob="true">Page Title</title>
```

---

### g-att

Dynamically sets HTML attributes at runtime.

The attribute name follows the pattern `g-att-{attribute-name}` (e.g., `g-att-disabled`, `g-att-data-id`).

#### When to Use

- When an attribute should be dynamic but not visible in HTML preview
- Setting form attributes dynamically: `disabled`, `readonly`, `checked`
- When the attribute's presence or value depends on data

#### Values

- **Empty string**: For boolean attributes like `disabled`
- **Literal value**: `"true"`, `"custom-value"`
- **Model path**: `[[ .Value ]]`
- **Go template**: `{{ .Value }}`

#### Examples

```html
<!-- Boolean attribute -->
<button data-g-att-disabled="">Click me!</button>
<!-- Result: <button disabled>Click me!</button> -->

<!-- Dynamic value -->
<input data-g-att-value="[[ .DefaultValue ]]" type="text">

<!-- Data attributes -->
<div data-g-att-data-user-id="[[ .User.ID ]]">Content</div>
<!-- Result: <div data-user-id="123">Content</div> -->

<!-- Preview-safe disabled button -->
<button data-g-att-disabled="[[ .IsSubmitting ]]">
    Submit <!-- Button works in preview, disabled at runtime if IsSubmitting -->
</button>
```

---

### g-attif

Conditionally adds or removes HTML attributes based on a boolean expression.

The attribute name follows the pattern `g-attif-{attribute-name}` (e.g., `g-attif-disabled`).

#### When to Use

- When an attribute should only be present under certain conditions
- Toggling boolean attributes like `disabled`, `readonly`, `checked`
- Conditional styling or data attributes

#### Values

- **Literal**: `"true"` or `"false"`
- **Model path**: `[[ .Condition ]]`
- **Go template**: `{{ .Condition }}`

#### Behavior

- If **true** and attribute doesn't exist: adds the attribute with empty value
- If **true** and attribute exists: leaves it unchanged
- If **false** and attribute exists: removes it
- If **false** and attribute doesn't exist: no effect

#### Examples

```html
<!-- Add disabled if condition is true -->
<button data-g-attif-disabled="[[ .IsLoading ]]">Submit</button>
<!-- If .IsLoading is true: <button disabled>Submit</button> -->
<!-- If .IsLoading is false: <button>Submit</button> -->

<!-- Remove existing attribute if condition is false -->
<input type="checkbox" checked data-g-attif-checked="[[ .IsSelected ]]">
<!-- If .IsSelected is false, the checked attribute is removed -->

<!-- Multiple conditional attributes -->
<input type="text"
       data-g-attif-disabled="[[ .IsDisabled ]]"
       data-g-attif-readonly="[[ .IsReadOnly ]]"
       data-g-attif-required="[[ .IsRequired ]]">
```

---

### g-class

Sets the class attribute dynamically. This is a shortcut for `g-att-class`.

#### When to Use

- Dynamically applying CSS classes based on data
- Cleaner syntax for the common case of dynamic classes

#### Values

Same as [`g-att-?`](#g-att).

#### Examples

```html
<div data-g-class="[[ .CssClass ]]">Content</div>
<!-- Equivalent to: <div data-g-att-class="[[ .CssClass ]]">Content</div> -->

<div data-g-class="card [[ .CardType ]]">Content</div>
```

---

### g-href

Sets the href attribute dynamically. This is a shortcut for `g-att-href`.

#### When to Use

- Dynamically building URLs based on data
- Links that depend on application state

#### Values

Same as [`g-att-?`](#g-att).

#### Examples

```html
<a data-g-href="[[ .ProfileUrl ]]">View Profile</a>
<!-- Equivalent to: <a data-g-att-href="[[ .ProfileUrl ]]">View Profile</a> -->

<a data-g-href="/users/[[ .UserId ]]">User Details</a>
```

---

### g-src

Sets the src attribute dynamically. This is a shortcut for `g-att-src`.

#### When to Use

- Dynamic image sources
- Dynamic script or iframe sources

#### Values

Same as [`g-att-?`](#g-att).

#### Examples

```html
<img data-g-src="[[ .AvatarUrl ]]" alt="Profile">
<!-- Equivalent to: <img data-g-att-src="[[ .AvatarUrl ]]" alt="Profile"> -->

<img data-g-src="/images/[[ .ImageId ]].png" alt="[[ .ImageAlt ]]">
```

---

## Escaping and Security

| Attribute | Escaping | Safe for User Input |
|-----------|----------|---------------------|
| `g-inner-text` | Always escaped (even with `Unescaped()`) | Yes |
| `g-outer-text` | Escaped by default, follows `Unescaped()` | Yes (by default) |
| `g-inner-html` | Never escaped | No |
| Attribute values | Always escaped | Yes |
| Text nodes | Escaped by default, follows `Unescaped()` | Yes (by default) |
| `g-as-template` | Go HTML template (escaped) | Depends on template |
| `g-as-unsafe-template` | Never escaped | No |

**Best Practices:**
- Always use `g-inner-text` for user-supplied content — it is unconditionally safe
- Never use `g-inner-html` or `g-as-unsafe-template` with untrusted data
- Attribute values are always escaped regardless of rendering mode
- The `Unescaped()` render option only affects text nodes and `g-outer-text`; it does **not** weaken `g-inner-text` or attribute escaping
