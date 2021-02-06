import { Controller } from "stimulus"

export default class extends Controller {
    static targets = ["modal","toggler","dropup"]
    static classes = [ "active" ]

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

    openModal(e){
       const targetModal = this.modalTargets.find(i => i.id === e.currentTarget.dataset.modalTargetId);
       targetModal.classList.add("is-active")
        e.preventDefault();
    }

    closeModal(e){
        if (e.type === "click"){
            const targetModal = this.modalTargets.find(i => i.id === e.currentTarget.dataset.modalTargetId);
            targetModal.classList.remove("is-active")
            e.preventDefault();
            return;
        }
    }

    keyDown(e){
        if (e.keyCode === 27){
            this.modalTargets.forEach(item => {
                item.classList.remove("is-active")
            })
        }

        if (e.keyCode === 37){
            window.history.back();
        }


        if (e.keyCode === 39){
            window.history.forward();
        }
    }

    toggle(e){
        if (!e.currentTarget.dataset.toggleIds){
            return;
        }
        const targetToggleIds = e.currentTarget.dataset.toggleIds.split(",");
        const targetToggleClass =   e.currentTarget.dataset.toggleClass;
        targetToggleIds.forEach(item => {
            document.getElementById(item).classList.toggle(targetToggleClass);
        })
    }

    toggleIsActive(e){
       this.dropupTarget.classList.toggle(this.activeClass)
    }
}

