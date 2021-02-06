# gomodest-template

A [modest](https://modestjs.works) template to build dynamic websites in Go, HTML and [sprinkles and spots](https://modestjs.works/book/part-2/same-ui-three-modest-ways/) of javascript.

See Example SAAS starter kit with authentication, billing based on the template: [gomodest](https://github.com/adnaan/gomodest)

## Usage
- [Use as a template](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template#creating-a-repository-from-a-template)
- cd /path/to/your/gomodest-template
- make watch (starts hot-reload for go, html and javascript changes)
## Dependencies

- Backend: 
  - [Go](https://golang.org/) 
  - [renderlayout: a wrapper over foolin/goview](https://github.com/adnaan/renderlayout)
  - [chi](https://github.com/go-chi/chi)
- Frontend:
  - [html/template](https://golang.org/pkg/html/template/)
  - [StimulusJS(sprinkles)](https://stimulus.hotwire.dev/)
  - [SvelteJS(spots)](https://svelte.dev/)
  - [Bulma CSS](https://bulma.io/)
  - [Webpack](https://webpack.js.org/)

## Project Structure

- templates/
    
    - layouts/
    - partials/
    - list of view files
  
- assets/
  
  - images/
  - src/
  
    - components/
    - controllers/
    - index.js
    - styles.scss



- `templates` is the root directory where all html/templates assets are found.
- `layouts` contains the layout files. Layout is a container for `partials` and `view files`
- `partials` contains the partial files. Partial is a reusable html template which can be used in one of two ways:

    - Included in a `layout` file: `{{include "partials/header"}}`
    - Included in a `view` file: `{{template "main" .}}`. When used in a view file, a partial must be enclosed in a `define` tag:
      
        ```html
            {{define "main"}}
              Hello {{.hello}}
            {{end}}
        ```
- `view` files are put in the root of the `templates` directory. They are contained within a `layout` must be enclosed in a `define content` tag:

    ```html
        {{define "content"}}
            App's {{.dashboard}}
        {{end}}
    ```
    `View` is rendered within a `layout`: 
    
    ```go
        indexLayout, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.DefaultHandler(func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
			return rl.M{
				"app_name": "gomdest-template",
			}, nil
		}))
       
        ...
        
        r.Get("/", indexLayout.Handle("home",
		func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
			return rl.M{
				"hello": "world",
			}, nil
		}))
    ```
  
    Here the `view`: `home` is rendered within the `index` layout.

Please see the `templates` directory.

- `assets` directory contains the public asset pipeline for the project.

  - `styles.scss` is a custom `scss` file for [bulma][https://bulma.io] as [documented here](https://bulma.io/documentation/customize/with-webpack/).
  - `index.js`  is the entrypoint for loading `stimulusjs` controllers sourced from this [example](https://github.com/hotwired/stimulus-starter.
  - `controllers` contains [stimulusjs controllers](https://stimulus.hotwire.dev/reference/controllers).
  - `components` contains single file svelte components.
  

## Using Svelte Components

A svelte component is loaded into the targeted div by a stimulujs controller: `controllers/svelte_controller.js`

### Step 1: Add data attributes to the target div.
```html
{{define "content"}}
<div class="columns is-mobile is-centered">
 <div class="column is-half-desktop">
  <div
          data-svelte-target="component"
          data-component-name="app"
          data-component-props="{{.Data}}">
  </div>
 </div>
</div>
</div>
{{end}}
```

### Step 2: Create and export svelte component

- Create a new svelte component in `src/components` and export it in `src/components/index.js`

```js
import app from "./App.svelte"

// export other components here.
export default {
    app: app,
}
```

The `controllers/svelte_controller.js` controller loads the svelte component in to the div with the required data attributes shown in step 1.


