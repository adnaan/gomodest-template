# gomodest-template

A [modest](https://modestjs.works) template to build dynamic web apps in Go, HTML and [sprinkles and spots](https://modestjs.works/book/part-2/same-ui-three-modest-ways/) of javascript.

## Why ?

- Build dynamic websites using the tools you already know(Go, HTML, CSS, Vanilla Javascript) for the most part.
- Use [bulma components](https://bulma.io/documentation/components/) to speed up prototyping a responsive and good-looking UI.
- Use [turbo & stimulusjs](https://hotwire.dev/) for most of the interactivity.
- For really complex interactivity use [Svelte](https://svelte.dev/) for a single div in a few spots.
- Lightweight and productive. Fast development cycle.
- Easy to start, easy to maintain.

For a more complete implementation using this technique please see [gomodest-starter](https://github.com/adnaan/gomodest-starter).

## Usage
- [Use as a github template](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template#creating-a-repository-from-a-template)
- `git clone https://github.com/<user>/<mytemplate>` and `cd /path/to/your/gomodest-template`
- `make watch` (starts hot-reload for go, html and javascript changes)
- open [localhost:3000](http://localhost:3000).

or 

```bash
brew install gh
gh repo create myapp --template adnaan/gomodest-template
cd myapp
make install # or (make install-x64)
# replace gomodest-template with your app name
go get github.com/piranha/goreplace
$(go env GOPATH)/bin/goreplace gomodest-template -r myapp
git add . && git commit -m "replace gomodest-template"
make watch # or make watch-x64
```


![gomodest tempalte home](screenshots/gomodest-template-index.png?raw=true "")

## TOC

* [Folder Structure](#folder-structure)
* [Views using html templates](#views-using-html-templates)
  + [Step 1: Add a layout partial](#step-1--add-a-layout-partial)
  + [Step 2: Add a layout](#step-2--add-a-layout)
  + [Step 4: Add a view partial](#step-4--add-a-view-partial)
  + [Step 5: Add a view](#step-5--add-a-view)
  + [Step 6: Render view](#step-6--render-view)
* [Interactivity using Javascript](#interactivity-using-javascript)
  + [Stimulus Controllers](#stimulus-controllers)
    - [Step 1: Add a controller](#step-1--add-a-controller)
    - [Step 2: Add data attributes to the target div](#step-2--add-data-attributes-to-the-target-div)
  + [Svelte Components](#svelte-components)
    - [Step 1: Add data attributes to the target div.](#step-1--add-data-attributes-to-the-target-div)
    - [Step 2: Create and export svelte component](#step-2--create-and-export-svelte-component)
    - [Step 3: Hydrate initial props from the server](#step-3--hydrate-initial-props-from-the-server)
* [Styling and Images](#styling-and-images)
* [Samples](#samples)
* [Dependencies](#dependencies)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

## Folder Structure

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



## Views using html templates

There are three kinds of `html/template` files in this project.

- `layout`: defines the base structure of a web page.
- `partial`: reusable snippets of html. It can be of two types: `layout partials` & `view partials`.
- `view`: the main container for the web page logic contained within a `layout`. It must be enclosed in a `define content` template definition. It can use view partials.

### Step 1: Add a layout partial

Create `header.html` file in `templates/partials`.

```html
<meta charset="UTF-8">
<meta name="description" content="A modest way to build golang web apps">
<meta name="viewport"content="width=device-width, initial-scale=1.0, maximum-scale=5.0, minimum-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
...
```

### Step 2: Add a layout

Create `index.html` file in `templates/layouts` and use the above partial.

```html
<!DOCTYPE html>

<html lang="en">
<head>
    <title>{{.app_name}}</title>
    {{include "partials/header"}}
</head>
<body ...>
...
</body>
</html>

```

### Step 4: Add a view partial

Create `main.html` in `templates/partials`

```html
{{define "main"}}
    <main>
        <div class="columns is-centered is-vcentered is-mobile py-5">
            <div class="column is-narrow" style="width: 70%">
                <h1 class="has-text-centered title">Hello {{.hello}}</h1>
            </div>
        </div>
    </main>
{{end}}
```

This is a different from the `layout partial` since it's closed in a `define` tag.

### Step 5: Add a view

Create `home.html` in `templates` and use the above partial.

```html
{{define "content"}}
<div class="columns is-mobile is-centered">
    <div class="column is-half-desktop">
        {{template "main" .}}
    </div>
</div>
{{end}}
```
Notice that a `view` is always enclosed in `define content` template definition.

### Step 6: Render view

To render the view with data we use a wrapper over the `html/template` package.

```go
r.Get("/", indexLayout.Handle("home",
    func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
        return rl.M{
            "hello": "world",
    }, nil
}))
```

To learn more about `html/template`, please look into this amazing [cheatsheet](https://curtisvermeeren.github.io/2017/09/14/Golang-Templates-Cheatsheet).


Reference:

- `templates/layout/index.html`
- `templates/partials/header.html`
- `templates/partials/main.html`
- `templates/home.html`
- `main.go`

## Interactivity using Javascript

For client-side interactivity we use a bit of javascript.

### Stimulus Controllers

A stimulus controller is a snippet of javascript which handles a single aspect of interactivity. To add a new svelte component:

#### Step 1: Add a controller

Create a file with suffix: `_controller.js` 

`util_controller.js`
```js
import { Controller } from "stimulus"

export default class extends Controller {
  ...

    connect(){

    }

    goto(e){
        if (e.currentTarget.dataset.goto){
            window.location = e.currentTarget.dataset.goto;
        }
    }

    goback(e){
       window.history.back();
    }

   ...
}


```

See complete implementation in `assets/src/controller/util_controller.js`. To understand how stimulus works, please see the [handbook](https://stimulus.hotwire.dev/handbook/introduction).

#### Step 2: Add data attributes to the target div

```html
<body data-controller="util svelte"
      data-action="keydown@window->util#keyDown "
      data-util-active-class="is-active">
  
    ...
  <button class="button"
    data-action="click->util#goto"
    data-goto="/">Home
  </button>
</body>
```

Here we are attaching two controllers to the `body` itself since they are used often. Later we can add action and data attributes to use them.

Reference:
- `templates/layout/index.html`
- `templates/404.html`
- `assets/src/controllers/util_controller.js`

### Svelte Components

A svelte component is loaded into the targeted div by a stimulujs controller: `controllers/svelte_controller.js`. This is hooked by declaring data attributes on the div which is to be contain the svelte component:

- `data-svelte-target`: Value is **required** to be `component`. It's used for identifying the divs as targets for the `svelte_controller`.
- `data-component-name`: The name of the component as exported in `src/components/index.js`

  ```js
    import app from "./App.svelte"
    // export other components here.
    export default {
        app: app,
    }
  ```
  
- `data-component-props`: A string map object which is passed as initial props to the svelte component.

To add a new svelte component:

#### Step 1: Add data attributes to the target div.
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

#### Step 2: Create and export svelte component

- Create a new svelte component in `src/components` and export it in `src/components/index.js`

```js
import app from "./App.svelte"

// export other components here.
export default {
    app: app,
}
```

The `controllers/svelte_controller.js` controller loads the svelte component in to the div with the required data attributes shown in step 1.


#### Step 3: Hydrate initial props from the server

It's possible to hydrate initial props from the server and pass onto the component. This is done by templating a string data object into the `data-component-props` attribute.

```go
r.Get("/app", indexLayout.Handle("app",
    func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
      appData := struct {
      Title string `json:"title"`
      }{
      Title: "Hello from server for the svelte component",
      }
    
      d, err := json.Marshal(&appData)
      if err != nil {
      return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
      }
    
      return rl.M{
      "Data": string(d), // notice struct data is converted into a string
      }, nil
}))
```

Reference:
- `templates/app.html`
- `src/controllers/svelte_controller.js`
- `src/components/*`
- `main.go`

## Styling and Images

[Bulma](https://bulma.io/) is included by default. Bulma is a productive css framework with prebuilt components and helper utilities.

- `assets/src/styles.scss`: to override default bulma variables. `webpack` bundles and copies css assets to `public/assets/css.
- `assets/images`: put image assets here. it will be auto-copied to `public/assets/images` by `webpack`. 

## Samples

Go to [localhost:3000/samples](http://localhost:3000/samples/) to a list of sample views. Copy-paste at will from the `templates/samples` directory.

## Dependencies

- Backend:
  - [Go](https://golang.org/)
  - [renderlayout: a wrapper over foolin/goview](https://github.com/adnaan/renderlayout)
  - [chi](https://github.com/go-chi/chi)
- Frontend:
  - [html/template](https://golang.org/pkg/html/template/)
  - [masterminds/sprig](http://masterminds.github.io/sprig/)
  - [StimulusJS(sprinkles)](https://stimulus.hotwire.dev/)
  - [SvelteJS(spots)](https://svelte.dev/)
  - [Bulma CSS](https://bulma.io/)
  - [Webpack](https://webpack.js.org/)
