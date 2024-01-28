# Designing Templates for Browser Preview

One of Gotmx's key benefits is that templates are valid HTML. You can open any template file directly in a browser and see how it looks — with realistic sample content, working styles, and functional navigation. This guide explains all the techniques for making templates that look great in preview while rendering correctly at runtime.

## Why This Matters

Traditional template engines use syntax like `{{ .Name }}` or `<%= name %>` that breaks HTML validation, confuses editors, and renders as garbled text in the browser. Gotmx uses standard `data-*` attributes, which browsers simply ignore. This means:

- **Designers** can work on templates in the browser without running any server
- **Developers** can see the structure and layout before writing any Go code
- **Reviews** are easier — open the template file and see what it will look like
- **Style iteration** is instant — edit HTML/CSS, refresh browser, no build step

## Technique 1: Placeholder Text

Every element with `data-g-inner-text` contains placeholder text that browsers display but gotmx replaces:

```html
<h1 data-g-inner-text="[[ .Title ]]">Welcome to Our Application</h1>
<p data-g-inner-text="[[ .Description ]]">
    This is a sample description that helps designers see
    how the layout looks with realistic content.
</p>
<span data-g-inner-text="[[ .Count ]]">42</span> items found
```

**In browser preview:** Shows "Welcome to Our Application", the full description, and "42 items found".

**At runtime:** The placeholder text is replaced by the actual model values.

**Tip:** Use realistic sample data in your placeholders, not "lorem ipsum". This helps designers understand the actual content and plan for edge cases like long names or large numbers.

## Technique 2: Preview-Only Elements with g-ignore

Add extra elements that make the preview look complete, but are excluded from rendering:

```html
<ul data-g-define="task-list">
    <li data-g-outer-repeat="[[ .Tasks ]]">
        <span data-g-inner-text="[[ .Title ]]">Configure database connection</span>
    </li>
    <!-- These items only appear in browser preview -->
    <li data-g-ignore="outer">Set up authentication provider</li>
    <li data-g-ignore="outer">Deploy to staging environment</li>
    <li data-g-ignore="outer">Run integration tests</li>
</ul>
```

**In browser preview:** Shows a list with 4 items — the first is the template, the other three are sample data.

**At runtime:** Only the first `<li>` is used as the repeat template. The ignored items are stripped. The actual list comes from the `.Tasks` data.

This is especially useful for:
- Lists and tables — one item is the template, extras are preview padding
- Navigation menus — show realistic menu structure
- Grids and card layouts — fill out the visual layout

## Technique 3: Dual Attributes for Paths

Use static attributes for browser preview and `data-g-*` attributes for runtime:

```html
<!-- Static src works when previewing the file from its directory -->
<!-- data-g-src provides the correct server path at runtime -->
<script type="module" src="../static/js/main.js" data-g-src="/static/js/main.js"></script>
<link rel="stylesheet" href="../static/css/style.css" data-g-href="/static/css/style.css" />
<link rel="icon" href="../static/favicon.ico" data-g-href="/static/favicon.ico" />
<img src="../static/images/logo.png" data-g-src="/static/images/logo.png" alt="Logo" />
```

**In browser preview:** The relative paths (`../static/...`) resolve correctly when opening the HTML file from its location in the project.

**At runtime:** The `data-g-src` and `data-g-href` attributes override the static values with absolute server paths.

This pattern is critical for:
- CSS and JavaScript includes
- Images and favicons
- Links to other pages

## Technique 4: Slot Default Content as Preview

Content inside `data-g-define-slot` elements is only for browser preview. At runtime, it is replaced by whatever the caller injects:

```html
<div data-g-define="page-layout">
    <header>
        <nav>Navigation Bar</nav>
    </header>
    <main data-g-define-slot="content">
        <!-- This entire block is preview-only. At runtime, the actual
             page content is injected here by the caller. -->
        <h1>Sample Page Title</h1>
        <p>This is placeholder content that shows the layout structure
           when previewing the template in a browser. It demonstrates
           how the main content area will look with typical content.</p>
        <div class="card">
            <h2>Sample Card</h2>
            <p>Card content goes here.</p>
        </div>
    </main>
    <footer>Footer Content</footer>
</div>
```

**In browser preview:** Shows a complete page with header, sample content, a card, and footer.

**At runtime:** The `<main>` element's content is replaced entirely by whatever the caller puts in the "content" slot.

## Technique 5: Full HTML Scaffolding Around Templates

Template files can include a complete HTML page structure for preview, even though only the `data-g-define` elements are used at runtime:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Instance List - Preview</title>
    <!-- Include styles so the preview looks correct -->
    <link rel="stylesheet" href="../static/css/style.css" />
</head>
<body>
    <!-- NOTE: This surrounding HTML is not used at runtime. Only the element
         with data-g-define is extracted as a template. The rest is scaffolding
         so that opening this file in a browser shows a realistic preview. -->

    <div data-g-define="instance-list">
        <h2>Instances</h2>
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody data-g-inner-repeat="[[ .Instances ]]">
                <tr>
                    <td data-g-inner-text="[[ .Name ]]">my-app-prod</td>
                    <td data-g-inner-text="[[ .Status ]]">Running</td>
                </tr>
            </tbody>
        </table>
    </div>
