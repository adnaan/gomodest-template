import {Controller} from "@hotwired/stimulus"
import * as Turbo from "@hotwired/turbo";
import debounce from "lodash.debounce"

export default class extends Controller {
    static values = {
        url: String,
        changeRequestId: String,
        action: String,
        target: String,
        targets: String,
        template: String,
        params: Object,
        redirect: String,
        inputDebounce: {type: Number, default: 1000},
    }

    initialize() {
        let connectURL = `ws://${window.location.host}${window.location.pathname}`
        if (window.location.protocol === "https:") {
            connectURL = `wss://${window.location.host}${window.location.pathname}`
        }
        this.onSocketReconnect  = () => {
            if (this.dispatcher) {
                if (!this.changeRequestIdValue) {
                    console.error("action submit requires changeRequestId")
                    return
                }
                this.dispatcher(this.changeRequestIdValue, this.actionValue, this.targetValue, this.targetsValue, this.TemplateValue, this.paramsValue)
            }
        }
        this.input = debounce(this.input,this.inputDebounceValue).bind(this);
        this.dispatcher = changeRequestsDispatcher(connectURL, [], this.onSocketReconnect)
    }

    connect() {
        if (this.redirectValue){
            window.location.href = this.redirectValue
        }
    }


    submit(e) {
        e.preventDefault()
        const {changeRequestId, action, target, targets, template, ...rest} = e.params
        if (!changeRequestId) {
            console.error("action submit requires changeRequestId")
            return
        }
        let json = {...rest};
        let formData = new FormData(e.currentTarget);
        formData.forEach((value, key) => json[key] = value);
        if (this.dispatcher) {
            this.dispatcher(changeRequestId, action, target, targets, template, json)
        }
    }

    change(e) {
        e.preventDefault()
        const {changeRequestId, action, target, targets, template, ...rest} = e.params
        if (!changeRequestId) {
            console.error("action submit requires changeRequestId")
            return
        }
        let json = {...rest};
        if (this.dispatcher) {
            this.dispatcher(changeRequestId, action, target, targets, template, json)
        }
    }

    input(e) {
        const {changeRequestId, action, target, targets, template, ...rest} = e.params
        if (!changeRequestId) {
            console.error("action submit requires changeRequestId")
            return
        }
        let json = {...rest};
        json[e.target.name] = e.target.value
        if (this.dispatcher) {
            this.dispatcher(changeRequestId, action, target, targets, template, json)
        }
    }

    navigate(e) {
        const {route} = e.params;
        if (!route){
            return
        }
        window.location.href = route;
    }

}

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const changeRequestsDispatcher = (url, socketOptions, onSocketReconnect) => {
    let socket, openPromise, reopenTimeoutHandler;
    let reopenCount = 0;

    // socket code copied from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
    // thank you https://github.com/arlac77 !!
    function reopenTimeout() {
        const n = reopenCount;
        reopenCount++;
        return reopenTimeouts[
            n >= reopenTimeouts.length - 1 ? reopenTimeouts.length - 1 : n
            ];
    }

    function closeSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
        }

        if (socket) {
            socket.close();
            socket = undefined;
        }
    }

    function reOpenSocket() {
        closeSocket();
        reopenTimeoutHandler = setTimeout(() => {

                onSocketReconnect()
                openSocket().then(() => Turbo.session.connectStreamSource(socket)).catch(e => {

                })
            },
            reopenTimeout());
    }

    async function openSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
            reopenTimeoutHandler = undefined;
        }

        // we are still in the opening phase
        if (openPromise) {
            return openPromise;
        }


        socket = new WebSocket(url, socketOptions);

        socket.onclose = event => reOpenSocket();

        openPromise = new Promise((resolve, reject) => {
            socket.onerror = error => {
                reject(error);
                openPromise = undefined;
            };
            socket.onopen = event => {
                reopenCount = 0;
                resolve();
                openPromise = undefined;
            };
        });
        return openPromise;
    }

    openSocket().then(() => {
        Turbo.session.connectStreamSource(socket);
    });
    return (id, action, target, targets, template, params) => {
        if (!id) {
            throw 'changeRequest.id is required';
        }
        const changeRequest = {
            id: id,
            action: action,
            target: target,
            targets: targets,
            template: template,
            params: params
        }
        const send = () => socket.send(JSON.stringify(changeRequest));
        if (!socket || socket && socket.readyState !== WebSocket.OPEN) openSocket().then(send);
        else send();
    }
}