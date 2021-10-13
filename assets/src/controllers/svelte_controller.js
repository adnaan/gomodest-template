import { Controller } from "@hotwired/stimulus";
import components from "../components";

export default class extends Controller {
    static targets = ["component"]
    connect() {
        if (this.componentTargets.length > 0){
            this.componentTargets.forEach(el => {
                const componentName = el.dataset.componentName;
                const componentProps = el.dataset.componentProps ? JSON.parse(el.dataset.componentProps): {};
                if (!(componentName in components)){
                    console.error(`svelte component: ${componentName}, not found!`)
                    return;
                }
                // console.log(componentProps)
                const app = new components[componentName]({
                    target: el,
                    props: componentProps
                });
            })
        }
    }
}