</body>
</html>
```

**In browser preview:** Opens as a complete web page with styles, a table header, and one sample row.

**At runtime:** Only the `<div data-g-define="instance-list">` element is registered as a template. Everything outside it (the `<html>`, `<head>`, `<body>`) is ignored.

## Technique 6: Conditional Variants as Preview

When an element shows different states based on conditions, include all variants in the template. In the browser all of them are visible, at runtime only the matching one renders:

```html
<td>
    <span data-g-if="[[ ?.Status=='running' ]]" class="badge badge-green"
          data-g-inner-text="[[ .Status ]]">running</span>
    <span data-g-if="[[ ?.Status=='stopped' ]]" class="badge badge-red"
          data-g-inner-text="[[ .Status ]]">stopped</span>
    <span data-g-if="[[ ?.Status=='pending' ]]" class="badge badge-yellow"
          data-g-inner-text="[[ .Status ]]">pending</span>
</td>
```

**In browser preview:** Shows all three badges stacked, letting the designer see every variant's styling.

**At runtime:** Only the matching badge renders based on the actual `.Status` value.

## Technique 7: Sample Script Data

For templates that include dynamic JavaScript data, put realistic sample data in the `<script>` tag and override it at runtime:

```html
<canvas id="chart"></canvas>
<script data-g-inner-html="[[ .ChartScript ]]">
    // Sample data for browser preview — replaced at runtime
    const chartData = [
        { month: "Jan", value: 120 },
        { month: "Feb", value: 250 },
        { month: "Mar", value: 180 },
    ];
    renderChart(document.getElementById('chart'), chartData);
</script>
```

**In browser preview:** The sample JavaScript runs and renders a chart with sample data.

**At runtime:** `data-g-inner-html` replaces the entire script content with dynamically generated JavaScript.

## Technique 8: Dynamic Attributes with Static Fallbacks

Use regular HTML attributes for preview and `data-g-att-*` to override them at runtime:

```html
<a href="#" data-g-att-href="[[ .ProfileUrl ]]"
   class="link" data-g-att-class="[[ .LinkClass ]]">
    <span data-g-inner-text="[[ .UserName ]]">John Doe</span>
</a>

<button type="button" data-g-attif-disabled="[[ .IsLoading ]]">
    <span data-g-inner-text="[[ .ButtonLabel ]]">Submit</span>
</button>
```

**In browser preview:** The link points to `#` (clickable but harmless), the button is enabled, and sample text is shown.

**At runtime:** The link points to the actual profile URL, the class is dynamic, and the button disables when loading.

## Putting It All Together

A well-designed gotmx template file looks like a complete, working HTML page:

```html
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="../static/css/app.css" />
</head>
<body>
    <!-- The actual template starts here -->
    <div data-g-define="user-dashboard" class="dashboard">
        <h1 data-g-inner-text="[[ .Greeting ]]">Good morning, Alice</h1>

        <section class="stats">
            <div class="stat-card">
                <span class="stat-value" data-g-inner-text="[[ .TaskCount ]]">7</span>
                <span class="stat-label">Open Tasks</span>
            </div>
            <div class="stat-card">
                <span class="stat-value" data-g-inner-text="[[ .MessageCount ]]">3</span>
                <span class="stat-label">New Messages</span>
            </div>
        </section>

        <h2>Recent Activity</h2>
        <ul class="activity-feed">
            <li data-g-outer-repeat="[[ .Activities ]]">
                <strong data-g-inner-text="[[ .User ]]">Bob</strong>
                <span data-g-inner-text="[[ .Action ]]">deployed version 2.4.1</span>
                <time data-g-inner-text="[[ .When ]]">2 hours ago</time>
            </li>
            <li data-g-ignore="outer">
                <strong>Carol</strong>
                <span>merged pull request #42</span>
                <time>5 hours ago</time>
            </li>
            <li data-g-ignore="outer">
                <strong>Dave</strong>
                <span>created issue "Fix login timeout"</span>
                <time>yesterday</time>
            </li>
        </ul>
    </div>
</body>
</html>
```

Open this file in a browser and you see a complete dashboard with stats, a greeting, and three activity items. At runtime, the data comes from Go, the ignored items disappear, and the repeat produces as many items as the data contains.

## Best Practices

1. **Use realistic sample data** — "John Doe" and "42" are more useful than "placeholder" and "0"
2. **Match the expected length** — if names will be long, use long sample names to test layout
3. **Include multiple variants** — show all conditional states so designers can style each one
4. **Comment the intent** — add HTML comments explaining what is preview-only vs runtime
5. **Keep static paths relative** — use `../static/` paths that resolve when opening the file from its directory
6. **Include enough ignored items** — a list with 3-4 sample items shows how the layout handles multiple entries
