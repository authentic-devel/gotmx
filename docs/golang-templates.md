# Working with Golang templates in Gotmx

Gotmx allows you to use Golang HTML and text templates inside Gotmx templates in case you need anything that is not
supported by Gotmx template attributes.

All Golang templates that are registered or discovered by Gotmx are associated within a single "root" template named
"GoTmx". That means all Golang templates can reference each other using the `{{template "name"}}`
or `{{template "name" pipeline}}` actions. Even if they are defined in different files.

Golang's templates can be used in these ways:

- inline inside every HTML attribute and inside almost all Gotmx attributes
  like `<div class="{{if .Active }} green {{else}} red {{end}}"></div>`
- referenced inside every HTML attribute and almost all Gotmx attributes like `<div class="[[ :TemplateName ]]"></div>`
- inside the content of an HTML element, if it has the attribute `g-as-template` or `g-as-unsafe-template`

Inside a Golang template you are "leaving" Gotmx world, so you can not use any of the Gotmx attributes.
But you can still call Gotmx templates (or any other type of template) using the `GTemplate` and `GTextTemplate`
functions, as long as they are registered with Gotmx.
The difference between `GTemplate` and `GTextTemplate` is that `GTemplate` will perform escaping of special HTML chars
and `GTextTemplate` will not.

```html
<!-- The content of the following div is defined to be a Golang template using the data-g-as-template attribute -->
<div data-g-as-template="my-golang-template">
    {{if .LoggedIn }}
    You are logged in!
    {{else}}
    You are logged out!
    {{end}}
    <div data-g-inner-text="This will not work. Gotmx attributes don't work inside a Golang template"></div>
    <!-- You can call any other template with the GTemplate or GTextTemplate function. No matter what the type of the
    called template is. For example a Gotmx template -->
    <div>{{ GTextTemplate "my-gotmx-template" .}}</div>
</div>

<!-- And we define another template as a Gotmx template -->
<div data-g-define="my-gotmx-template">
    This can even be called from a Golang template using the GTemplate or GTextTemplate function.
</div>
```

When using a template in an attribute, either inline or using `[[ :TemplateName ]]`, the resulting string will always be
escaped, even if the referenced template was an unsafe Golang template.

## Golang templates under the hood

Under the hood Gotmx collects all Golang templates during parsing HTML and associates them with two root templates.
A root template from the package `html/template` and a root template from the package `text/template`.
Two root templates are needed because at render time, templates could be rendered as HTML or as text (escaped or unescaped).

Every "innerHTML" of an element that has the attribute `g-as-template` or `g-as-unsafe-template` is registered and
associated with the two root templates Gotmx manages.
Also, every attribute in the template HTML is checked, whether it contains the `{{` and `}}` delimiters.
If it does, the attribute value is considered to be a Golang template and is also registered and associated with the
two root templates.
If a template does not have an explicit name, which is the case for `g-as-template` or `g-as-unsafe-template` with an
empty value and all templates inside attributes, Gotmx will generate a unique name for it (an auto-incrementing ID).
Template literals in attributes are replaced by a call to the generated template name, for example:
```html
<div data-g-inner-text="{{if .MyBool}} red {{else}} green {{end}}"></div>
<!-- will be replaced by a simple Gotmx template call -->
<div data-g-inner-text="[[ :_tmpl_1 ]]"></div>
```

This way, all templates can reference each other and only have to be parsed once at startup.



