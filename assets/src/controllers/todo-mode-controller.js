import {Controller} from "@hotwired/stimulus";

export default class extends Controller {
    static targets = ["view","edit","delete"]
    static values = {
        hiddenClass: { type: String, default: 'is-hidden' },
        mode: {type: String, default: 'view'},
    }

    view(){
        this.modeValue = "view"
    }

    edit(){
        this.modeValue = "edit"
    }

    delete(){
       this.modeValue = "delete"
    }

    modeValueChanged(){
        this.showCurrentMode()
    }

    showCurrentMode(){
        switch (this.modeValue) {
            case "view":
                this.viewTarget.classList.remove(this.hiddenClassValue);
                this.editTarget.classList.add(this.hiddenClassValue);
                this.deleteTarget.classList.add(this.hiddenClassValue);
                break;
            case "edit":
                this.viewTarget.classList.add(this.hiddenClassValue);
                this.editTarget.classList.remove(this.hiddenClassValue);
                this.deleteTarget.classList.add(this.hiddenClassValue);
                break;
            case "delete":
                this.viewTarget.classList.add(this.hiddenClassValue);
                this.editTarget.classList.add(this.hiddenClassValue);
                this.deleteTarget.classList.remove(this.hiddenClassValue);
                break;
        }
    }
}